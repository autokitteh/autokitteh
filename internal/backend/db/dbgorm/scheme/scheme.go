package scheme

import (
	"encoding/json"
	"fmt"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"
	"gorm.io/datatypes"
	"gorm.io/gorm"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	commonv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/common/v1"
	deploymentsv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/deployments/v1"
	eventsv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/events/v1"
	integrationsv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/integrations/v1"
	sessionsv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/sessions/v1"
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
	&Var{},
	&Deployment{},
	&Env{},
	&Event{},
	&EventRecord{},
	&Integration{},
	&Project{},
	&Secret{},
	&Session{},
	&SessionCallAttempt{},
	&SessionCallSpec{},
	&SessionLogRecord{},
	&Signal{},
	&Trigger{},
}

type Build struct {
	BuildID   sdktypes.UUID `gorm:"primaryKey;type:uuid;not null"`
	Data      []byte
	CreatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

func ParseBuild(b Build) (sdktypes.Build, error) {
	build, err := sdktypes.StrictBuildFromProto(&sdktypes.BuildPB{
		BuildId:   sdktypes.NewIDFromUUID[sdktypes.BuildID](&b.BuildID).String(),
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

	// enforce foreign keys
	// Integration *Integration FIXME: ENG-590
	Project *Project

	// TODO(ENG-111): Also call "Preload()" where relevant
}

func ParseConnection(c Connection) (sdktypes.Connection, error) {
	conn, err := sdktypes.StrictConnectionFromProto(&sdktypes.ConnectionPB{
		ConnectionId:  sdktypes.NewIDFromUUID[sdktypes.ConnectionID](&c.ConnectionID).String(),
		IntegrationId: sdktypes.NewIDFromUUID[sdktypes.IntegrationID](c.IntegrationID).String(),
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
	ScopeID  sdktypes.UUID `gorm:"primaryKey;index;type:uuid;not null"`
	Name     string        `gorm:"primaryKey;index"`
	Value    string
	IsSecret bool

	IntegrationID sdktypes.UUID `gorm:"index;type:uuid"`

	// enforce foreign keys
	// Integration *Integration // FIXME: ENG-590
}

type Integration struct {
	// Unique internal identifier.
	IntegrationID sdktypes.UUID `gorm:"primaryKey;type:uuid;not null"`

	// Unique external (and URL-safe) identifier.
	UniqueName string `gorm:"uniqueIndex"`

	// Optional user-facing metadata.

	DisplayName string
	Description string
	LogoURL     string
	UserLinks   datatypes.JSON
	// TODO: Tags

	// TODO(ENG-346): Connection UI specification instead of a URL.
	ConnectionURL string

	// TODO: Functions

	// TODO: Events

	// TODO(ENG-112): https://gorm.io/docs/models.html#gorm-Model?

	APIKey     string
	SigningKey string
}

func ParseIntegration(i Integration) (sdktypes.Integration, error) {
	var uls map[string]string
	err := json.Unmarshal(i.UserLinks, &uls)
	if err != nil {
		return sdktypes.InvalidIntegration, fmt.Errorf("integration user links: %w", err)
	}

	integ, err := sdktypes.StrictIntegrationFromProto(&integrationsv1.Integration{
		IntegrationId: sdktypes.NewIDFromUUID[sdktypes.IntegrationID](&i.IntegrationID).String(),
		UniqueName:    i.UniqueName,
		DisplayName:   i.DisplayName,
		Description:   i.Description,
		LogoUrl:       i.LogoURL,
		UserLinks:     uls,
		// TODO: Tags
		// TODO(ENG-346): Connection UI specification instead of a URL.
		ConnectionUrl: i.ConnectionURL,
		// TODO: Functions
		// TODO: Events
	})
	if err != nil {
		return sdktypes.InvalidIntegration, fmt.Errorf("invalid integration record: %w", err)
	}
	return integ, nil
}

type Project struct {
	ProjectID sdktypes.UUID `gorm:"primaryKey;type:uuid;not null"`
	Name      string        `gorm:"uniqueIndex"`
	RootURL   string
	Resources []byte
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

func ParseProject(r Project) (sdktypes.Project, error) {
	p, err := sdktypes.StrictProjectFromProto(&sdktypes.ProjectPB{
		ProjectId: sdktypes.NewIDFromUUID[sdktypes.ProjectID](&r.ProjectID).String(),
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
	IntegrationID *sdktypes.UUID `gorm:"index;type:uuid"`
	ConnectionID  *sdktypes.UUID `gorm:"index;type:uuid"`

	EventType string `gorm:"index:idx_event_type_seq,priority:1;index:idx_event_type"`
	Data      datatypes.JSON
	Memo      datatypes.JSON
	CreatedAt time.Time
	Seq       uint64 `gorm:"primaryKey;autoIncrement:true,index:idx_event_type_seq,priority:2"`

	DeletedAt gorm.DeletedAt `gorm:"index"`

	// enforce foreign keys
	// Integration *Integration // FIXME: ENG-590
	Connection *Connection
}

func ParseEvent(e Event) (sdktypes.Event, error) {
	var data map[string]sdktypes.Value
	if err := json.Unmarshal(e.Data, &data); err != nil {
		return sdktypes.InvalidEvent, fmt.Errorf("event data: %w", err)
	}

	var memo map[string]string
	if err := json.Unmarshal(e.Memo, &memo); err != nil {
		return sdktypes.InvalidEvent, fmt.Errorf("event memo: %w", err)
	}

	return sdktypes.StrictEventFromProto(&sdktypes.EventPB{
		EventId:      sdktypes.NewIDFromUUID[sdktypes.EventID](&e.EventID).String(),
		ConnectionId: sdktypes.NewIDFromUUID[sdktypes.ConnectionID](e.ConnectionID).String(),
		EventType:    e.EventType,
		Data:         kittehs.TransformMapValues(data, sdktypes.ToProto),
		Memo:         memo,
		CreatedAt:    timestamppb.New(e.CreatedAt),
		Seq:          e.Seq,
	})
}

type EventRecord struct {
	Seq       uint32        `gorm:"primaryKey"`
	EventID   sdktypes.UUID `gorm:"primaryKey;type:uuid;not null"`
	State     int32         `gorm:"index"`
	CreatedAt time.Time

	// enforce foreign keys
	Event *Event `gorm:"references:EventID"`
}

func ParseEventRecord(e EventRecord) (sdktypes.EventRecord, error) {
	return sdktypes.StrictEventRecordFromProto(&sdktypes.EventRecordPB{
		Seq:       e.Seq,
		EventId:   sdktypes.NewIDFromUUID[sdktypes.EventID](&e.EventID).String(),
		State:     eventsv1.EventState(e.State),
		CreatedAt: timestamppb.New(e.CreatedAt),
	})
}

type Env struct {
	EnvID     sdktypes.UUID  `gorm:"primaryKey;type:uuid;not null"`
	ProjectID *sdktypes.UUID `gorm:"index;type:uuid"`
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
		EnvId:     sdktypes.NewIDFromUUID[sdktypes.EnvID](&e.EnvID).String(),
		ProjectId: sdktypes.NewIDFromUUID[sdktypes.ProjectID](e.ProjectID).String(),
		Name:      e.Name,
	})
}

type Trigger struct {
	TriggerID sdktypes.UUID `gorm:"primaryKey;type:uuid;not null"`

	ProjectID    sdktypes.UUID `gorm:"index;type:uuid;not null"`
	ConnectionID sdktypes.UUID `gorm:"index;type:uuid;not null"`
	EnvID        sdktypes.UUID `gorm:"index;type:uuid;not null"`
	Name         string
	EventType    string
	Filter       string
	CodeLocation string
	Data         datatypes.JSON

	// enforce foreign keys
	Project    *Project
	Env        *Env
	Connection *Connection

	// Makes sure name is unique - this is the env_id with name.
	// If name is emptyy, will be env_id with a random string.
	UniqueName string `gorm:"uniqueIndex"`
}

func ParseTrigger(e Trigger) (sdktypes.Trigger, error) {
	loc, err := sdktypes.ParseCodeLocation(e.CodeLocation)
	if err != nil {
		return sdktypes.InvalidTrigger, fmt.Errorf("loc: %w", err)
	}

	var data map[string]sdktypes.Value
	if err := json.Unmarshal(e.Data, &data); err != nil {
		return sdktypes.InvalidTrigger, fmt.Errorf("data: %w", err)
	}

	return sdktypes.StrictTriggerFromProto(&sdktypes.TriggerPB{
		TriggerId:    sdktypes.NewIDFromUUID[sdktypes.TriggerID](&e.TriggerID).String(),
		EnvId:        sdktypes.NewIDFromUUID[sdktypes.EnvID](&e.EnvID).String(),
		ConnectionId: sdktypes.NewIDFromUUID[sdktypes.ConnectionID](&e.ConnectionID).String(),
		EventType:    e.EventType,
		Filter:       e.Filter,
		CodeLocation: loc.ToProto(),
		Name:         e.Name,
		Data:         kittehs.TransformMapValues(data, sdktypes.ToProto),
	})
}

type SessionLogRecord struct {
	SessionID sdktypes.UUID `gorm:"index;type:uuid;not null"`
	Data      datatypes.JSON

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

	if err := json.Unmarshal(s.Inputs, &inputs); err != nil {
		return sdktypes.InvalidSession, fmt.Errorf("inputs: %w", err)
	}

	session, err := sdktypes.StrictSessionFromProto(&sdktypes.SessionPB{
		SessionId:    sdktypes.NewIDFromUUID[sdktypes.SessionID](&s.SessionID).String(),
		BuildId:      sdktypes.NewIDFromUUID[sdktypes.BuildID](s.BuildID).String(),
		EnvId:        sdktypes.NewIDFromUUID[sdktypes.EnvID](s.EnvID).String(),
		DeploymentId: sdktypes.NewIDFromUUID[sdktypes.DeploymentID](s.DeploymentID).String(),
		EventId:      sdktypes.NewIDFromUUID[sdktypes.EventID](s.EventID).String(),
		Entrypoint:   ep.ToProto(),
		Inputs:       kittehs.TransformMapValues(inputs, sdktypes.ToProto),
		CreatedAt:    timestamppb.New(s.CreatedAt),
		UpdatedAt:    timestamppb.New(s.UpdatedAt),
		State:        sessionsv1.SessionStateType(s.CurrentStateType),
	})
	if err != nil {
		return sdktypes.InvalidSession, err
	}

	return session, err
}

type Deployment struct {
	DeploymentID sdktypes.UUID  `gorm:"primaryKey;type:uuid;not null"`
	EnvID        *sdktypes.UUID `gorm:"index;type:uuid"`
	BuildID      *sdktypes.UUID `gorm:"type:uuid"`
	State        int32
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    gorm.DeletedAt `gorm:"index"`

	// enforce foreign keys
	Env   *Env
	Build *Build
}

func ParseDeployment(d Deployment) (sdktypes.Deployment, error) {
	deployment, err := sdktypes.StrictDeploymentFromProto(&sdktypes.DeploymentPB{
		DeploymentId: sdktypes.NewIDFromUUID[sdktypes.DeploymentID](&d.DeploymentID).String(),
		BuildId:      sdktypes.NewIDFromUUID[sdktypes.BuildID](d.BuildID).String(),
		EnvId:        sdktypes.NewIDFromUUID[sdktypes.EnvID](d.EnvID).String(),
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
		DeploymentId: sdktypes.NewIDFromUUID[sdktypes.DeploymentID](&d.DeploymentID).String(),
		BuildId:      sdktypes.NewIDFromUUID[sdktypes.BuildID](d.BuildID).String(),
		EnvId:        sdktypes.NewIDFromUUID[sdktypes.EnvID](d.EnvID).String(),
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
	SignalID     string        `gorm:"primaryKey"`
	ConnectionID sdktypes.UUID `gorm:"index:idx_connection_id_event_type;type:uuid;not null"`
	CreatedAt    time.Time
	WorkflowID   string
	Filter       string

	// enforce foreign key
	Connection *Connection
}
