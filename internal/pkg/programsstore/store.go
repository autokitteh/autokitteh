package programsstore

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/autokitteh/stores/pkvstore"

	"go.autokitteh.dev/sdk/api/apiprogram"
	"go.autokitteh.dev/sdk/api/apiproject"
)

var ErrNotFound = pkvstore.ErrNotFound

type Store struct{ Store pkvstore.Store }

type File struct {
	Path           *apiprogram.Path   `json:"path"`
	FetchedVersion string             `json:"fetched_version"` // != path.version if path.version is empty.
	FetchedAt      time.Time          `json:"fetched_at"`
	Source         []byte             `json:"source"`
	Module         *apiprogram.Module `json:"module"` // compiled
}

// TODO: make the batch atomic.
func (s *Store) Update(
	ctx context.Context,
	pid apiproject.ProjectID,
	fs []*File,
) error {
	for _, f := range fs {
		data, err := json.Marshal(f)
		if err != nil {
			return fmt.Errorf("marshal: %w", err)
		}

		// TODO: this will collide when path.version is empty, when multiple invocations
		//       on the same project occur concurrently.
		if err := s.Store.Put(ctx, pid.String(), f.Path.String(), data); err != nil {
			return fmt.Errorf("put %q: %w", f.Path.String(), err)
		}

		if fver := f.FetchedVersion; f.Path.Version() == "" && fver != "" {
			// Store also for specific version.

			path := f.Path.WithVersion(fver)

			if err := s.Store.Put(ctx, pid.String(), path.String(), data); err != nil {
				return fmt.Errorf("put %q: %w", f.Path.String(), err)
			}
		}
	}

	return nil
}

func (s *Store) Get(
	ctx context.Context,
	pid apiproject.ProjectID,
	paths []*apiprogram.Path,
	noContent bool,
) (fs []*File, err error) {
	fs = make([]*File, 0, len(paths))

	ks := make([]string, len(paths))
	for i, p := range paths {
		ks[i] = p.String()
	}

	if paths == nil {
		if ks, err = s.Store.List(ctx, pid.String()); err != nil {
			return nil, fmt.Errorf("list: %w", err)
		}
	}

	for _, k := range ks {
		v, err := s.Store.Get(ctx, pid.String(), k)
		if err != nil {
			if errors.Is(err, pkvstore.ErrNotFound) {
				continue
			}

			return nil, fmt.Errorf("get %q: %w", k, err)
		}

		f := new(File)

		if err := json.Unmarshal(v, &f); err != nil {
			return nil, fmt.Errorf("unmarshal %q: %w", k, err)
		}

		fs = append(fs, f)
	}

	return
}

func (s *Store) Setup(ctx context.Context) error { return s.Store.Setup(ctx) }

func (s *Store) Teardown(ctx context.Context) error { return s.Store.Teardown(ctx) }
