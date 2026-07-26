package handler

import (
	"agent_feedback/helper"
	"agent_feedback/internal/models"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
)

type FeedbackProcessor interface {
	Process(
		ctx context.Context,
		input models.FeedbackInput,
	) (*models.FeedbackReport, error)
}

type FeedbackHandler struct {
	agent FeedbackProcessor
}

func NewFeedbackHandler(
	agent FeedbackProcessor,
) *FeedbackHandler {
	return &FeedbackHandler{
		agent: agent,
	}
}

func (h *FeedbackHandler) Analyze(
	w http.ResponseWriter,
	r *http.Request,
) {
	if r.Method != http.MethodPost {
		helper.WriteJSON(
			w,
			http.StatusMethodNotAllowed,
			models.ErrorResponse{
				Error:   "method_not_allowed",
				Message: "only POST is supported",
			},
		)
		return
	}

	defer r.Body.Close()

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	var request models.FeedbackInput

	if err := decoder.Decode(&request); err != nil {
		helper.WriteJSON(
			w,
			http.StatusBadRequest,
			models.ErrorResponse{
				Error:   "invalid_request",
				Message: fmt.Sprintf("invalid JSON body: %v", err),
			},
		)
		return
	}

	input := models.FeedbackInput{
		ID: fmt.Sprintf(
			"FB-%d",
			time.Now().UnixMilli(),
		),
		CustomerID: strings.TrimSpace(request.CustomerID),
		Email:      strings.TrimSpace(request.Email),
		Message:    strings.TrimSpace(request.Message),
		Channel:    request.Channel,
		Timestamp:  time.Now().UTC(),
	}

	log.Printf(
		"[HTTP] analyze feedback id=%s customer_id=%s",
		input.ID,
		input.CustomerID,
	)

	report, err := h.agent.Process(
		r.Context(),
		input,
	)
	if err != nil {
		log.Printf(
			"[HTTP] agent processing failed id=%s error=%v",
			input.ID,
			err,
		)

		helper.WriteJSON(
			w,
			http.StatusInternalServerError,
			models.ErrorResponse{
				Error:   "agent_processing_failed",
				Message: "failed to analyze customer feedback",
			},
		)
		return
	}

	helper.WriteJSON(
		w,
		http.StatusOK,
		models.AnalyzeFeedbackResponse{
			FeedbackID: input.ID,
			Report:     report,
		},
	)
}
