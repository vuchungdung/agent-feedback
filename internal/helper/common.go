package helper

import (
	"agent_feedback/internal/models"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
)

func MarshalToolResult(value any) (string, error) {
	result, err := json.Marshal(value)
	if err != nil {
		return "", fmt.Errorf(
			"marshal tool result: %w",
			err,
		)
	}

	return string(result), nil
}

func BuildToolErrorResult(toolName string, toolErr error) string {
	payload := map[string]any{
		"success":      false,
		"tool":         toolName,
		"error":        toolErr.Error(),
		"human_review": true,
		"instruction": `
Continue generating the report, state that this data source
was unavailable, and do not invent the missing information.
`,
	}

	result, err := json.Marshal(payload)
	if err != nil {
		return fmt.Sprintf(
			`{"success":false,"tool":%q,"error":%q}`,
			toolName,
			toolErr.Error(),
		)
	}

	return string(result)
}

func ValidateFeedbackInput(input models.FeedbackInput) error {
	if strings.TrimSpace(input.Message) == "" {
		return errors.New("feedback text is required")
	}

	if strings.TrimSpace(input.CustomerID) == "" &&
		strings.TrimSpace(input.Email) == "" {
		return errors.New(
			"either customer_id or email is required",
		)
	}

	return nil
}

func WriteJSON(
	w http.ResponseWriter,
	status int,
	value any,
) {
	w.Header().Set(
		"Content-Type",
		"application/json",
	)

	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(value); err != nil {
		log.Printf(
			"[HTTP] failed to encode response: %v",
			err,
		)
	}
}
