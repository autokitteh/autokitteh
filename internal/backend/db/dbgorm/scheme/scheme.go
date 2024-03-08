package scheme

import (
	"encoding/json"
	"fmt"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"
	"gorm.io/datatypes"
	"gorm.io/gorm"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	deploymentsv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/deployments/v1"
	eventsv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/events/v1"
	integrationsv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/integrations/v1"
	sessionsv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/sessions/v1"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

// TODO(ENG-192): use proper foreign keys and normalize model.

// TODO: keep some log of actions performed. Something that
// can be used for recovery from unintended/malicious actions.

var Tables = []any{
	&Build{},
	&Connection{},
	&Deployment{},
	&Env{},
	&EnvVar{},
	&Event{},
	&EventRecord{},
	&Integration{},
	&Project{},
	&Session{},
	&SessionCallAttempt{},
	&SessionCallSpec{},
	&SessionLogRecord{},
	&Signal{},
	&Trigger{},
}

type Build struct {
	BuildID   string `gorm:"primaryKey"`
	Data      []byte
	CreatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

func ParseBuild(b Build) (sdktypes.Build, error) {
	build, err := sdktypes.StrictBuildFromProto(&sdktypes.BuildPB{
		BuildId:   b.BuildID,
		CreatedAt: timestamppb.New(b.CreatedAt),
	})
	if err != nil {
		return sdktypes.InvalidBuild, fmt.Errorf("invalid record: %w", err)
	}

	return build, nil
}

type Connection struct {
	ConnectionID  string `gorm:"primaryKey"`
	IntegrationID string
	// TODO(ENG-111): Integration Integration `gorm:"foreignKey:IntegrationID"`
	// TODO(ENG-111): Also call "Preload()" where relevant
	IntegrationToken string
	ProjectID        string
	Name             string
}

func ParseConnection(c Connection) (sdktypes.Connection, error) {
	conn, err := sdktypes.StrictConnectionFromProto(&sdktypes.ConnectionPB{
		ConnectionId:     c.ConnectionID,
		IntegrationId:    c.IntegrationID,
		IntegrationToken: c.IntegrationToken,
		ProjectId:        c.ProjectID,
		Name:             c.Name,
	})
	if err != nil {
		return sdktypes.InvalidConnection, fmt.Errorf("invalid connection record: %w", err)
	}
	return conn, nil
}

type Integration struct {
	// Unique internal identifier.
	IntegrationID string `gorm:"primaryKey"`

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
		IntegrationId: i.IntegrationID,
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
	ProjectID string `gorm:"primaryKey"`
	Name      string `gorm:"uniqueIndex"`
	RootURL   string
	Resources []byte
}

func ParseProject(r Project) (sdktypes.Project, error) {
	p, err := sdktypes.StrictProjectFromProto(&sdktypes.ProjectPB{
		ProjectId: r.ProjectID,
		Name:      r.Name,
	})
	if err != nil {
		return sdktypes.InvalidProject, fmt.Errorf("invalid project record: %w", err)
	}

	return p, nil
}

type Event struct {
	EventID          string `gorm:"uniqueIndex"`
	IntegrationID    string `gorm:"index"`
	IntegrationToken string `gorm:"index"`
	OriginalEventID  string
	EventType        string `gorm:"index:idx_connection_id_event_type_seq,priority:2;index:idx_event_type"`
	ConnectionID     string `gorm:"index:idx_connection_id_event_type_seq,priority:1"`
	Data             datatypes.JSON
	Memo             datatypes.JSON
	CreatedAt        time.Time
	Seq              uint64 `gorm:"primaryKey;autoIncrement:true,index:idx_connection_id_event_type_seq,priority:3"`
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
		EventId:          e.EventID,
		IntegrationId:    e.IntegrationID,
		IntegrationToken: e.IntegrationToken,
		OriginalEventId:  e.OriginalEventID,
		EventType:        e.EventType,
		Data:             kittehs.TransformMapValues(data, sdktypes.ToProto),
		Memo:             memo,
		CreatedAt:        timestamppb.New(e.CreatedAt),
		Seq:              e.Seq,
	})
}

type EventRecord struct {
	Seq     uint32 `gorm:"primaryKey"`
	EventID string `gorm:"primaryKey"`
	// Event     Event
	State     int32 `gorm:"index"`
	CreatedAt time.Time
}

func ParseEventRecord(e EventRecord) (sdktypes.EventRecord, error) {
	return sdktypes.StrictEventRecordFromProto(&sdktypes.EventRecordPB{
		Seq:       e.Seq,
		EventId:   e.EventID,
		State:     eventsv1.EventState(e.State),
		CreatedAt: timestamppb.New(e.CreatedAt),
	})
}

type Env struct {
	EnvID     string `gorm:"primaryKey"`
	ProjectID string `gorm:"index;foreignKey"`
	Name      string
	DeletedAt gorm.DeletedAt

	// {pid.uuid}/{name}. easier to detect dups.
	// See OrgMember for more.
	MembershipID string `gorm:"uniqueIndex"`

	// just for the foreign key. Wihtout it gorm won't enforce it
	Project Project
}

// enforce foreign keys constrains while soft-deleting
func (e *Env) BeforeDelete(db *gorm.DB) error {
	var count int64
	db.Model(&Deployment{}).Where("deleted_at is NULL and env_id = ?", e.EnvID).Count(&count)
	if count > 0 {
		return fmt.Errorf("FOREIGN KEY: %w", gorm.ErrForeignKeyViolated)
	}
	return nil
}

func ParseEnv(r Env) (sdktypes.Env, error) {
	return sdktypes.StrictEnvFromProto(&sdktypes.EnvPB{
		EnvId:     r.EnvID,
		ProjectId: r.ProjectID,
		Name:      r.Name,
	})
}

