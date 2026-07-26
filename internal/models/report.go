package models

type FeedbackReport struct {
	Summary string `json:"summary"`

	Category string `json:"category"`
	Urgency  string `json:"urgency"`

	CustomerContext CustomerContext `json:"customer_context"`

	References []ReportReference `json:"references"`

	SuggestedActions []string `json:"suggested_actions"`

	Confidence float64 `json:"confidence"`

	HumanReviewRequired bool   `json:"human_review_required"`
	HumanReviewReason   string `json:"human_review_reason,omitempty"`
}

type CustomerContext struct {
	Available bool   `json:"available"`
	Summary   string `json:"summary"`
}

type ReportReference struct {
	Type  string `json:"type"`
	ID    string `json:"id"`
	Title string `json:"title"`
}
