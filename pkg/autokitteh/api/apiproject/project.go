package apiproject

import (
	"time"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"

	pbproject "github.com/autokitteh/autokitteh/api/gen/stubs/go/project"

	"github.com/autokitteh/autokitteh/pkg/autokitteh/api/apiaccount"
)

type ProjectPB = pbproject.Project

type Project struct{ pb *pbproject.Project }

func (p *Project) PB() *pbproject.Project {
	if p == nil || p.pb == nil {
		return nil
	}

	return proto.Clone(p.pb).(*pbproject.Project)
}

func (p *Project) Clone() *Project { return &Project{pb: p.PB()} }

func (p *Project) ID() ProjectID { return ProjectID(p.pb.Id) }
func (p *Project) AccountName() apiaccount.AccountName {
	return apiaccount.AccountName(p.pb.AccountName)
}
func (p *Project) Settings() *ProjectSettings { return MustProjectSettingsFromProto(p.pb.Settings) }

func ProjectFromProto(pb *pbproject.Project) (*Project, error) {
	if pb == nil {
		return nil, nil
	}

	if err := pb.Validate(); err != nil {
		return nil, err
	}

	// TODO: more validation?
	return (&Project{pb: pb}).Clone(), nil
}

func NewProject(id ProjectID, aname apiaccount.AccountName, d *ProjectSettings, createdAt time.Time, updatedAt *time.Time) (*Project, error) {
	var pbupdatedat *timestamppb.Timestamp
	if updatedAt != nil {
		pbupdatedat = timestamppb.New(*updatedAt)
	}

	return ProjectFromProto(
		&pbproject.Project{
			Id:          id.String(),
			AccountName: aname.String(),
			Settings:    d.PB(),
			CreatedAt:   timestamppb.New(createdAt),
			UpdatedAt:   pbupdatedat,
		},
	)
}
