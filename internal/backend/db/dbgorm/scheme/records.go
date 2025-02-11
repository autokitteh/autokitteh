package scheme

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gorm.io/datatypes"
	"gorm.io/gorm"

	"go.autokitteh.dev/autokitteh/internal/backend/types"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	commonv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/common/v1"
	deploymentsv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/deployments/v1"
	sessionsv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/sessions/v1"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

// TODO: keep some log of actions performed. Something that
// can be used for recovery from unintended/malicious actions.

type Build struct {
	Base

	ProjectID uuid.UUID `gorm:"index;type:uuid;not null"`
	BuildID   uuid.UUID `gorm:"primaryKey;type:uuid;not null"`
	Data      []byte

	DeletedAt gorm.DeletedAt `gorm:"index"`

	// enforce foreign keys
	Project *Project
}

func (Build) IDFieldName() string { return "build_id" }

func ParseBuild(b Build) (sdktypes.Build, error) {
	build, err := sdktypes.StrictBuildFromProto(&sdktypes.BuildPB{
		BuildId:   sdktypes.NewIDFromUUID[sdktypes.BuildID](b.BuildID).String(),
		ProjectId: sdktypes.NewIDFromUUID[sdktypes.ProjectID](b.ProjectID).String(),
	})
	if err != nil {
		return sdktypes.InvalidBuild, fmt.Errorf("invalid record: %w", err)
	}

	return build, nil
}

