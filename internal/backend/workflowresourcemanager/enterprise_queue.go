//go:build enterprise
// +build enterprise

package workflowresourcemanager

import (
	"sync"

	"go.autokitteh.dev/autokitteh/internal/backend/db"
)

type q struct {
	jobs []job
	lock *sync.Mutex
	db   db.DB
}

// Need to implement db access
func newQueue(db db.DB) *q {
	return &q{
		jobs: make([]job, 0),
		lock: &sync.Mutex{},
		db:   db,
	}
}

func (q *q) push(job job) {
	q.lock.Lock()
	defer q.lock.Unlock()
	q.jobs = append(q.jobs, job)
}

func (q *q) popX(n int) []job {
	q.lock.Lock()
	defer q.lock.Unlock()
	if len(q.jobs) == 0 {
		return nil
	}
	if len(q.jobs) < n {
		n = len(q.jobs)
	}
	jobs := q.jobs[:n]
	q.jobs = q.jobs[n:]
	return jobs
}

func (q *q) len() int {
	q.lock.Lock()
	defer q.lock.Unlock()
	return len(q.jobs)
}
