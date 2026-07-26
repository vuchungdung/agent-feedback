package models

type AnalyzeFeedbackResponse struct {
	FeedbackID string          `json:"feedback_id"`
	Report     *FeedbackReport `json:"report"`
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}
