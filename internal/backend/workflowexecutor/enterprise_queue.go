//go:build enterprise
// +build enterprise

package workflowexecutor

import (
	"sync"

	"go.autokitteh.dev/autokitteh/internal/backend/db"
	"go.autokitteh.dev/autokitteh/internal/backend/db/dbgorm/scheme"
)

type q struct {
	requests []scheme.WorkflowExecutionRequest
	lock     *sync.Mutex
	db       db.DB
}

// Need to implement db access
func newQueue(db db.DB) *q {
	return &q{
		requests: make([]scheme.WorkflowExecutionRequest, 0),
		lock:     &sync.Mutex{},
		db:       db,
	}
}

func (q *q) push(request scheme.WorkflowExecutionRequest) {
	q.lock.Lock()
	defer q.lock.Unlock()
	q.requests = append(q.requests, request)
}

func (q *q) popX(n int) []scheme.WorkflowExecutionRequest {
	q.lock.Lock()
	defer q.lock.Unlock()
	if len(q.requests) == 0 {
		return nil
	}
	if len(q.requests) < n {
		n = len(q.requests)
	}
	requests := q.requests[:n]
	q.requests = q.requests[n:]
	return requests
}

func (q *q) len() int {
	q.lock.Lock()
	defer q.lock.Unlock()
	return len(q.requests)
}
