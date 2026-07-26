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

	actions := make(
		[]string,
		0,
		len(report.RecommendedNextSteps),
	)

	for _, action := range report.RecommendedNextSteps {
		action = strings.TrimSpace(action)

		if action != "" {
			actions = append(actions, action)
		}
	}

	report.RecommendedNextSteps = actions

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

	if len(report.RecommendedNextSteps) == 0 {
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

func (r *JSONReporter) applyHumanReviewRules(
	report *models.FeedbackReport,
	classification models.Classification,
) {
	reasons := make([]models.ReviewReason, 0)

	categoryUnknown :=
		classification.Category == models.CategoryUnknown ||
			report.Category == string(models.CategoryUnknown)

	lowConfidence :=
		classification.Confidence < r.humanReviewThreshold ||
			report.Confidence < r.humanReviewThreshold

	if categoryUnknown {
		reasons = appendUnique(
			reasons,
			models.ReviewReasonAmbiguousCategory,
		)
	}

	if lowConfidence {
		reasons = appendUnique(
			reasons,
			models.ReviewReasonLowConfidence,
		)
	}

	if categoryUnknown || lowConfidence || !report.CustomerContext.Available {
		reasons = appendUnique(
			reasons,
			models.ReviewReasonMissingInformation,
		)
	}

	if !report.CustomerContext.Available {
		reasons = appendUnique(
			reasons,
			models.ReviewReasonCustomerNotFound,
		)
	}

	if requiresCategoryReferences(report) {
		if !hasReferenceType(report.References, "policy") {
			reasons = appendUnique(
				reasons,
				models.ReviewReasonMissingPolicy,
			)
		}

		if !hasReferenceType(report.References, "workflow") {
			reasons = appendUnique(
				reasons,
				models.ReviewReasonMissingWorkflow,
			)
		}
	}

	if containsRestrictedAction(report) {
		reasons = appendUnique(
			reasons,
			models.ReviewReasonRestrictedAction,
		)
	}

	report.HumanReviewReasons = reasons
	report.HumanReviewRequired = len(reasons) > 0
}

func requiresCategoryReferences(
	report *models.FeedbackReport,
) bool {
	if report.Category == string(models.CategoryUnknown) {
		return false
	}

	if report.Confidence < 0.6 {
		return false
	}

	return true
}

func containsRestrictedAction(report *models.FeedbackReport) bool {
	for _, action := range report.ConditionalActions {
		if isRestrictedText(action.Action) {
			return true
		}
	}

	for _, step := range report.RecommendedNextSteps {
		if isRestrictedText(step) {
			return true
		}
	}

	return false
}

func isRestrictedText(value string) bool {
	value = strings.ToLower(value)

	return containsAny(
		value,
		"refund",
		"issue credit",
		"compensate",
		"cancel subscription",
		"close account",
		"suspend account",
		"change account balance",
		"approve exception",
	)
}

func containsAny(value string, candidates ...string) bool {
	for _, candidate := range candidates {
		if strings.Contains(value, candidate) {
			return true
		}
	}

	return false
}

func appendUnique(
	values []models.ReviewReason,
	value models.ReviewReason,
) []models.ReviewReason {
	for _, existing := range values {
		if existing == value {
			return values
		}
	}

	return append(values, value)
}

func hasReferenceType(references []models.ReportReference, referenceType string) bool {
	for _, reference := range references {
		if reference.Type == referenceType {
			return true
		}
	}

	return false
}