type Connection struct {
	Base

	ProjectID uuid.UUID `gorm:"index;type:uuid;not null"`

	ConnectionID  uuid.UUID  `gorm:"primaryKey;type:uuid;not null"`
	IntegrationID *uuid.UUID `gorm:"index;type:uuid"`
	Name          string
	StatusCode    int32 `gorm:"index"`
	StatusMessage string

	UpdatedBy uuid.UUID `gorm:"type:uuid"`
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`

	// TODO(ENG-111): Also call "Preload()" where relevant

	Project *Project
}

func (Connection) IDFieldName() string { return "connection_id" }

func ParseConnection(c Connection) (sdktypes.Connection, error) {
	conn, err := sdktypes.StrictConnectionFromProto(&sdktypes.ConnectionPB{
		ConnectionId:  sdktypes.NewIDFromUUID[sdktypes.ConnectionID](c.ConnectionID).String(),
		IntegrationId: sdktypes.NewIDFromUUIDPtr[sdktypes.IntegrationID](c.IntegrationID).String(),
		ProjectId:     sdktypes.NewIDFromUUID[sdktypes.ProjectID](c.ProjectID).String(),
		Name:          c.Name,
		Status: &sdktypes.StatusPB{
			Code:    commonv1.Status_Code(c.StatusCode),
			Message: c.StatusMessage,
		},
	})
	if err != nil {
		return sdktypes.InvalidConnection, fmt.Errorf("invalid connection record: %w", err)
	}
	return conn, nil
}

type Var struct {
	Base

	// varID is scopeID. just mapped directly for reusing the join code
	VarID         uuid.UUID `gorm:"primaryKey;index;type:uuid;not null"`
	ScopeID       uuid.UUID `gorm:"-"`
	Name          string    `gorm:"primaryKey;index;not null"`
	Value         string
	IsSecret      bool
	IntegrationID uuid.UUID `gorm:"index;type:uuid"` // var lookup by integration id

	UpdatedBy uuid.UUID `gorm:"type:uuid"`
	UpdatedAt time.Time
}

// simple hook to populate ScopeID after retrieving a Var from the database
func (v *Var) AfterFind(tx *gorm.DB) (err error) {
	v.ScopeID = v.VarID
	return nil
}

func (v *Var) BeforeCreate(tx *gorm.DB) (err error) {
	if v.VarID != v.ScopeID {
		return gorm.ErrInvalidField
	}
	return nil
}

type Project struct {
	Base

	ProjectID uuid.UUID `gorm:"primaryKey;type:uuid;not null"`
	OrgID     uuid.UUID `gorm:"index;type:uuid"` // TODO(authz-migration): not null.
	Name      string    `gorm:"index;not null"`
	RootURL   string
	Resources []byte

	UpdatedBy uuid.UUID `gorm:"type:uuid"`
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

func (Project) IDFieldName() string { return "project_id" }

func ParseProject(r Project) (sdktypes.Project, error) {
	p, err := sdktypes.StrictProjectFromProto(&sdktypes.ProjectPB{
		ProjectId: sdktypes.NewIDFromUUID[sdktypes.ProjectID](r.ProjectID).String(),
		OrgId:     sdktypes.NewIDFromUUID[sdktypes.OrgID](r.OrgID).String(),
		Name:      r.Name,
	})
	if err != nil {
		return sdktypes.InvalidProject, fmt.Errorf("invalid project record: %w", err)
	}

	return p, nil
}

// Secret is a database table that simply stores sensitive key-value
// pairs, for usage by the "db" mode of autokitteh's secrets manager.
// WARNING: This is not secure in any way, and not durable by default.
// It is intended only for temporary, local, non-production purposes.
type Secret struct {
	Key   string `gorm:"primaryKey"`
	Value string `gorm:"type:string"`
}

type Event struct {
	Base

	ProjectID     uuid.UUID  `gorm:"index;type:uuid"` // TODO(authz-migration): not null.
	EventID       uuid.UUID  `gorm:"uniqueIndex;type:uuid;not null"`
	DestinationID uuid.UUID  `gorm:"index;type:uuid;not null"`
	IntegrationID *uuid.UUID `gorm:"index;type:uuid"`
	ConnectionID  *uuid.UUID `gorm:"index;type:uuid"`
	TriggerID     *uuid.UUID `gorm:"index;type:uuid"`

	EventType string `gorm:"index:idx_event_type_seq,priority:1;index:idx_event_type"`
	Data      datatypes.JSON
	Memo      datatypes.JSON
	Seq       uint64 `gorm:"primaryKey;autoIncrement:true,index:idx_event_type_seq,priority:2"`

	// enforce foreign keys
	Connection *Connection `gorm:"constraint:OnDelete:SET NULL"`
	Trigger    *Trigger    `gorm:"constraint:OnDelete:SET NULL"`

	Project *Project
}

func (Event) IDFieldName() string { return "event_id" }

func ParseEvent(e Event) (sdktypes.Event, error) {
	var data map[string]sdktypes.Value
	if len(e.Data) != 0 {
		if err := json.Unmarshal(e.Data, &data); err != nil {
			return sdktypes.InvalidEvent, fmt.Errorf("event data: %w", err)
		}
	}

	var memo map[string]string
	if err := json.Unmarshal(e.Memo, &memo); err != nil {
		return sdktypes.InvalidEvent, fmt.Errorf("event memo: %w", err)
	}

	var did sdktypes.EventDestinationID

	if uuid := e.ConnectionID; uuid != nil {
		did = sdktypes.NewEventDestinationID(sdktypes.NewIDFromUUIDPtr[sdktypes.ConnectionID](uuid))
	} else if uuid := e.TriggerID; uuid != nil {
		did = sdktypes.NewEventDestinationID(sdktypes.NewIDFromUUIDPtr[sdktypes.TriggerID](uuid))
	} else {
		return sdktypes.InvalidEvent, sdkerrors.NewInvalidArgumentError("event must have a connection or trigger")
	}

	return sdktypes.StrictEventFromProto(&sdktypes.EventPB{
		EventId:       sdktypes.NewIDFromUUID[sdktypes.EventID](e.EventID).String(),
		EventType:     e.EventType,
		Data:          kittehs.TransformMapValues(data, sdktypes.ToProto),
		Memo:          memo,
		CreatedAt:     timestamppb.New(e.CreatedAt),
		Seq:           e.Seq,
		DestinationId: did.String(),
	})
}

type Trigger struct {
	Base

	ProjectID    uuid.UUID  `gorm:"index;type:uuid;not null"`
	TriggerID    uuid.UUID  `gorm:"primaryKey;type:uuid;not null"`
	ConnectionID *uuid.UUID `gorm:"index;type:uuid"`

	SourceType   string `gorm:"index"`
	EventType    string
	Filter       string
	CodeLocation string

	Name string
	// Makes sure name is unique - this is the project_id with name.
	UniqueName string `gorm:"uniqueIndex;not null"` // project_id + name

	WebhookSlug string `gorm:"index"`
	Schedule    string

	UpdatedBy uuid.UUID `gorm:"type:uuid"`
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`

	// enforce foreign keys
	Connection *Connection
	Project    *Project
}

