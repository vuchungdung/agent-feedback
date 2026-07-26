package agent

import (
	"agent_feedback/internal/models"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
)

type Reporter interface {
	Parse(rawContent string, classification models.Classification) (*models.FeedbackReport, error)
}

type JSONReporter struct {
	humanReviewThreshold float64
}

func NewJSONReporter(humanReviewThreshold float64) *JSONReporter {
	if humanReviewThreshold <= 0 ||
		humanReviewThreshold > 1 {
		humanReviewThreshold = 0.6
	}

	return &JSONReporter{
		humanReviewThreshold: humanReviewThreshold,
	}
}

func (r *JSONReporter) Parse(rawContent string, classification models.Classification) (*models.FeedbackReport, error) {
	content, err := extractJSONObject(rawContent)
	if err != nil {
		return nil, fmt.Errorf(
			"extract report JSON: %w",
			err,
		)
	}

	var report models.FeedbackReport

	if err := unmarshalStrictJSON(content, &report); err != nil {
		return nil, fmt.Errorf(
			"decode feedback report: %w",
			err,
		)
	}

	normalizeReport(&report)

	if err := validateReport(&report); err != nil {
		return nil, fmt.Errorf(
			"validate feedback report: %w",
			err,
		)
	}

	r.applyHumanReviewRules(&report, classification)

	return &report, nil
}

func unmarshalStrictJSON(content []byte, target any) error {
	decoder := json.NewDecoder(
		bytes.NewReader(content),
	)

	decoder.DisallowUnknownFields()

	if err := decoder.Decode(target); err != nil {
		return err
	}

	var extra any

	err := decoder.Decode(&extra)

	if !errors.Is(err, io.EOF) {
		if err == nil {
			return errors.New(
				"unexpected content after report JSON",
			)
		}

		return fmt.Errorf(
			"decode trailing content: %w",
			err,
		)
	}

	return nil
}

func extractJSONObject(rawContent string) ([]byte, error) {
	content := strings.TrimSpace(rawContent)

	if content == "" {
		return nil, errors.New(
			"empty report content",
		)
	}

	content = removeMarkdownFence(content)

	start := strings.Index(content, "{")
	if start == -1 {
		return nil, errors.New(
			"report does not contain JSON object",
		)
	}

	end := strings.LastIndex(content, "}")
	if end == -1 || end < start {
		return nil, errors.New(
			"report contains incomplete JSON object",
		)
	}

	jsonContent := strings.TrimSpace(
		content[start : end+1],
	)

	if !json.Valid([]byte(jsonContent)) {
		return nil, errors.New(
			"report contains invalid JSON",
		)
	}

	return []byte(jsonContent), nil
}

func removeMarkdownFence(content string) string {
	content = strings.TrimSpace(content)

	if !strings.HasPrefix(content, "```") {
		return content
	}

	lines := strings.Split(content, "\n")

	if len(lines) < 3 {
		return content
	}

	lines = lines[1:]

	if len(lines) > 0 &&
		strings.TrimSpace(lines[len(lines)-1]) == "```" {
		lines = lines[:len(lines)-1]
	}

	return strings.TrimSpace(
		strings.Join(lines, "\n"),
	)
}

func normalizeReport(report *models.FeedbackReport) {
	report.Summary = strings.TrimSpace(
		report.Summary,
	)

	report.Category = strings.ToLower(
		strings.TrimSpace(report.Category),
	)

	report.Urgency = strings.ToLower(
		strings.TrimSpace(report.Urgency),
	)

	report.CustomerContext.Summary =
		strings.TrimSpace(
			report.CustomerContext.Summary,
		)

	report.HumanReviewReason =
		strings.TrimSpace(
			report.HumanReviewReason,
		)

	actions := make(
		[]string,
		0,
		len(report.SuggestedActions),
	)

	for _, action := range report.SuggestedActions {
		action = strings.TrimSpace(action)

		if action != "" {
			actions = append(actions, action)
		}
	}

	report.SuggestedActions = actions

	for index := range report.References {
		report.References[index].Type =
			strings.ToLower(
				strings.TrimSpace(
					report.References[index].Type,
				),
			)

		report.References[index].ID =
			strings.TrimSpace(
				report.References[index].ID,
			)

		report.References[index].Title =
			strings.TrimSpace(
				report.References[index].Title,
			)
	}
}

