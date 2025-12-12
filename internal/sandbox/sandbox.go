package sandbox

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"slices"

	"github.com/delta/code-runner/internal/config"
)

type Sandbox struct {
	cfg           *config.Config
	submissionDir string
	cmd           *exec.Cmd
	stdin         io.WriteCloser
	stdout        io.ReadCloser
	stderr        io.ReadCloser
	outScanner    *bufio.Scanner
	errScanner    *bufio.Scanner
}

func NewSandbox(cfg *config.Config, submissionDir string) (*Sandbox, error) {
	cmd := exec.Command(
		cfg.NsjailPath,
		"-C",
		cfg.NsjailCfgPath,
		"--bindmount_ro",
		fmt.Sprintf("%s:%s", submissionDir, cfg.Jail.SubmissionPath),
	)

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, err
	}

	s := &Sandbox{
		cfg:           cfg,
		submissionDir: submissionDir,
		cmd:           cmd,
		stdin:         stdin,
		stdout:        stdout,
		stderr:        stderr,
		outScanner:    bufio.NewScanner(stdout),
		errScanner:    bufio.NewScanner(stderr),
	}

	return s, nil
}

func (s *Sandbox) Start() error {
	if err := s.cmd.Start(); err != nil {
		return err
	}

	return nil
}

func (s *Sandbox) Send(inp any) error {
	str, err := json.Marshal(inp)
	if err != nil {
		return err
	}

	str = slices.Concat(str, []byte("\n"))

	_, err = s.stdin.Write(str)
	if err != nil {
		return err
	}

	return nil
}

func (s *Sandbox) RecvOutput(v any) error {
	if s.outScanner.Scan() {
		err := json.Unmarshal(s.outScanner.Bytes(), v)
		if err != nil {
			return err
		}
		return nil
	}

	err := s.outScanner.Err()
	if err != nil {
		return err
	}
	return fmt.Errorf("Reached EOF")
}

func (s *Sandbox) RecvError() []byte {
	if s.errScanner.Scan() {
		return s.errScanner.Bytes()
	}
	return nil
	// err := s.errScanner.Err()
	// if err != nil {
	// 	return nil, err
	// }
	// return nil, fmt.Errorf("Reached EOF")
}

func (s *Sandbox) Destroy() error {
	if err := s.stdin.Close(); err != nil {
		return err
	}

	if err := s.stdout.Close(); err != nil {
		return err
	}

	if err := s.stderr.Close(); err != nil {
		return err
	}

	if err := s.cmd.Wait(); err != nil {
		return err
	}

	return nil
}