func (Trigger) IDFieldName() string { return "trigger_id" }

func ParseTrigger(e Trigger) (sdktypes.Trigger, error) {
	loc, err := sdktypes.ParseCodeLocation(e.CodeLocation)
	if err != nil {
		return sdktypes.InvalidTrigger, fmt.Errorf("loc: %w", err)
	}

	srcType, err := sdktypes.ParseTriggerSourceType(e.SourceType)
	if err != nil {
		return sdktypes.InvalidTrigger, fmt.Errorf("source type: %w", err)
	}

	if srcType == sdktypes.TriggerSourceTypeUnspecified {
		srcType = sdktypes.TriggerSourceTypeConnection
	}

	return sdktypes.StrictTriggerFromProto(&sdktypes.TriggerPB{
		TriggerId:    sdktypes.NewIDFromUUID[sdktypes.TriggerID](e.TriggerID).String(),
		SourceType:   srcType.ToProto(),
		ConnectionId: sdktypes.NewIDFromUUIDPtr[sdktypes.ConnectionID](e.ConnectionID).String(),
		ProjectId:    sdktypes.NewIDFromUUID[sdktypes.ProjectID](e.ProjectID).String(),
		EventType:    e.EventType,
		Filter:       e.Filter,
		CodeLocation: loc.ToProto(),
		Name:         e.Name,
		WebhookSlug:  e.WebhookSlug,
		Schedule:     e.Schedule,
	})
}

type SessionLogRecord struct {
	SessionID uuid.UUID `gorm:"primaryKey:SessionID;type:uuid;not null"`
	Seq       uint64    `gorm:"primaryKey;not null"`
	Data      datatypes.JSON
	Type      string `gorm:"index"`

	// enforce foreign keys
	Session *Session `gorm:"constraint:OnDelete:CASCADE"`
}

func ParseSessionLogRecord(c SessionLogRecord) (spec sdktypes.SessionLogRecord, err error) {
	err = json.Unmarshal(c.Data, &spec)
	return
}

type SessionCallSpec struct {
	SessionID uuid.UUID `gorm:"primaryKey:SessionID;type:uuid;not null"`
	Seq       uint32    `gorm:"primaryKey"`
	Data      datatypes.JSON

	// enforce foreign keys
	Session *Session `gorm:"constraint:OnDelete:CASCADE"`
}

func ParseSessionCallSpec(c SessionCallSpec) (spec sdktypes.SessionCallSpec, err error) {
	err = json.Unmarshal(c.Data, &spec)
	return
}

type SessionCallAttempt struct {
	SessionID uuid.UUID `gorm:"uniqueIndex:idx_session_id_seq_attempt,priority:1;type:uuid;not null"`
	Seq       uint32    `gorm:"uniqueIndex:idx_session_id_seq_attempt,priority:2"`
	Attempt   uint32    `gorm:"uniqueIndex:idx_session_id_seq_attempt,priority:3"`
	Start     datatypes.JSON
	Complete  datatypes.JSON

	// enforce foreign keys
	Session *Session `gorm:"constraint:OnDelete:CASCADE"`
}

func ParseSessionCallAttemptStart(c SessionCallAttempt) (d sdktypes.SessionCallAttemptStart, err error) {
	err = json.Unmarshal(c.Start, &d)
	return
}

