package models

import (
	"time"
)

type FeedbackInput struct {
	ID         string
	Message    string    `json:"message" validate:"required"`
	CustomerID string    `json:"customer_id,omitempty"`
	Email      string    `json:"email,omitempty"`
	Channel    string    `json:"channel" validate:"required,oneof=email web app social"`
	Timestamp  time.Time `json:"timestamp"`
}
