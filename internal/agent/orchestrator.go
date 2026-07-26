package agent

import (
	"agent_feedback/helper"
	"agent_feedback/internal/models"
	"agent_feedback/internal/prompts"
	"agent_feedback/internal/tools"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/tmc/langchaingo/llms"
)

const (
	maxToolIterations = 6
)

type Agent struct {
	llm          llms.Model
	reporter     Reporter
	classifier   FeedbackClassifier
	customerTool tools.CustomerTool
	policyTool   tools.PolicyTool
	workflowTool tools.WorkflowTool
}

func NewAgent(
	llm llms.Model,
	reporter Reporter,
	classifier FeedbackClassifier,
	customer tools.CustomerTool,
	policy tools.PolicyTool,
	workflow tools.WorkflowTool,
) *Agent {
	return &Agent{
		llm:          llm,
		reporter:     reporter,
		classifier:   classifier,
		customerTool: customer,
		policyTool:   policy,
		workflowTool: workflow,
	}
}

func (a *Agent) Process(ctx context.Context, input models.FeedbackInput) (*models.FeedbackReport, error) {
	if err := helper.ValidateFeedbackInput(input); err != nil {
		return nil, err
	}

	log.Printf(
		"[Agent] received feedback customer_id=%s channel=%s",
		input.CustomerID,
		input.Channel,
	)

	classification, err := a.classifier.Classify(ctx, input.Message)
	if err != nil {
		return nil, fmt.Errorf("classify feedback: %w", err)
	}

	log.Printf(
		"[Agent] classification category=%s urgency=%s confidence=%.2f",
		classification.Category,
		classification.Urgency,
		classification.Confidence,
	)

	// Step 2: initialize conversation history.
	messages := []llms.MessageContent{
		llms.TextParts(
			llms.ChatMessageTypeSystem,
			prompts.AgentSystemPrompt,
		),
		llms.TextParts(
			llms.ChatMessageTypeHuman,
			prompts.BuildHumanPrompt(input, *classification),
		),
	}

	// Step 3: define the tools visible to the LLM.
	availableTools := a.toolDefinitions()

	// Step 4: agentic tool-calling loop.
	for iteration := 0; iteration < maxToolIterations; iteration++ {
		log.Printf("[Agent] LLM iteration=%d", iteration+1)

		response, err := a.llm.GenerateContent(
			ctx,
			messages,
			llms.WithTools(availableTools),
		)
		if err != nil {
			return nil, fmt.Errorf(
				"generate agent response at iteration %d: %w",
				iteration+1,
				err,
			)
		}

		if len(response.Choices) == 0 {
			return nil, errors.New("LLM returned no choices")
		}

		choice := response.Choices[0]

		// No tool call means the model has produced its final report.
		if len(choice.ToolCalls) == 0 {
			finalContent := strings.TrimSpace(
				choice.Content,
			)

			if finalContent == "" {
				return nil, errors.New(
					"LLM returned neither tool calls nor final content",
				)
			}

			report, err := a.reporter.Parse(
				finalContent,
				*classification,
			)
			if err != nil {
				return nil, fmt.Errorf(
					"parse final feedback report: %w",
					err,
				)
			}

			log.Printf(
				"[Agent] report generated category=%s urgency=%s confidence=%.2f human_review=%t",
				report.Category,
				report.Urgency,
				report.Confidence,
				report.HumanReviewRequired,
			)

			return report, nil
		}

		log.Printf(
			"[Agent] model requested %d tool call(s)",
			len(choice.ToolCalls),
		)

		/*
			Append the assistant message containing all tool calls.

			This is important because the next request must contain:

			assistant -> tool_calls
			tool      -> tool result
		*/
		assistantParts := make([]llms.ContentPart, 0, len(choice.ToolCalls)+1)

		// Preserve assistant text when the model returns both content and calls.
		if strings.TrimSpace(choice.Content) != "" {
			assistantParts = append(
				assistantParts,
				llms.TextContent{
					Text: choice.Content,
				},
			)
		}

		for _, toolCall := range choice.ToolCalls {
			assistantParts = append(assistantParts, toolCall)
		}

		messages = append(
			messages,
			llms.MessageContent{
				Role:  llms.ChatMessageTypeAI,
				Parts: assistantParts,
			},
		)

		// Execute every tool requested in this response.
		for _, toolCall := range choice.ToolCalls {
			log.Printf(
				"[Agent] tool_call id=%s name=%s arguments=%s",
				toolCall.ID,
				toolCall.FunctionCall.Name,
				toolCall.FunctionCall.Arguments,
			)

			toolResult, toolErr := a.executeTool(ctx, toolCall)

			/*
				Do not immediately terminate the entire agent when a lookup
				fails.

				Feed the failure back to the LLM so it can produce a report
				with "data unavailable" and request human review.
			*/
			if toolErr != nil {
				log.Printf(
					"[Agent] tool failed name=%s error=%v",
					toolCall.FunctionCall.Name,
					toolErr,
				)

				toolResult = helper.BuildToolErrorResult(
					toolCall.FunctionCall.Name,
					toolErr,
				)
			} else {
				log.Printf(
					"[Agent] tool completed name=%s",
					toolCall.FunctionCall.Name,
				)
			}

			/*
				The ToolCallID must match the ID of the model's tool call.
				Otherwise OpenAI cannot associate the result with the call.
			*/
			messages = append(
				messages,
				llms.MessageContent{
					Role: llms.ChatMessageTypeTool,
					Parts: []llms.ContentPart{
						llms.ToolCallResponse{
							ToolCallID: toolCall.ID,
							Name:       toolCall.FunctionCall.Name,
							Content:    toolResult,
						},
					},
				},
			)
		}
	}

	return nil, fmt.Errorf(
		"agent exceeded maximum tool iterations: %d",
		maxToolIterations,
	)
}

