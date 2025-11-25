package sdktypes

import (
	"fmt"

	"google.golang.org/protobuf/proto"

	projectv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/projects/v1"
)

type (
	CheckViolation projectv1.CheckViolation
	ViolationLevel = projectv1.CheckViolation_Level
)

const (
	ViolationInfo    ViolationLevel = projectv1.CheckViolation_LEVEL_INFO
	ViolationError   ViolationLevel = projectv1.CheckViolation_LEVEL_ERROR
	ViolationWarning ViolationLevel = projectv1.CheckViolation_LEVEL_WARNING
)

func NewCheckViolationf(filename string, ruleID string, f string, vs ...any) *CheckViolation {
	return &CheckViolation{
		Location: &CodeLocationPB{Path: filename},
		Level:    CheckRules[ruleID].Level,
		Message:  fmt.Sprintf(f, vs...),
		RuleId:   ruleID,
	}
}

func (cv *CheckViolation) clone() *CheckViolation {
	return (*CheckViolation)(proto.CloneOf((*projectv1.CheckViolation)(cv)))
}

func (cv *CheckViolation) SetShortMessage(shortMsg string) *CheckViolation {
	cv = cv.clone()
	cv.ShortMessage = shortMsg
	return cv
}

func (cv *CheckViolation) SetSubject(subj string) *CheckViolation {
	cv = cv.clone()
	cv.Subject = subj
	return cv
}

func (cv *CheckViolation) SetSubjectf(format string, args ...any) *CheckViolation {
	return cv.SetSubject(fmt.Sprintf(format, args...))
}

const (
	ProjectSizeTooLargeRuleID     = "E1"
	DuplicateConnectionNameRuleID = "E2"
	DuplicateTriggerNameRuleID    = "E3"
	BadCallFormatRuleID           = "E4"
	FileNotFoundRuleID            = "E5"
	SyntaxErrorRuleID             = "E6"
	MissingHandlerRuleID          = "E7"
	NonexistingConnectionRuleID   = "E8"
	MalformedNameRuleID           = "E9"
	InvalidManifestRuleID         = "E10"
	FileCannotExportRuleID        = "E11"
	InvalidEventFilterRuleID      = "E12"
	InvalidPyRequirementsRuleID   = "E13"
	UnknownIntegrationRuleID      = "E14"

	NoTriggersDefinedRuleID                     = "W1"
	PyRequirementsPackageAlreadyInstalledRuleID = "W2"

	EmptyVariableRuleID = "I1"
)

type CheckRule struct {
	Title string
	Level ViolationLevel
}

var CheckRules = map[string]CheckRule{ // ID -> Description
	ProjectSizeTooLargeRuleID:     {"Project size too large", ViolationError},
	DuplicateConnectionNameRuleID: {"Duplicate connection name", ViolationError},
	DuplicateTriggerNameRuleID:    {"Duplicate trigger name", ViolationError},
	BadCallFormatRuleID:           {"Bad `call` format", ViolationError},
	FileNotFoundRuleID:            {"File not found", ViolationError},
	SyntaxErrorRuleID:             {"Syntax error", ViolationError},
	MissingHandlerRuleID:          {"Missing handler", ViolationError},
	NonexistingConnectionRuleID:   {"Nonexisting connection", ViolationError},
	MalformedNameRuleID:           {"Malformed name", ViolationError},
	InvalidManifestRuleID:         {"Invalid manifest", ViolationError},

	EmptyVariableRuleID:     {"Empty variable", ViolationWarning},
	NoTriggersDefinedRuleID: {"No triggers defined", ViolationWarning},
}
