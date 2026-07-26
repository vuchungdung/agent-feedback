package prompts

import (
	"agent_feedback/internal/models"
	"fmt"
)

func BuildHumanPrompt(
	input models.FeedbackInput,
	classification models.Classification,
) string {

	return fmt.Sprintf(`
Customer ID:
%s

Email:
%s

Channel:
%s

Timestamp:
%s

Feedback:

%s

Initial Classification

Category:
%s

Urgency:
%s

Confidence:
%.2f

Reason:
%s

Please gather any additional information using available tools before generating the final report.
`,
		input.CustomerID,
		input.Email,
		input.Channel,
		input.Timestamp,
		input.Message,

		classification.Category,
		classification.Urgency,
		classification.Confidence,
		classification.Reason,
	)
}
