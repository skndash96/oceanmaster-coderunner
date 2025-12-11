package sandbox

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/delta/code-runner/internal/config"
)

// spawns a sandbox
// with proper file mount and cleanup
type Sandbox struct {
	cfg      *config.Config
	codePath string
}

func NewSandbox(cfg *config.Config, codePath string) *Sandbox {
	return &Sandbox{
		cfg:      cfg,
		codePath: codePath,
	}
}

func (s *Sandbox) Start() error {
	cmd := exec.Command(
		s.cfg.NsjailPath,
		"-C",
		s.cfg.NsjailCfgPath,
		"--bindmount_ro",
		fmt.Sprintf("%s:%s", s.codePath, s.cfg.Jail.SubmissionPath),
	)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start sandbox: %w", err)
	}

	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("failed to wait for sandbox: %w", err)
	}

	return nil
}