func ParseSessionCallAttemptComplete(c SessionCallAttempt) (d sdktypes.SessionCallAttemptComplete, err error) {
	err = json.Unmarshal(c.Complete, &d)
	return
}

type Session struct {
	Base

	ProjectID        uuid.UUID  `gorm:"index;type:uuid;not null"`
	SessionID        uuid.UUID  `gorm:"primaryKey;type:uuid;not null"`
	BuildID          uuid.UUID  `gorm:"index;type:uuid;not null"`
	DeploymentID     *uuid.UUID `gorm:"index;type:uuid"`
	EventID          *uuid.UUID `gorm:"index;type:uuid"`
	CurrentStateType int        `gorm:"index"`
	Entrypoint       string
	Inputs           datatypes.JSON
	Memo             datatypes.JSON

	UpdatedBy uuid.UUID `gorm:"type:uuid"`
	UpdatedAt time.Time

	// enforce foreign keys
	Build      *Build
	Deployment *Deployment
	Project    *Project
	Event      *Event `gorm:"references:EventID;constraint:OnDelete:SET NULL"`
}

func (Session) IDFieldName() string { return "session_id" }

func ParseSession(s Session) (sdktypes.Session, error) {
	ep, err := sdktypes.ParseCodeLocation(s.Entrypoint)
	if err != nil {
		return sdktypes.InvalidSession, fmt.Errorf("entrypoint: %w", err)
	}

	var inputs map[string]sdktypes.Value
	if len(s.Inputs) != 0 {
		if err := json.Unmarshal(s.Inputs, &inputs); err != nil {
			return sdktypes.InvalidSession, fmt.Errorf("inputs: %w", err)
		}
	}

	var memo map[string]string
	if len(s.Memo) != 0 {
		if err := json.Unmarshal(s.Memo, &memo); err != nil {
			return sdktypes.InvalidSession, fmt.Errorf("memo: %w", err)
		}
	}

	session, err := sdktypes.StrictSessionFromProto(&sdktypes.SessionPB{
		SessionId:    sdktypes.NewIDFromUUID[sdktypes.SessionID](s.SessionID).String(),
		BuildId:      sdktypes.NewIDFromUUID[sdktypes.BuildID](s.BuildID).String(),
		ProjectId:    sdktypes.NewIDFromUUID[sdktypes.ProjectID](s.ProjectID).String(),
		DeploymentId: sdktypes.NewIDFromUUIDPtr[sdktypes.DeploymentID](s.DeploymentID).String(),
		EventId:      sdktypes.NewIDFromUUIDPtr[sdktypes.EventID](s.EventID).String(),
		Entrypoint:   ep.ToProto(),
		Inputs:       kittehs.TransformMapValues(inputs, sdktypes.ToProto),
		CreatedAt:    timestamppb.New(s.CreatedAt),
		UpdatedAt:    timestamppb.New(s.UpdatedAt),
		State:        sessionsv1.SessionStateType(s.CurrentStateType),
		Memo:         memo,
	})
	if err != nil {
		return sdktypes.InvalidSession, err
	}

	return session, err
}

