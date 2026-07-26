package models

type Category string

const (
	CategoryBug             Category = "bug_report"
	CategoryBilling         Category = "billing"
	CategoryFeatureRequest  Category = "feature_request"
	CategoryPraise          Category = "praise"
	CategoryChurnRisk       Category = "churn_risk"
	CategoryPolicyViolation Category = "policy_violation"
	CategoryUnknown         Category = "unknown"
)

type Urgency string

const (
	Low    Urgency = "low"
	Medium Urgency = "medium"
	High   Urgency = "high"
)

type Classification struct {
	Category   Category
	Urgency    Urgency
	Confidence float64
	Reason     string
}
