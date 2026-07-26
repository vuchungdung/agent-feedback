package tools

import (
	"agent_feedback/internal/models"
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/tmc/langchaingo/llms"
)

type CustomerTool struct {
	dataFile  string
	customers map[string]models.Customer
}

func NewCustomerTool(path string) (*CustomerTool, error) {

	bytes, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var customers []models.Customer

	if err := json.Unmarshal(bytes, &customers); err != nil {
		return nil, err
	}

	customerMap := make(map[string]models.Customer)
	for _, p := range customers {
		customerMap[p.ID] = p
	}

	return &CustomerTool{
		dataFile:  path,
		customers: customerMap,
	}, nil
}

func (t *CustomerTool) Name() string {
	return "get_customer"
}

func (t *CustomerTool) Description() string {
	return `
Lookup customer information by customer ID.

Input:
{
    "customer_id":"CUS001"
}

Returns:
- customer tier
- tenure
- previous support tickets
- total orders
- customer tags

Use this tool whenever customer information is required before generating a support report.
`
}

func (t *CustomerTool) Execute(
	ctx context.Context,
	customerID string,
) (*models.Customer, error) {
	customer, ok := t.customers[customerID]
	if !ok {
		return nil, fmt.Errorf("customer not found")
	}
	return &customer, nil
}

func (t *CustomerTool) Definition() llms.Tool {
	return llms.Tool{
		Type: "function",
		Function: &llms.FunctionDefinition{
			Name:        t.Name(),
			Description: t.Description(),
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"customer_id": map[string]any{
						"type":        "string",
						"description": "Unique customer ID, for example CUS001",
					},
				},
				"required":             []string{"customer_id"},
				"additionalProperties": false,
			},
		},
	}
}
