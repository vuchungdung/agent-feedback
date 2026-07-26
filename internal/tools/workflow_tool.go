package tools

import (
	"agent_feedback/internal/models"
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/tmc/langchaingo/llms"
)

type WorkflowTool struct {
	workflows map[string]models.Workflow
}

func NewWorkflowTool(path string) (*WorkflowTool, error) {

	bytes, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var workflows []models.Workflow

	if err := json.Unmarshal(bytes, &workflows); err != nil {
		return nil, err
	}

	data := make(map[string]models.Workflow)

	for _, workflow := range workflows {
		data[workflow.Category] = workflow
	}

	return &WorkflowTool{
		workflows: data,
	}, nil
}

func (t *WorkflowTool) Name() string {
	return "get_workflow"
}

func (t *WorkflowTool) Description() string {
	return `
Lookup the standard customer support workflow for a feedback category.

Input:

{
    "category":"billing"
}

Returns:

- workflow name

- workflow description

- ordered workflow steps

- escalation target

Use this tool whenever you need to determine the operational process that Customer Support should follow before suggesting actions.
`
}

func (t *WorkflowTool) Execute(ctx context.Context, category string) (*models.Workflow, error) {
	workflow, ok := t.workflows[category]
	if !ok {
		return nil, fmt.Errorf("workflow not found for category %s", category)
	}
	return &workflow, nil
}

func (t *WorkflowTool) Definition() llms.Tool {
	return llms.Tool{
		Type: "function",
		Function: &llms.FunctionDefinition{
			Name:        t.Name(),
			Description: t.Description(),
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"category": map[string]any{
						"type": "string",
						"enum": []string{
							"bug_report",
							"billing",
							"feature_request",
							"praise",
							"churn_risk",
							"policy_violation",
							"unknown",
						},
					},
				},
				"required":             []string{"category"},
				"additionalProperties": false,
			},
		},
	}
}
