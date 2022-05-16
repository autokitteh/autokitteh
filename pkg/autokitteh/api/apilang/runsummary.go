package apilang

import (
	"time"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"

	pblang "github.com/autokitteh/autokitteh/gen/proto/stubs/go/lang"
	"github.com/autokitteh/autokitteh/pkg/autokitteh/api/apiprogram"
)

type RunStateLogRecord struct{ pb *pblang.RunStateLogRecord }

func (r *RunStateLogRecord) PB() *pblang.RunStateLogRecord {
	if r == nil || r.pb == nil {
		return nil
	}

	return proto.Clone(r.pb).(*pblang.RunStateLogRecord)
}

func (r *RunStateLogRecord) Clone() *RunStateLogRecord { return &RunStateLogRecord{pb: r.PB()} }

func (r *RunStateLogRecord) WithSource(path *apiprogram.Path) *RunStateLogRecord {
	r = r.Clone()
	r.pb.Source = path.PB()
	return r
}

func MustRunStateLogRecordFromProto(pb *pblang.RunStateLogRecord) *RunStateLogRecord {
	r, err := RunStateLogRecordFromProto(pb)
	if err != nil {
		panic(err)
	}
	return r
}

func RunStateLogRecordFromProto(pb *pblang.RunStateLogRecord) (*RunStateLogRecord, error) {
	if err := pb.Validate(); err != nil {
		return nil, err
	}

	// TODO: more validation?

	return (&RunStateLogRecord{pb: pb}).Clone(), nil
}

func NewRunStateLogRecord(s *RunState, t *time.Time) *RunStateLogRecord {
	if t == nil {
		tt := time.Now()
		t = &tt
	}

	return &RunStateLogRecord{
		pb: &pblang.RunStateLogRecord{
			T:     timestamppb.New(*t),
			State: s.PB(),
		},
	}
}

//--

type RunSummary struct{ pb *pblang.RunSummary }

func (s *RunSummary) PB() *pblang.RunSummary {
	if s == nil || s.pb == nil {
		return nil
	}

	return proto.Clone(s.pb).(*pblang.RunSummary)
}

func (s *RunSummary) Clone() *RunSummary { return &RunSummary{pb: s.PB()} }

func (s *RunSummary) Add(log *RunStateLogRecord) {
	s.pb.Log = append(s.pb.Log, log.PB())

	if p := log.pb.State.GetPrint(); p != nil {
		s.pb.Prints = append(s.pb.Prints, p.Text)
	}
}

func (s *RunSummary) Flatten() (log []*RunStateLogRecord, prints []string) {
	return s.flatten(nil)
}

func (s *RunSummary) flatten(source *apiprogram.Path) (log []*RunStateLogRecord, prints []string) {
	if s == nil || s.pb == nil {
		return nil, nil
	}

	prints = s.pb.Prints

	var loadSource *apiprogram.Path

	for _, pbl := range s.pb.Log {
		if loadSource != nil {
			if lret := pbl.State.GetLoadret(); lret != nil {
				flog, fprints := MustRunSummaryFromProto(lret.RunSummary).flatten(loadSource)

				prints = append(prints, fprints...)
				log = append(log, flog...)
			}

			loadSource = nil
		}

		if lcall := pbl.State.GetLoad(); lcall != nil {
			loadSource = apiprogram.MustPathFromProto(lcall.Path)
		}

		log = append(log, MustRunStateLogRecordFromProto(pbl).WithSource(source))
	}

	return
}

func MustRunSummaryFromProto(pb *pblang.RunSummary) *RunSummary {
	s, err := RunSummaryFromProto(pb)
	if err != nil {
		panic(err)
	}
	return s
}

func RunSummaryFromProto(pb *pblang.RunSummary) (*RunSummary, error) {
	if pb == nil {
		return nil, nil
	}

	if err := pb.Validate(); err != nil {
		return nil, err
	}

	// TODO: more validation?

	return (&RunSummary{pb: pb}).Clone(), nil
}

func NewRunSummary(log []*RunStateLogRecord, prints []string) *RunSummary {
	pblog := make([]*pblang.RunStateLogRecord, len(log))
	for i, r := range log {
		pblog[i] = r.PB()
	}

	return &RunSummary{
		pb: &pblang.RunSummary{
			Prints: prints,
			Log:    pblog,
		},
	}
}