type Deployment struct {
	Base

	ProjectID    uuid.UUID `gorm:"index;type:uuid;not null"`
	DeploymentID uuid.UUID `gorm:"primaryKey;type:uuid;not null"`
	BuildID      uuid.UUID `gorm:"type:uuid;not null"`
	State        int32     `gorm:"index"`

	UpdatedBy uuid.UUID `gorm:"type:uuid"`
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`

	// enforce foreign keys
	Build   *Build
	Project *Project
}

func (Deployment) IDFieldName() string { return "deployment_id" }

func (d *Deployment) BeforeUpdate(tx *gorm.DB) (err error) {
	if tx.Statement.Changed() { // if any fields changed
		tx.Statement.SetColumn("UpdatedAt", kittehs.Now())
	}
	return nil
}

func ParseDeployment(d Deployment) (sdktypes.Deployment, error) {
	deployment, err := sdktypes.StrictDeploymentFromProto(&sdktypes.DeploymentPB{
		DeploymentId: sdktypes.NewIDFromUUID[sdktypes.DeploymentID](d.DeploymentID).String(),
		BuildId:      sdktypes.NewIDFromUUID[sdktypes.BuildID](d.BuildID).String(),
		ProjectId:    sdktypes.NewIDFromUUID[sdktypes.ProjectID](d.ProjectID).String(),
		State:        deploymentsv1.DeploymentState(d.State),
		CreatedAt:    timestamppb.New(d.CreatedAt),
		UpdatedAt:    timestamppb.New(d.UpdatedAt),
	})
	if err != nil {
		return sdktypes.InvalidDeployment, fmt.Errorf("invalid record: %w", err)
	}

	return deployment, nil
}

type DeploymentWithStats struct {
	Deployment
	Created   uint32
	Running   uint32
	Error     uint32
	Completed uint32
	Stopped   uint32
}

func ParseDeploymentWithSessionStats(d DeploymentWithStats) (sdktypes.Deployment, error) {
	deployment, err := sdktypes.StrictDeploymentFromProto(&sdktypes.DeploymentPB{
		DeploymentId: sdktypes.NewIDFromUUID[sdktypes.DeploymentID](d.DeploymentID).String(),
		BuildId:      sdktypes.NewIDFromUUID[sdktypes.BuildID](d.BuildID).String(),
		ProjectId:    sdktypes.NewIDFromUUID[sdktypes.ProjectID](d.ProjectID).String(),
		State:        deploymentsv1.DeploymentState(d.State),
		CreatedAt:    timestamppb.New(d.CreatedAt),
		UpdatedAt:    timestamppb.New(d.UpdatedAt),
		SessionsStats: []*deploymentsv1.Deployment_SessionStats{
			{
				Count: d.Created,
				State: sdktypes.SessionStateTypeCreated.ToProto(),
			},
			{
				Count: d.Running,
				State: sdktypes.SessionStateTypeRunning.ToProto(),
			},
			{
				Count: d.Error,
				State: sdktypes.SessionStateTypeError.ToProto(),
			},
			{
				Count: d.Completed,
				State: sdktypes.SessionStateTypeCompleted.ToProto(),
			},
			{
				Count: d.Stopped,
				State: sdktypes.SessionStateTypeStopped.ToProto(),
			},
		},
	})
	if err != nil {
		return sdktypes.InvalidDeployment, fmt.Errorf("invalid record: %w", err)
	}

	return deployment, nil
}

type Signal struct {
	SignalID      uuid.UUID  `gorm:"primaryKey;type:uuid;not null"`
	DestinationID uuid.UUID  `gorm:"index;type:uuid;not null"`
	ConnectionID  *uuid.UUID `gorm:"type:uuid"`
	TriggerID     *uuid.UUID `gorm:"type:uuid"`
	CreatedAt     time.Time
	WorkflowID    string
	Filter        string

	// enforce foreign key
	Connection *Connection
	Trigger    *Trigger
}

func ParseSignal(r *Signal) (*types.Signal, error) {
	var dstID sdktypes.EventDestinationID

	switch {
	case r.ConnectionID != nil:
		dstID = sdktypes.NewEventDestinationID(sdktypes.NewIDFromUUIDPtr[sdktypes.ConnectionID](r.ConnectionID))
	case r.TriggerID != nil:
		dstID = sdktypes.NewEventDestinationID(sdktypes.NewIDFromUUIDPtr[sdktypes.TriggerID](r.TriggerID))
	default:
		return nil, sdkerrors.NewInvalidArgumentError("signal must have a connection or trigger")
	}

	return &types.Signal{
		ID:            r.SignalID,
		DestinationID: dstID,
		WorkflowID:    r.WorkflowID,
		Filter:        r.Filter,
	}, nil
}

type Value struct {
	Base

	ProjectID uuid.UUID `gorm:"index;type:uuid;not null"`
	Key       string    `gorm:"primaryKey;not null"`
	Value     []byte

	UpdatedBy uuid.UUID `gorm:"type:uuid"`
	UpdatedAt time.Time

	Project *Project
}

type User struct {
	Base

	UserID       uuid.UUID `gorm:"primaryKey;type:uuid;not null"`
	Email        string    `gorm:"uniqueIndex;not null"`
	DisplayName  string
	Disabled     bool      // deprecated, leave for backward compatibility.
	Status       int32     `gorm:"index"`
	DefaultOrgID uuid.UUID `gorm:"type:uuid"`

	UpdatedBy uuid.UUID `gorm:"type:uuid"`
	UpdatedAt time.Time
}

func ParseUser(r User) (sdktypes.User, error) {
	s := sdktypes.UserStatusActive

	if r.Status == 0 {
		// legacy.
		if r.Disabled {
			s = sdktypes.UserStatusDisabled
		}
	} else {
		var err error
		s, err = sdktypes.UserStatusFromProto(sdktypes.UserStatusPB(r.Status))
		if err != nil {
			return sdktypes.InvalidUser, fmt.Errorf("invalid user status: %w", err)
		}
	}

	return sdktypes.NewUser().
			WithID(sdktypes.NewIDFromUUID[sdktypes.UserID](r.UserID)).
			WithDisplayName(r.DisplayName).
			WithDefaultOrgID(sdktypes.NewIDFromUUID[sdktypes.OrgID](r.DefaultOrgID)).
			WithEmail(r.Email).
			WithStatus(s),
		nil
}

type Org struct {
	Base

	OrgID       uuid.UUID `gorm:"primaryKey;type:uuid;not null"`
	DisplayName string
	Name        string `gorm:"index"`

	UpdatedBy uuid.UUID `gorm:"type:uuid"`
	UpdatedAt time.Time
}

func ParseOrg(r Org) (sdktypes.Org, error) {
	n, err := sdktypes.ParseSymbol(r.Name)
	if err != nil {
		return sdktypes.InvalidOrg, fmt.Errorf("invalid org name: %w", err)
	}

	return sdktypes.NewOrg().
			WithID(sdktypes.NewIDFromUUID[sdktypes.OrgID](r.OrgID)).
			WithDisplayName(r.DisplayName).
			WithName(n),
		nil
}

type OrgMember struct {
	Base

	OrgID  uuid.UUID      `gorm:"primaryKey;type:uuid;not null"`
	UserID uuid.UUID      `gorm:"primaryKey;type:uuid;not null"`
	Status int            `gorm:"index"`
	Roles  datatypes.JSON // [str]

	UpdatedBy uuid.UUID `gorm:"type:uuid"`
	UpdatedAt time.Time

	Org  *Org
	User *User
}

func ParseOrgMember(r OrgMember) (sdktypes.OrgMember, error) {
	roles := make([]sdktypes.Symbol, 0)
	if len(r.Roles) > 0 {
		if err := json.Unmarshal(r.Roles, &roles); err != nil {
			return sdktypes.InvalidOrgMember, fmt.Errorf("roles: %w", err)
		}
	}

	s, err := sdktypes.OrgMemberStatusFromProto(sdktypes.OrgMemberStatusPB(r.Status))
	if err != nil {
		return sdktypes.InvalidOrgMember, fmt.Errorf("status: %w", err)
	}

	return sdktypes.NewOrgMember(
		sdktypes.NewIDFromUUID[sdktypes.OrgID](r.OrgID),
		sdktypes.NewIDFromUUID[sdktypes.UserID](r.UserID),
	).WithStatus(s).WithRoles(roles...), nil
}

// TODO: Remove after migration to new ownership is done.
type Ownership struct {
	EntityID   uuid.UUID `gorm:"primaryKey;type:uuid;not null"`
	EntityType string    `gorm:"not null"`

	UserID string `gorm:"not null"`
}
