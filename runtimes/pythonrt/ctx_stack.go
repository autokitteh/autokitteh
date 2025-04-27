package pythonrt

import (
	"context"
	"sync"
)

type CtxStack struct {
	mu    sync.RWMutex
	items []context.Context
}

func (s *CtxStack) Push(ctx context.Context) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.items = append(s.items, ctx)
}

func (s *CtxStack) Pop() context.Context {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(s.items) == 0 {
		return nil
	}

	ctx := s.items[len(s.items)-1]
	s.items = s.items[:len(s.items)-1]

	return ctx
}

func (s *CtxStack) Top() context.Context {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if len(s.items) == 0 {
		return nil
	}

	return s.items[len(s.items)-1]
}
