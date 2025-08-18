package sdktypes

import (
	projectv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/projects/v1"
)

type (
	CheckViolation = projectv1.CheckViolation
	ViolationLevel = projectv1.CheckViolation_Level
)

const (
	ViolationError   ViolationLevel = projectv1.CheckViolation_LEVEL_ERROR
	ViolationWarning ViolationLevel = projectv1.CheckViolation_LEVEL_WARNING
)

func NewCheckViolation(filename string, ruleID string, message string) *CheckViolation {
	return &CheckViolation{
		Location: &CodeLocationPB{Path: filename},
		Level:    CheckRules[ruleID].Level,
		Message:  message,
		RuleId:   ruleID,
	}
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

	EmptyVariableRuleID     = "W1"
	NoTriggersDefinedRuleID = "W2"
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
