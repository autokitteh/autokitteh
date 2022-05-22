package apievent

import (
	"time"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"

	pbevent "github.com/autokitteh/autokitteh/api/gen/stubs/go/event"
)

type ProjectEventStateRecord struct {
	pb *pbevent.ProjectEventStateRecord
}

func (sr *ProjectEventStateRecord) PB() *pbevent.ProjectEventStateRecord {
	return proto.Clone(sr.pb).(*pbevent.ProjectEventStateRecord)
}

func (sr *ProjectEventStateRecord) Clone() *ProjectEventStateRecord {
	return &ProjectEventStateRecord{pb: sr.PB()}
}

func (sr *ProjectEventStateRecord) T() time.Time { return sr.pb.T.AsTime() }
func (sr *ProjectEventStateRecord) State() *ProjectEventState {
	return MustProjectEventStateFromProto(sr.pb.State)
}

func ProjectEventStateRecordFromProto(pb *pbevent.ProjectEventStateRecord) (*ProjectEventStateRecord, error) {
	if err := pb.Validate(); err != nil {
		return nil, err
	}

	// TODO: more validation?
	return (&ProjectEventStateRecord{pb: pb}).Clone(), nil
}

func NewProjectEventStateRecord(s *ProjectEventState, t time.Time) (*ProjectEventStateRecord, error) {
	return ProjectEventStateRecordFromProto(&pbevent.ProjectEventStateRecord{
		State: s.PB(),
		T:     timestamppb.New(t),
	})
}
