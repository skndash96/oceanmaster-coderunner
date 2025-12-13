package engine

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/delta/code-runner/internal/config"
	"github.com/delta/code-runner/internal/sandbox"
)

type Match struct {
	ID         string
	Player1    string
	Player2    string
	Player1Dir string
	Player2Dir string
}

func NewMatch(id, p1, p2, p1Dir, p2Dir string) *Match {
	return &Match{
		ID:         id,
		Player1:    p1,
		Player2:    p2,
		Player1Dir: p1Dir,
		Player2Dir: p2Dir,
	}
}

func (m *Match) Start(cfg *config.Config) error {
	matchCtx, cancelCtx := context.WithCancel(context.Background())
	defer cancelCtx()

	s1, err := sandbox.NewSandbox(matchCtx, cfg.NsjailPath, cfg.NsjailCfgPath, m.Player1Dir, cfg.JailSubmissionPath)
	if err != nil {
		return fmt.Errorf("s1 failed to start: %w", err)
	}
	defer s1.Destroy()

	s2, err := sandbox.NewSandbox(matchCtx, cfg.NsjailPath, cfg.NsjailCfgPath, m.Player2Dir, cfg.JailSubmissionPath)
	if err != nil {
		return fmt.Errorf("s2 failed to start: %w", err)
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
		gameState GameState = NewGameState()
		isP1Turn            = true
	)

	for {
		turnCtx, cancelTurn := context.WithTimeout(matchCtx, time.Duration(cfg.JailTickTimeLimit))
		actions := []Action{}
		var turnErr error

		if isP1Turn {
			turnErr = doTurn(turnCtx, s1, &gameState, &actions)
		} else {
			turnErr = doTurn(turnCtx, s2, &gameState, &actions)
		}

		cancelTurn()

		if turnErr != nil {
			// should i end the match as failed?
			// or skip turn and wait till N consecutive turn errors to end the match?
			if isP1Turn {
				return fmt.Errorf("p1 algo turn failed: %w", turnErr)
			} else {
				return fmt.Errorf("p2 algo turn failed: %w", turnErr)
			}
		}

		// argpass by value
		newState, ok, hasEnded := UpdateGameState(gameState, actions)
		if ok {
			gameState = newState
		}

		isP1Turn = !isP1Turn

		if hasEnded {
			break
		}
	}

	return nil
}

func doTurn(turnCtx context.Context, s *sandbox.Sandbox, state *GameState, out *[]Action) error {
	if err := s.Send(state); err != nil {
		return fmt.Errorf("send error: %w", err)
	}

	if err := s.RecvOutput(turnCtx, out); err != nil {
		return fmt.Errorf("receive error: %w", err)
	}

	return nil
}

func streamErrors(ctx context.Context, label string, s *sandbox.Sandbox) {
	for {
		data, err := s.RecvError(ctx)
		if err != nil {
			return
		}
		if strings.TrimSpace(string(data)) != "" {
			fmt.Printf("[ERROR] %s: %s\n", label, string(data))
		}
	}
}
