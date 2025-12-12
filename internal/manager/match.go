package manager

import (
	"context"
	"fmt"
	"slices"
	"time"

	"github.com/delta/code-runner/internal/config"
	"github.com/delta/code-runner/internal/sandbox"
)

type match struct {
	ID    string
	p1    string
	p2    string
	p1Dir string
	p2Dir string
}

func (m *match) Start(cfg *config.Config) error {
	matchCtx, cancelMatch := context.WithCancel(context.Background())
	defer cancelMatch()

	s1, err := sandbox.NewSandbox(matchCtx, cfg.NsjailPath, cfg.NsjailCfgPath, m.p1Dir, cfg.JailSubmissionPath)
	if err != nil {
		return err
	}
	defer s1.Destroy()

	s2, err := sandbox.NewSandbox(matchCtx, cfg.NsjailPath, cfg.NsjailCfgPath, m.p2Dir, cfg.JailSubmissionPath)
	if err != nil {
		return err
	}
	defer s2.Destroy()

	go streamErrors(matchCtx, "p1", s1)
	go streamErrors(matchCtx, "p2", s2)

	if err := s1.Start(); err != nil {
		return fmt.Errorf("p1 failed to start: %w", err)
	}
	if err := s2.Start(); err != nil {
		return fmt.Errorf("p2 failed to start: %w", err)
	}

	var (
		gameState []int = []int{}
		isP1Turn  = true
		tick      = 0
	)

	for tick < 30 {
		tick++
		fmt.Printf("Tick %d\n", tick)

		turnCtx, cancelTurn := context.WithTimeout(matchCtx, time.Duration(cfg.Jail.TurnTimeout))
		actions := []int{}
		var turnErr error

		if isP1Turn {
			fmt.Println("Sending state to p1")
			turnErr = doTurn(turnCtx, s1, gameState, &actions)
		} else {
			fmt.Println("Sending state to p2")
			turnErr = doTurn(turnCtx, s2, gameState, &actions)
		}

		cancelTurn()

		if turnErr != nil {
			return fmt.Errorf("bot turn failed: %v %w", isP1Turn, turnErr)
		}

		fmt.Println("Actions:", actions)
		gameState = slices.Concat(gameState, actions)
		isP1Turn = !isP1Turn
	}

	fmt.Println("Match complete.")
	return nil
}

func doTurn(ctx context.Context, s *sandbox.Sandbox, state []int, out *[]int) error {
	if err := s.Send(state); err != nil {
		return fmt.Errorf("send error: %w", err)
	}

	type result struct {
		err error
	}

	ch := make(chan result, 1)
	go func() {
		err := s.RecvOutput(out)
		ch <- result{err: err}
	}()

	select {
	case r := <-ch:
		return r.err
	case <-ctx.Done():
		return fmt.Errorf("turn timeout: %w", ctx.Err())
	}
}

func streamErrors(ctx context.Context, label string, s *sandbox.Sandbox) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			data, err := s.RecvError()
			if err != nil {
				return
			}
			fmt.Printf("[ERROR] %s: %s\n", label, string(data))
		}
	}
}
