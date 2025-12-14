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
	cmd    *exec.Cmd
	stdin  io.WriteCloser
	stdout io.ReadCloser
	stderr io.ReadCloser

	outR *bufio.Reader
	errR *bufio.Reader
}

func NewSandbox(ctx context.Context, nsjailPath, nsjailCfgPath, submissionDir, jailSubmissionDir string) (*Sandbox, error) {
	cmd := exec.CommandContext(
		ctx,
		nsjailPath,
		"-C", nsjailCfgPath,
		"--bindmount_ro", fmt.Sprintf("%s:%s", submissionDir, jailSubmissionDir),
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
		cmd:    cmd,
		stdin:  stdin,
		stdout: stdout,
		stderr: stderr,
		outR:   bufio.NewReader(stdout),
		errR:   bufio.NewReader(stderr),
	}

	return s, nil
}

func (s *Sandbox) Start() error {
	return s.cmd.Start()
}

func (s *Sandbox) Send(inp any) error {
	b, err := json.Marshal(inp)
	if err != nil {
		return err
	}
	b = append(b, '\n')

	_, err = s.stdin.Write(b)

	return err
}

func (s *Sandbox) RecvOutput(ctx context.Context, v any) error {
	line, err := readLine(ctx, s.outR)
	if err != nil {
		return err
	}

	err = json.Unmarshal(line, v)
	if err != nil {
		return fmt.Errorf("unmarshal output: %v got %s", err, string(line))
	}

	return nil
}

func (s *Sandbox) RecvError(ctx context.Context) ([]byte, error) {
	return readLine(ctx, s.errR)
}

func (s *Sandbox) Destroy() error {
	if s.cmd.Process != nil {
		_ = s.cmd.Process.Kill()
		_ = s.cmd.Wait()
	}

	_ = s.stdin.Close()
	_ = s.stdout.Close()
	_ = s.stderr.Close()

	return nil
}

func readLine(ctx context.Context, r *bufio.Reader) ([]byte, error) {
	type result struct {
		line []byte
		err  error
	}

	ch := make(chan result, 1)

	go func() {
		line, err := r.ReadBytes('\n')
		ch <- result{line, err}
	}()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case res := <-ch:
		if res.err != nil {
			return nil, res.err
		}
		return res.line, nil
	}
}
