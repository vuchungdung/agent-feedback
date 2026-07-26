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
Analyze the following customer submission.

Feedback ID: %s
Customer ID: %s
Channel: %s

Customer feedback:
<customer_feedback>
%s
</customer_feedback>

Initial classification:
- category: %s
- urgency: %s
- confidence: %.2f
- reason: %s

Apply the system rules and return the required JSON report.
`,
		input.ID,
		input.CustomerID,
		input.Channel,
		input.Message,
		classification.Category,
		classification.Urgency,
		classification.Confidence,
		classification.Reason,
	)
}
