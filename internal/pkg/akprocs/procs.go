package akprocs

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/gorilla/mux"

	L "gitlab.com/softkitteh/autokitteh/pkg/l"
	"gitlab.com/softkitteh/autokitteh/pkg/procs"
)

type Config struct {
	procs.Config

	// If not set, can run everywhere.
	RootPath     string `envconfig:"EXEC_ROOT_PATH" default:"." json:"exec_root_path"`
	ReadyAddress string `envconfig:"READY_ADDRESS" json:"ready_address"`
}

type Procs struct {
	Config Config
	L      L.Nullable

	procs *procs.Procs
}

func (ps *Procs) Register(r *mux.Router) {
	ps.procs = &procs.Procs{Config: ps.Config.Config, L: L.N(ps.L.Named("procs"))}

	r.HandleFunc("/ready", ps.procs.HTTPHandler)
}

func (ps *Procs) Start(path string, args []string, env map[string]string) (cmd *procs.Cmd, addr string, err error) {
	if ps.procs == nil {
		panic("must be registered first")
	}

	if ps.Config.RootPath != "" {
		rootPath, err := filepath.Abs(ps.Config.RootPath)
		if err != nil {
			return nil, "", fmt.Errorf("cannot deduce abs path %q", rootPath)
		}

		path, err = filepath.Abs(filepath.Join(rootPath, path))
		if err != nil {
			return nil, "", fmt.Errorf("cannot deduce abs path for %q", path)
		}

		rel, err := filepath.Rel(rootPath, path)
		if err != nil {
			return nil, "", fmt.Errorf("cannot deduce rel path of %q to %q", path, rootPath)
		}

		if rel[0] == '.' || rel[0] == '/' || rel[0] == '\\' {
			return nil, "", fmt.Errorf("%q must be under %q", path, rootPath)
		}
	}

	cmd = &procs.Cmd{
		Path: filepath.Base(path),
		Args: append([]string{filepath.Base(path)}, args...),
		Dir:  filepath.Dir(path),
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
