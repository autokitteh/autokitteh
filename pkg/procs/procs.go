package procs

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os/exec"
	"sync"
	"time"

	L "gitlab.com/softkitteh/autokitteh/pkg/l"
)

type Cmd = exec.Cmd

type proc struct{ readyCh chan []byte }

type Config struct {
	ReadyTimeout time.Duration `envconfig:"READY_TIMEOUT" json:"ready_timeout"`
}

func (c Config) readyTimeout() time.Duration {
	if t := c.ReadyTimeout; t != 0 {
		return t
	}

	return time.Second
}

type Procs struct {
	Config Config
	L      L.Nullable

	mu    sync.Mutex
	procs map[string]*proc // pid -> proc
}

func (ps *Procs) Ready(id string, handshake []byte) error {
	l := ps.L.With("id", id)

	l.Debug("received ready")

	ps.mu.Lock()
	defer ps.mu.Unlock()

	p := ps.procs[id]
	if p == nil {
		return L.Error(l, "not found")
	}

	delete(ps.procs, id)

	p.readyCh <- handshake

	return nil
}

func (ps *Procs) HTTPHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")

	defer r.Body.Close()
	rx, err := ioutil.ReadAll(r.Body)
	if err != nil {
		ps.L.Error("body ready error")
		http.Error(w, "unable to read body", http.StatusInternalServerError)
		return
	}

	ps.Ready(id, rx)
}

func (ps *Procs) Start(cmd *Cmd) ([]byte, error) {
	p := &proc{readyCh: make(chan []byte, 1)}

	ps.mu.Lock()
	if ps.procs == nil {
		ps.procs = make(map[string]*proc, 32)
	}

	if err := cmd.Start(); err != nil {
		ps.mu.Unlock()
		return nil, fmt.Errorf("start: %w", err)
	}

	ps.procs[fmt.Sprintf("%d", cmd.Process.Pid)] = p
	ps.mu.Unlock()

	ps.L.Debug("started", "pid", cmd.Process.Pid)

	exitCh := make(chan error, 1)

	go func() { exitCh <- cmd.Wait() }()

	select {
	case <-time.After(ps.Config.readyTimeout()):
		return nil, fmt.Errorf("exec ready did not received in due time")
	case err := <-exitCh:
		return nil, fmt.Errorf("process exited with error: %w", err)
	case handshake := <-p.readyCh:
		return handshake, nil
	}
}
