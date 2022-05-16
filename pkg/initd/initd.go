package initd

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	L "github.com/autokitteh/autokitteh/pkg/l"
)

type Config struct {
	Dir string            `envconfig:"DIR" json:"dir"`
	Env map[string]string `envconfig:"ENV" json:"env"`
}

type Initd struct {
	Config Config
	Env    map[string]string
	L      L.Nullable
}

func (i *Initd) Start() error {
	if i.Config.Dir == "" {
		return nil
	}

	env := os.Environ()
	for k, v := range i.Config.Env {
		env = append(env, fmt.Sprintf("%s=%s", k, v))
	}
	for k, v := range i.Env {
		env = append(env, fmt.Sprintf("%s=%s", k, v))
	}

	fis, err := ioutil.ReadDir(i.Config.Dir)
	if err != nil {
		return fmt.Errorf("readdir: %w", err)
	}

	for _, fi := range fis {
		mode := fi.Mode()

		if !mode.IsRegular() || mode&0111 == 0 {
			continue
		}

		if err := i.exec(filepath.Join(i.Config.Dir, fi.Name()), env); err != nil {
			return fmt.Errorf("%q: %w", fi.Name(), err)
		}
	}

	return nil
}

func (i *Initd) exec(path string, env []string) error {
	l := i.L.With("path", path)

	cmd := &exec.Cmd{Path: path, Env: env}

	l.Debug("starting")

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start: %w", err)
	}

	l.Info("started", "pid", cmd.Process.Pid)

	go i.monitor(l, cmd)

	return nil
}

func (i *Initd) monitor(l L.L, cmd *exec.Cmd) {
	if err := cmd.Wait(); err != nil {
		l.Error("process exited", "err", err)
		return
	}

	// TODO: restart process with some policy?
}
