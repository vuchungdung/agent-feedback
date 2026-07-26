package tools

import (
	"agent_feedback/internal/models"
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/tmc/langchaingo/llms"
)

type PolicyTool struct {
	dataFile string
	policies map[string]models.Policy
}

func NewPolicyTool(path string) (*PolicyTool, error) {
	bytes, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var policies []models.Policy

	if err := json.Unmarshal(bytes, &policies); err != nil {
		return nil, err
	}

	policyMap := make(map[string]models.Policy)

	for _, p := range policies {
		policyMap[p.Category] = p
	}

	return &PolicyTool{
		dataFile: path,
		policies: policyMap,
	}, nil
}

func (t *PolicyTool) Name() string {
	return "get_policy"
}

func (t *PolicyTool) Description() string {
	return `
Lookup company policy based on a feedback category.

Input:

{
    "category":"billing"
}

Returns:

- policy title

- policy description

- business rules

- reference id

Use this tool before recommending any customer action.
`
}

func (t *PolicyTool) Execute(ctx context.Context, category string) (*models.Policy, error) {
	policy, ok := t.policies[category]
	if !ok {
		return nil, fmt.Errorf("policy not found")
	}
	return &policy, nil
}

func (t *PolicyTool) Definition() llms.Tool {
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
