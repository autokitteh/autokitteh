package akprocs

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/gorilla/mux"

	"github.com/autokitteh/L"
	"github.com/autokitteh/procs"
)

type Config struct {
	procs.ProcsConfig

	ReadyAddress string `envconfig:"READY_ADDRESS" json:"ready_address"`
}

type Procs struct {
	Config Config
	L      L.Nullable

	procs *procs.Procs
}

func (ps *Procs) Register(r *mux.Router) {
	ps.procs = &procs.Procs{Config: ps.Config.ProcsConfig, L: L.N(ps.L.Named("procs"))}

	r.HandleFunc("/ready", ps.procs.HTTPHandler)
}

func (ps *Procs) Start(path string, args []string, env map[string]string) (cmd *procs.Cmd, addr string, err error) {
	if ps.procs == nil {
		panic("must be registered first")
	}

	if path, err = exec.LookPath(path); err != nil {
		return nil, "", fmt.Errorf("LookPath(%q): %w", path, err)
	}

	cmd = &procs.Cmd{
		Path: path,
		Args: append([]string{path}, args...),
		Env: append(
			os.Environ(),
			fmt.Sprintf("AK_PROC_READY_ADDRESS=%s", ps.Config.ReadyAddress),
		),
	}

	for k, v := range env {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
	}

	var rx []byte
	if rx, err = ps.procs.Start(cmd); err != nil {
		return
	}

	addr = string(rx)

	l := ps.L.With("pid", cmd.Process.Pid, "addr", addr)

	l.Debug("ready")

	return
}