func (a *Agent) toolDefinitions() []llms.Tool {
	return []llms.Tool{
		a.customerTool.Definition(),
		a.policyTool.Definition(),
		a.workflowTool.Definition(),
	}
}

func (a *Agent) executeTool(ctx context.Context, call llms.ToolCall) (string, error) {
	if call.FunctionCall == nil {
		return "", errors.New("tool call has no function definition")
	}

	switch call.FunctionCall.Name {
	case a.customerTool.Name():
		return a.executeCustomerTool(
			ctx,
			call.FunctionCall.Arguments,
		)

	case a.policyTool.Name():
		return a.executePolicyTool(
			ctx,
			call.FunctionCall.Arguments,
		)

	case a.workflowTool.Name():
		return a.executeWorkflowTool(
			ctx,
			call.FunctionCall.Arguments,
		)

	default:
		return "", fmt.Errorf(
			"unsupported tool: %s",
			call.FunctionCall.Name,
		)
	}
}

type getCustomerArguments struct {
	CustomerID string `json:"customer_id"`
}

func (a *Agent) executeCustomerTool(ctx context.Context, rawArguments string) (string, error) {
	var args getCustomerArguments

	if err := json.Unmarshal(
		[]byte(rawArguments),
		&args,
	); err != nil {
		return "", fmt.Errorf(
			"decode get_customer arguments: %w",
			err,
		)
	}

	args.CustomerID = strings.TrimSpace(args.CustomerID)
	if args.CustomerID == "" {
		return "", errors.New(
			"get_customer requires customer_id",
		)
	}

	customer, err := a.customerTool.Execute(
		ctx,
		args.CustomerID,
	)
	if err != nil {
		return "", fmt.Errorf(
			"get customer %q: %w",
			args.CustomerID,
			err,
		)
	}

	return helper.MarshalToolResult(customer)
}

type getPolicyArguments struct {
	Category string `json:"category"`
}

func (a *Agent) executePolicyTool(ctx context.Context, rawArguments string) (string, error) {
	var args getPolicyArguments

	if err := json.Unmarshal([]byte(rawArguments), &args); err != nil {
		return "", fmt.Errorf(
			"decode get_policy arguments: %w",
			err,
		)
	}

	args.Category = strings.TrimSpace(args.Category)
	if args.Category == "" {
		return "", errors.New(
			"get_policy requires category",
		)
	}

	policy, err := a.policyTool.Execute(
		ctx,
		args.Category,
	)
	if err != nil {
		return "", fmt.Errorf(
			"get policy for category %q: %w",
			args.Category,
			err,
		)
	}

	return helper.MarshalToolResult(policy)
}

type getWorkflowArguments struct {
	Category string `json:"category"`
}

func (a *Agent) executeWorkflowTool(ctx context.Context, rawArguments string) (string, error) {
	var args getWorkflowArguments

	if err := json.Unmarshal(
		[]byte(rawArguments),
		&args,
	); err != nil {
		return "", fmt.Errorf(
			"decode get_workflow arguments: %w",
			err,
		)
	}

	args.Category = strings.TrimSpace(args.Category)
	if args.Category == "" {
		return "", errors.New(
			"get_workflow requires category",
		)
	}

	workflow, err := a.workflowTool.Execute(
		ctx,
		args.Category,
	)
	if err != nil {
		return "", fmt.Errorf(
			"get workflow for category %q: %w",
			args.Category,
			err,
		)
	}

	return helper.MarshalToolResult(workflow)
}
