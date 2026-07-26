package models

type ReviewReason string

const (
	ReviewReasonLowConfidence      ReviewReason = "low_confidence"
	ReviewReasonAmbiguousCategory  ReviewReason = "ambiguous_category"
	ReviewReasonMissingInformation ReviewReason = "missing_information"
	ReviewReasonMissingPolicy      ReviewReason = "missing_policy"
	ReviewReasonMissingWorkflow    ReviewReason = "missing_workflow"
	ReviewReasonRestrictedAction   ReviewReason = "restricted_action"
	ReviewReasonCustomerNotFound   ReviewReason = "customer_not_found"
)

type FeedbackReport struct {
	Summary  string `json:"summary"`
	Category string `json:"category"`
	Urgency  string `json:"urgency"`

	CustomerContext CustomerContext `json:"customer_context"`

	CustomerClaims []string `json:"customer_claims"`
	VerifiedFacts  []string `json:"verified_facts"`
	Unknowns       []string `json:"unknowns"`

	References []ReportReference `json:"references"`

	RecommendedNextSteps []string            `json:"recommended_next_steps"`
	ConditionalActions   []ConditionalAction `json:"conditional_actions"`

	Confidence float64 `json:"confidence"`

	HumanReviewRequired bool           `json:"human_review_required"`
	HumanReviewReasons  []ReviewReason `json:"human_review_reasons"`
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

type ConditionalAction struct {
	Condition             string `json:"condition"`
	Action                string `json:"action"`
	RequiresHumanApproval bool   `json:"requires_human_approval"`
}
