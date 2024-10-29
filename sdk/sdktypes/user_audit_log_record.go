package sdktypes

import (
	"encoding/json"
	"errors"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"

	userv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/users/v1"
)

type UserAuditLogRecord struct {
	object[*UserAuditLogRecordPB, UserAuditLogRecordTraits]
}

var InvalidUserAuditLogRecord UserAuditLogRecord

type UserAuditLogRecordPB = userv1.UserAuditLogRecord

type UserAuditLogRecordTraits struct{}

func (UserAuditLogRecordTraits) Validate(m *UserAuditLogRecordPB) error {
	return errors.Join(
		idField[UserID]("id", m.UserId),
		jsonField("data", m.Data),
	)
}

func (UserAuditLogRecordTraits) StrictValidate(m *UserAuditLogRecordPB) error {
	return errors.Join(
		mandatory("id", m.UserId),
		mandatory("description", m.Description),
		mandatory("timestamp", m.Timestamp),
	)
}

func UserAuditLogRecordFromProto(m *UserAuditLogRecordPB) (UserAuditLogRecord, error) {
	return FromProto[UserAuditLogRecord](m)
}

func (u UserAuditLogRecord) Data() (data any) {
	kittehs.Must0(json.Unmarshal(u.read().Data, &data))
	return
}

func (u UserAuditLogRecord) Description() string  { return u.read().Description }
func (u UserAuditLogRecord) UserID() UserID       { return kittehs.Must1(ParseUserID(u.read().UserId)) }
func (u UserAuditLogRecord) Timestamp() time.Time { return u.read().Timestamp.AsTime() }
func (u UserAuditLogRecord) Error() string        { return u.read().Error }

func (u UserAuditLogRecord) WithError(err string) UserAuditLogRecord {
	return UserAuditLogRecord{u.forceUpdate(func(m *UserAuditLogRecordPB) {
		m.Error = err
	})}
}

func NewUserAuditLogRecord(uid UserID, desc string, data any) (UserAuditLogRecord, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return InvalidUserAuditLogRecord, sdkerrors.NewInvalidArgumentError("invalid data: %w", err)
	}

	return UserAuditLogRecordFromProto(&UserAuditLogRecordPB{
		UserId:      uid.String(),
		Description: desc,
		Data:        jsonData,
		Timestamp:   timestamppb.Now(),
	})
}
