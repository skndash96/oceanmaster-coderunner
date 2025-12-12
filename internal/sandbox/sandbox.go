package sandbox

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
)

type Sandbox struct {
	ctx           context.Context
	submissionDir string
	cmd           *exec.Cmd
	stdin         io.WriteCloser
	stdout        io.ReadCloser
	stderr        io.ReadCloser
	outScanner    *bufio.Reader
	errScanner    *bufio.Reader
}

func NewSandbox(ctx context.Context, nsjailPath, nsjailCfgPath, submissionDir, jailSubmissionDir string) (*Sandbox, error) {
	cmd := exec.CommandContext(
		ctx,
		nsjailPath,
		"-C",
		nsjailCfgPath,
		"--bindmount_ro",
		fmt.Sprintf("%s:%s", submissionDir, jailSubmissionDir),
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
		submissionDir: submissionDir,
		cmd:           cmd,
		stdin:         stdin,
		stdout:        stdout,
		stderr:        stderr,
		outScanner:    bufio.NewReader(stdout),
		errScanner:    bufio.NewReader(stderr),
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
	json, err := json.Marshal(inp)
	if err != nil {
		return err
	}

	json = append(json, '\n')

	_, err = s.stdin.Write(json)
	if err != nil {
		return err
	}

	return nil
}

func (s *Sandbox) RecvOutput(v any) error {
	bytes, err := scan(s.outScanner)
	if err != nil {
		return err
	}

	err = json.Unmarshal(bytes, v)
	if err != nil {
		return err
	}

	return nil
}

func (s *Sandbox) RecvError() ([]byte, error) {
	return scan(s.errScanner)
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

	if s.cmd.Process != nil {
		if err := s.cmd.Process.Kill(); err != nil {
			return err
		}
	}

	return nil
}

func scan(b *bufio.Reader) ([]byte, error) {
	bytes, err := b.ReadBytes('\n')
	if err != nil {
		return nil, err
	}
	return bytes, err
}
