package sdktypes

import (
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	commonv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/common/v1"
)

type statusCodeTraits struct{}

var _ enumTraits = statusCodeTraits{}

func (statusCodeTraits) Prefix() string           { return "CODE_" }
func (statusCodeTraits) Names() map[int32]string  { return commonv1.Status_Code_name }
func (statusCodeTraits) Values() map[string]int32 { return commonv1.Status_Code_value }

type StatusCode struct {
	enum[statusCodeTraits, commonv1.Status_Code]
}

func statusCodeFromProto(e commonv1.Status_Code) StatusCode {
	return kittehs.Must1(StatusCodeFromProto(e))
}

var (
	PossibleStatusCodesNames = AllEnumNames[statusCodeTraits]()

	StatusCodeUnspecified = statusCodeFromProto(commonv1.Status_CODE_UNSPECIFIED)
	StatusCodeOK          = statusCodeFromProto(commonv1.Status_CODE_OK)
	StatusCodeWarning     = statusCodeFromProto(commonv1.Status_CODE_WARNING)
	StatusCodeError       = statusCodeFromProto(commonv1.Status_CODE_ERROR)
)

func StatusCodeFromProto(e commonv1.Status_Code) (StatusCode, error) {
	return EnumFromProto[StatusCode](e)
}

func ParseStatusCode(raw string) (StatusCode, error) {
	return ParseEnum[StatusCode](raw)
}