func validateReport(report *models.FeedbackReport) error {
	if report == nil {
		return errors.New("report is nil")
	}

	if report.Summary == "" {
		return errors.New(
			"report summary is required",
		)
	}

	if !isValidReportCategory(report.Category) {
		return fmt.Errorf(
			"invalid report category: %q",
			report.Category,
		)
	}

	if !isValidUrgency(report.Urgency) {
		return fmt.Errorf(
			"invalid report urgency: %q",
			report.Urgency,
		)
	}

	if report.Confidence < 0 ||
		report.Confidence > 1 {
		return fmt.Errorf(
			"report confidence must be between 0 and 1: %.2f",
			report.Confidence,
		)
	}

	if report.CustomerContext.Available &&
		report.CustomerContext.Summary == "" {
		return errors.New(
			"customer context summary is required when customer data is available",
		)
	}

	if len(report.SuggestedActions) == 0 {
		return errors.New(
			"at least one suggested action is required",
		)
	}

	for index, reference := range report.References {
		if err := validateReference(reference); err != nil {
			return fmt.Errorf(
				"invalid reference at index %d: %w",
				index,
				err,
			)
		}
	}

	return nil
}

func isValidReportCategory(
	category string,
) bool {
	switch category {
	case "bug_report",
		"billing",
		"feature_request",
		"praise",
		"churn_risk",
		"policy_violation",
		"unknown":
		return true

	default:
		return false
	}
}

func isValidUrgency(urgency string) bool {
	switch urgency {
	case "low", "medium", "high":
		return true

	default:
		return false
	}
}

func validateReference(reference models.ReportReference) error {
	switch reference.Type {
	case "policy", "workflow":
	default:
		return fmt.Errorf(
			"unsupported reference type: %q",
			reference.Type,
		)
	}

	if reference.ID == "" {
		return errors.New(
			"reference ID is required",
		)
	}

	if reference.Title == "" {
		return errors.New(
			"reference title is required",
		)
	}

	return nil
}

func (r *JSONReporter) applyHumanReviewRules(report *models.FeedbackReport, classification models.Classification) {
	reasons := make([]string, 0)

	if classification.Confidence <
		r.humanReviewThreshold {
		reasons = append(
			reasons,
			fmt.Sprintf(
				"classification confidence is below %.2f",
				r.humanReviewThreshold,
			),
		)
	}

	if strings.EqualFold(
		string(classification.Category),
		"unknown",
	) {
		reasons = append(
			reasons,
			"feedback category is ambiguous",
		)
	}

	if report.Confidence <
		r.humanReviewThreshold {
		reasons = append(
			reasons,
			fmt.Sprintf(
				"report confidence is below %.2f",
				r.humanReviewThreshold,
			),
		)
	}

	classificationCategory := strings.ToLower(
		strings.TrimSpace(
			string(classification.Category),
		),
	)

	if report.Category != classificationCategory {
		reasons = append(
			reasons,
			fmt.Sprintf(
				"report category %q differs from classifier category %q",
				report.Category,
				classificationCategory,
			),
		)
	}

	if !report.CustomerContext.Available {
		reasons = append(
			reasons,
			"customer data is unavailable",
		)
	}

	if !hasReferenceType(
		report.References,
		"policy",
	) {
		reasons = append(
			reasons,
			"policy reference is unavailable",
		)
	}

	if !hasReferenceType(
		report.References,
		"workflow",
	) {
		reasons = append(
			reasons,
			"workflow reference is unavailable",
		)
	}

	if report.HumanReviewRequired &&
		report.HumanReviewReason != "" {
		reasons = append(
			reasons,
			report.HumanReviewReason,
		)
	}

	reasons = uniqueStrings(reasons)

	if len(reasons) == 0 {
		report.HumanReviewRequired = false
		report.HumanReviewReason = ""
		return
	}

	report.HumanReviewRequired = true
	report.HumanReviewReason = strings.Join(
		reasons,
		"; ",
	)
}

func hasReferenceType(references []models.ReportReference, referenceType string) bool {
	for _, reference := range references {
		if reference.Type == referenceType {
			return true
		}
	}

	return false
}

func uniqueStrings(values []string) []string {
	seen := make(map[string]struct{})
	result := make([]string, 0, len(values))

	for _, value := range values {
		value = strings.TrimSpace(value)

		if value == "" {
			continue
		}

		if _, exists := seen[value]; exists {
			continue
		}

		seen[value] = struct{}{}
		result = append(result, value)
	}

	return result
}
