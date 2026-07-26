package agent

import (
	"agent_feedback/internal/models"
	"agent_feedback/internal/prompts"
	"context"
	"encoding/json"

	"github.com/tmc/langchaingo/llms"
)

type FeedbackClassifier interface {
	Classify(ctx context.Context, feedback string) (*models.Classification, error)
}

type OpenAIClassifier struct {
	llm llms.Model
}

func NewOpenAIClassifier(llm llms.Model) *OpenAIClassifier {
	return &OpenAIClassifier{
		llm: llm,
	}
}

func (c *OpenAIClassifier) Classify(ctx context.Context, feedback string) (*models.Classification, error) {
	messages := []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeSystem, prompts.ClassifierPrompt),
		llms.TextParts(llms.ChatMessageTypeHuman, feedback),
	}
	resp, err := c.llm.GenerateContent(ctx, messages)
	if err != nil {
		return nil, err
	}
	var result models.Classification
	err = json.Unmarshal([]byte(resp.Choices[0].Content), &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}
