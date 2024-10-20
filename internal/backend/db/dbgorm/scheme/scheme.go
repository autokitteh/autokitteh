package scheme

import (
	"encoding/json"
	"fmt"
	"time"

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

func UUIDOrNil(uuid sdktypes.UUID) *sdktypes.UUID {
	zero := sdktypes.UUID{}
	if uuid == zero {
		return nil
	}

	return &uuid
}

var Tables = []any{
	&Build{},
	&Connection{},
	&Deployment{},
	&Env{},
	&Event{},
	&Project{},
	&Secret{},
	&Session{},
	&SessionCallAttempt{},
	&SessionCallSpec{},
	&SessionLogRecord{},
	&Signal{},
	&Trigger{},
	&User{},
	&Var{},
}

type Build struct {
	BuildID   sdktypes.UUID  `gorm:"primaryKey;type:uuid;not null"`
	ProjectID *sdktypes.UUID `gorm:"index;type:uuid"`
	Data      []byte
	CreatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`

	// enforce foreign keys
	Project *Project
}

func ParseBuild(b Build) (sdktypes.Build, error) {
	build, err := sdktypes.StrictBuildFromProto(&sdktypes.BuildPB{
		BuildId:   sdktypes.IDFromUUID[sdktypes.BuildID](b.BuildID).String(),
		ProjectId: sdktypes.IDFromUUIDPtr[sdktypes.ProjectID](b.ProjectID).String(),
		CreatedAt: timestamppb.New(b.CreatedAt),
	})
	if err != nil {
		return sdktypes.InvalidBuild, fmt.Errorf("invalid record: %w", err)
	}

	return build, nil
}

type Connection struct {
	ConnectionID  sdktypes.UUID  `gorm:"primaryKey;type:uuid;not null"`
	IntegrationID *sdktypes.UUID `gorm:"index;type:uuid"`
	ProjectID     *sdktypes.UUID `gorm:"index;type:uuid"`
	Name          string
	StatusCode    int32 `gorm:"index"`
	StatusMessage string

	DeletedAt gorm.DeletedAt `gorm:"index"`

	// enforce foreign keys
	Project *Project

	// TODO(ENG-111): Also call "Preload()" where relevant
}

func ParseConnection(c Connection) (sdktypes.Connection, error) {
	conn, err := sdktypes.StrictConnectionFromProto(&sdktypes.ConnectionPB{
		ConnectionId:  sdktypes.IDFromUUID[sdktypes.ConnectionID](c.ConnectionID).String(),
		IntegrationId: sdktypes.IDFromUUIDPtr[sdktypes.IntegrationID](c.IntegrationID).String(),
		ProjectId:     sdktypes.IDFromUUIDPtr[sdktypes.ProjectID](c.ProjectID).String(),
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
	// varID is scopeID. just mapped directly for reusing the join code
	VarID         sdktypes.UUID `gorm:"primaryKey;index;type:uuid;not null"`
	ScopeID       sdktypes.UUID `gorm:"-"`
	Name          string        `gorm:"primaryKey;index;not null"`
	Value         string
	IsSecret      bool
	IntegrationID sdktypes.UUID `gorm:"index;type:uuid"` // var lookup by integration id
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
	ProjectID sdktypes.UUID `gorm:"primaryKey;type:uuid;not null"`
	Name      string        `gorm:"index;not null"`
	RootURL   string
	Resources []byte
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

func ParseProject(r Project) (sdktypes.Project, error) {
	p, err := sdktypes.StrictProjectFromProto(&sdktypes.ProjectPB{
		ProjectId: sdktypes.IDFromUUID[sdktypes.ProjectID](r.ProjectID).String(),
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
	EventID       sdktypes.UUID  `gorm:"uniqueIndex;type:uuid;not null"`
	DestinationID sdktypes.UUID  `gorm:"index;type:uuid;not null"`
	IntegrationID *sdktypes.UUID `gorm:"index;type:uuid"`
	ConnectionID  *sdktypes.UUID `gorm:"index;type:uuid"`
	TriggerID     *sdktypes.UUID `gorm:"index;type:uuid"`

	EventType string `gorm:"index:idx_event_type_seq,priority:1;index:idx_event_type"`
	Data      datatypes.JSON
	Memo      datatypes.JSON
	CreatedAt time.Time
	Seq       uint64 `gorm:"primaryKey;autoIncrement:true,index:idx_event_type_seq,priority:2"`

	DeletedAt gorm.DeletedAt `gorm:"index"`

	// enforce foreign keys
	Connection *Connection
	Trigger    *Trigger
}

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
		did = sdktypes.NewEventDestinationID(sdktypes.IDFromUUIDPtr[sdktypes.ConnectionID](uuid))
	} else if uuid := e.TriggerID; uuid != nil {
		did = sdktypes.NewEventDestinationID(sdktypes.IDFromUUIDPtr[sdktypes.TriggerID](uuid))
	} else {
		return sdktypes.InvalidEvent, sdkerrors.NewInvalidArgumentError("event must have a connection or trigger")
	}

	return sdktypes.StrictEventFromProto(&sdktypes.EventPB{
		EventId:       sdktypes.IDFromUUID[sdktypes.EventID](e.EventID).String(),
		EventType:     e.EventType,
		Data:          kittehs.TransformMapValues(data, sdktypes.ToProto),
		Memo:          memo,
		CreatedAt:     timestamppb.New(e.CreatedAt),
		Seq:           e.Seq,
		DestinationId: did.String(),
	})
}

type Env struct {
	EnvID     sdktypes.UUID `gorm:"primaryKey;type:uuid;not null"`
	ProjectID sdktypes.UUID `gorm:"index;type:uuid;not null"`
	Name      string
	DeletedAt gorm.DeletedAt `gorm:"index"`

	// {pid.uuid}/{name}. easier to detect dups.
	// See OrgMember for more.
	MembershipID string `gorm:"uniqueIndex"`

	// enforce foreign keys
	Project *Project
}

func ParseEnv(e Env) (sdktypes.Env, error) {
	return sdktypes.StrictEnvFromProto(&sdktypes.EnvPB{
		EnvId:     sdktypes.IDFromUUID[sdktypes.EnvID](e.EnvID).String(),
		ProjectId: sdktypes.IDFromUUID[sdktypes.ProjectID](e.ProjectID).String(),
		Name:      e.Name,
	})
}

type Trigger struct {
	TriggerID    sdktypes.UUID  `gorm:"primaryKey;type:uuid;not null"`
	ProjectID    sdktypes.UUID  `gorm:"index;type:uuid;not null"`
	ConnectionID *sdktypes.UUID `gorm:"index;type:uuid"`
	EnvID        sdktypes.UUID  `gorm:"index;type:uuid;not null"`

	SourceType   string `gorm:"index"`
	EventType    string
	Filter       string
	CodeLocation string

	Name string
	// Makes sure name is unique - this is the env_id with name.
	UniqueName string `gorm:"uniqueIndex;not null"` // env_id + name

	WebhookSlug string `gorm:"index"`
	Schedule    string

	DeletedAt gorm.DeletedAt `gorm:"index"`

	// enforce foreign keys
	Project    *Project
	Connection *Connection
	Env        *Env
}

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
		TriggerId:    sdktypes.IDFromUUID[sdktypes.TriggerID](e.TriggerID).String(),
		EnvId:        sdktypes.IDFromUUID[sdktypes.EnvID](e.EnvID).String(),
		SourceType:   srcType.ToProto(),
		ConnectionId: sdktypes.IDFromUUIDPtr[sdktypes.ConnectionID](e.ConnectionID).String(),
		EventType:    e.EventType,
		Filter:       e.Filter,
		CodeLocation: loc.ToProto(),
		Name:         e.Name,
		WebhookSlug:  e.WebhookSlug,
		Schedule:     e.Schedule,
	})
}

type SessionLogRecord struct {
	SessionID sdktypes.UUID `gorm:"primaryKey:SessionID;type:uuid;not null"`
	Seq       uint64        `gorm:"primaryKey;not null"`
	Data      datatypes.JSON
	Type      string `gorm:"index"`

	// enforce foreign keys
	Session *Session
}

func ParseSessionLogRecord(c SessionLogRecord) (spec sdktypes.SessionLogRecord, err error) {
	err = json.Unmarshal(c.Data, &spec)
	return
}

type SessionCallSpec struct {
	SessionID sdktypes.UUID `gorm:"primaryKey:SessionID;type:uuid;not null"`
	Seq       uint32        `gorm:"primaryKey"`
	Data      datatypes.JSON

	// enforce foreign keys
	Session *Session
}

func ParseSessionCallSpec(c SessionCallSpec) (spec sdktypes.SessionCallSpec, err error) {
	err = json.Unmarshal(c.Data, &spec)
	return
}

type SessionCallAttempt struct {
	SessionID sdktypes.UUID `gorm:"uniqueIndex:idx_session_id_seq_attempt,priority:1;type:uuid;not null"`
	Seq       uint32        `gorm:"uniqueIndex:idx_session_id_seq_attempt,priority:2"`
	Attempt   uint32        `gorm:"uniqueIndex:idx_session_id_seq_attempt,priority:3"`
	Start     datatypes.JSON
	Complete  datatypes.JSON

	// enforce foreign keys
	Session *Session
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
	SessionID        sdktypes.UUID  `gorm:"primaryKey;type:uuid;not null"`
	BuildID          *sdktypes.UUID `gorm:"index;type:uuid"`
	EnvID            *sdktypes.UUID `gorm:"index;type:uuid"`
	DeploymentID     *sdktypes.UUID `gorm:"index;type:uuid"`
	EventID          *sdktypes.UUID `gorm:"index;type:uuid"`
	CurrentStateType int            `gorm:"index"`
	Entrypoint       string
	Inputs           datatypes.JSON
	CreatedAt        time.Time
	UpdatedAt        time.Time
	DeletedAt        gorm.DeletedAt `gorm:"index"`
	Memo             datatypes.JSON

	// enforce foreign keys
	Build      *Build
	Env        *Env
	Deployment *Deployment
	Event      *Event `gorm:"references:EventID"`
}

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
		SessionId:    sdktypes.IDFromUUID[sdktypes.SessionID](s.SessionID).String(),
		BuildId:      sdktypes.IDFromUUIDPtr[sdktypes.BuildID](s.BuildID).String(),
		EnvId:        sdktypes.IDFromUUIDPtr[sdktypes.EnvID](s.EnvID).String(),
		DeploymentId: sdktypes.IDFromUUIDPtr[sdktypes.DeploymentID](s.DeploymentID).String(),
		EventId:      sdktypes.IDFromUUIDPtr[sdktypes.EventID](s.EventID).String(),
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
	DeploymentID sdktypes.UUID  `gorm:"primaryKey;type:uuid;not null"`
	EnvID        *sdktypes.UUID `gorm:"index;type:uuid"`
	BuildID      sdktypes.UUID  `gorm:"type:uuid;not null"`
	State        int32
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    gorm.DeletedAt `gorm:"index"`

	// enforce foreign keys
	Env   *Env
	Build *Build
}

func (d *Deployment) BeforeUpdate(tx *gorm.DB) (err error) {
	if tx.Statement.Changed() { // if any fields changed
		tx.Statement.SetColumn("UpdatedAt", time.Now())
	}
	return nil
}

func ParseDeployment(d Deployment) (sdktypes.Deployment, error) {
	deployment, err := sdktypes.StrictDeploymentFromProto(&sdktypes.DeploymentPB{
		DeploymentId: sdktypes.IDFromUUID[sdktypes.DeploymentID](d.DeploymentID).String(),
		BuildId:      sdktypes.IDFromUUID[sdktypes.BuildID](d.BuildID).String(),
		EnvId:        sdktypes.IDFromUUIDPtr[sdktypes.EnvID](d.EnvID).String(),
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
		DeploymentId: sdktypes.IDFromUUID[sdktypes.DeploymentID](d.DeploymentID).String(),
		BuildId:      sdktypes.IDFromUUID[sdktypes.BuildID](d.BuildID).String(),
		EnvId:        sdktypes.IDFromUUIDPtr[sdktypes.EnvID](d.EnvID).String(),
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
	SignalID      sdktypes.UUID  `gorm:"primaryKey;type:uuid;not null"`
	DestinationID sdktypes.UUID  `gorm:"index;type:uuid;not null"`
	ConnectionID  *sdktypes.UUID `gorm:"type:uuid"`
	TriggerID     *sdktypes.UUID `gorm:"type:uuid"`
	CreatedAt     time.Time
	WorkflowID    string
	Filter        string

	// enforce foreign key
	Connection *Connection
	Trigger    *Trigger
}

func ParseSignal(r *Signal) (*types.Signal, error) {
	var dstID sdktypes.EventDestinationID

	if r.ConnectionID != nil {
		dstID = sdktypes.NewEventDestinationID(sdktypes.IDFromUUIDPtr[sdktypes.ConnectionID](r.ConnectionID))
	} else if r.TriggerID != nil {
		dstID = sdktypes.NewEventDestinationID(sdktypes.IDFromUUIDPtr[sdktypes.TriggerID](r.TriggerID))
	} else {
		return nil, sdkerrors.NewInvalidArgumentError("signal must have a connection or trigger")
	}

	return &types.Signal{
		ID:            r.SignalID,
		DestinationID: dstID,
		WorkflowID:    r.WorkflowID,
		Filter:        r.Filter,
	}, nil
}

type User struct {
	UserID      sdktypes.UUID     `gorm:"primaryKey;type:uuid;not null"`
	Credentials []UserCredentials `gorm:"foreignKey:UserID"`
}

type UserCredentials struct {
	UserID     sdktypes.UUID `gorm:"primaryKey;type:uuid;not null"`
	Provider   string        `gorm:"primaryKey;type:string;not null"`
	ProviderID string        `gorm:"primaryKey;type:string;not null"`
}