type EnvVar struct {
	EnvID string `gorm:"index"`
	Name  string
	Value string // not set if is_secret.

	// Set only if is_secret. will not be fetched by get, only by reveal.
	SecretValue string // TODO: encrypt?
	IsSecret    bool

	// {eid.uuid}/{name}. easier to detect dups.
	// See OrgMember for more.
	MembershipID string `gorm:"uniqueIndex"`
}

func ParseEnvVar(r EnvVar) (sdktypes.EnvVar, error) {
	v := r.Value

	if r.IsSecret {
		v = r.SecretValue
	}

	return sdktypes.StrictEnvVarFromProto(&sdktypes.EnvVarPB{
		EnvId:    r.EnvID,
		Name:     r.Name,
		Value:    v,
		IsSecret: r.IsSecret,
	})
}

type Trigger struct {
	TriggerID string `gorm:"primaryKey"`

	ProjectID    string `gorm:"index"`
	EnvID        string `gorm:"index"`
	ConnectionID string `gorm:"index"`
	Connection   Connection
	EventType    string
	CodeLocation string
}

func ParseTrigger(e Trigger) (sdktypes.Trigger, error) {
	loc, err := sdktypes.ParseCodeLocation(e.CodeLocation)
	if err != nil {
		return sdktypes.InvalidTrigger, fmt.Errorf("loc: %w", err)
	}

	return sdktypes.StrictTriggerFromProto(&sdktypes.TriggerPB{
		TriggerId:    e.TriggerID,
		EnvId:        e.EnvID,
		ConnectionId: e.ConnectionID,
		EventType:    e.EventType,
		CodeLocation: loc.ToProto(),
	})
}

type SessionLogRecord struct {
	SessionID string `gorm:"index"`
	Data      datatypes.JSON
}

func ParseSessionLogRecord(c SessionLogRecord) (spec sdktypes.SessionLogRecord, err error) {
	err = json.Unmarshal(c.Data, &spec)
	return
}

type SessionCallSpec struct {
	SessionID string `gorm:"primaryKey"`
	Seq       uint32 `gorm:"primaryKey"`
	Data      datatypes.JSON
}

func ParseSessionCallSpec(c SessionCallSpec) (spec sdktypes.SessionCallSpec, err error) {
	err = json.Unmarshal(c.Data, &spec)
	return
}

type SessionCallAttempt struct {
	SessionID string `gorm:"uniqueIndex:idx_session_id_seq_attempt,priority:1"`
	Seq       uint32 `gorm:"uniqueIndex:idx_session_id_seq_attempt,priority:2"`
	Attempt   uint32 `gorm:"uniqueIndex:idx_session_id_seq_attempt,priority:3"`
	Start     datatypes.JSON
	Complete  datatypes.JSON
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
	SessionID        string `gorm:"primaryKey"`
	DeploymentID     string `gorm:"index;foreignKey"`
	EventID          string `gorm:"index"`
	CurrentStateType int    `gorm:"index"`
	Entrypoint       string
	Inputs           datatypes.JSON
	CreatedAt        time.Time      //`gorm:"default:current_timestamp"`
	UpdatedAt        time.Time      //`gorm:"default:current_timestamp"`
	DeletedAt        gorm.DeletedAt `gorm:"index"`

	// just for foreign key constraint. Without it gorm won't enforce it
	Deployment *Deployment
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
		SessionId:    s.SessionID,
		DeploymentId: s.DeploymentID,
		EventId:      s.EventID,
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
	DeploymentID string `gorm:"primaryKey"`
	EnvID        string `gorm:"foreignKey"`
	BuildID      string `gorm:"foreignKey"`
	State        int32
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    gorm.DeletedAt `gorm:"index"`

	// just for foreign key constraint. Without it gorm won't enforce it
	Env   Env
	Build Build
}

func ParseDeployment(d Deployment) (sdktypes.Deployment, error) {
	deployment, err := sdktypes.StrictDeploymentFromProto(&sdktypes.DeploymentPB{
		DeploymentId: d.DeploymentID,
		BuildId:      d.BuildID,
		EnvId:        d.EnvID,
		State:        deploymentsv1.DeploymentState(d.State),
		CreatedAt:    timestamppb.New(d.CreatedAt),
		UpdatedAt:    timestamppb.New(d.UpdatedAt),
	})
	if err != nil {
		return sdktypes.InvalidDeployment, fmt.Errorf("invalid record: %w", err)
	}

	return deployment, nil
}

// gorm don't cascade soft deletes. hook is transactional.
func (d *Deployment) AfterDelete(db *gorm.DB) error {
	return db.Where("deployment_id = ?", d.DeploymentID).Delete(&Session{}).Error
}

type DeploymentWithStats struct {
	Deployment
	Created   uint32
	Running   uint32
	Error     uint32
	Completed uint32
}

func ParseDeploymentWithSessionStats(d DeploymentWithStats) (sdktypes.Deployment, error) {
	deployment, err := sdktypes.StrictDeploymentFromProto(&sdktypes.DeploymentPB{
		DeploymentId: d.DeploymentID,
		BuildId:      d.BuildID,
		EnvId:        d.EnvID,
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
		},
	})
	if err != nil {
		return sdktypes.InvalidDeployment, fmt.Errorf("invalid record: %w", err)
	}

	return deployment, nil
}

type Signal struct {
	SignalID     string `gorm:"primaryKey"`
	ConnectionID string `gorm:"index:idx_connection_id_event_type"`
	Connection   Connection
	CreatedAt    time.Time
	WorkflowID   string
	EventType    string `gorm:"index:idx_connection_id_event_type"`
}
