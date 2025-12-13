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
	gl         *GameLogger
}

func NewMatch(id, p1, p2, p1Dir, p2Dir string, gl *GameLogger) *Match {
	return &Match{
		ID:         id,
		Player1:    p1,
		Player2:    p2,
		Player1Dir: p1Dir,
		Player2Dir: p2Dir,
		gl:         gl,
	}
}

func (m *Match) Start(cfg *config.Config) error {
	m.gl.Log(GameLogDebug, "Starting match")

	matchCtx, cancelCtx := context.WithCancel(context.Background())
	defer cancelCtx()

	s1, err := sandbox.NewSandbox(matchCtx, cfg.NsjailPath, cfg.NsjailCfgPath, m.Player1Dir, cfg.JailSubmissionPath)
	if err != nil {
		return fmt.Errorf("p1 failed to start: %w", err)
	}
	defer s1.Destroy()

	s2, err := sandbox.NewSandbox(matchCtx, cfg.NsjailPath, cfg.NsjailCfgPath, m.Player2Dir, cfg.JailSubmissionPath)
	if err != nil {
		return fmt.Errorf("p2 failed to start: %w", err)
	}
	defer s2.Destroy()

	go streamErrors(matchCtx, s1, m.gl, "p1")
	go streamErrors(matchCtx, s2, m.gl, "p2")

	if err := s1.Start(); err != nil {
		return fmt.Errorf("p1 failed to start: %w", err)
	}
	if err := s2.Start(); err != nil {
		return fmt.Errorf("p2 failed to start: %w", err)
	}

	m.gl.Log(GameLogDebug, "Algorithms started")

	var (
		gameState GameState = NewGameState()
		isP1Turn            = true
	)

	m.gl.Log(GameLogGameState, gameState)

	for {
		turnCtx, cancelTurn := context.WithTimeout(matchCtx, 6*time.Second)

		actions := []Action{}
		var turnErr error

		trStart := time.Now()
		if isP1Turn {
			turnErr = doTurn(turnCtx, s1, m.gl, "p1", &gameState, &actions)
			m.gl.Log(GameLogDebug, fmt.Sprintf("Completed Turn (elapsed %s)", time.Since(trStart)))
		} else {
			turnErr = doTurn(turnCtx, s2, m.gl, "p2", &gameState, &actions)
			m.gl.Log(GameLogDebug, fmt.Sprintf("Completed Turn (elapsed %s)", time.Since(trStart)))
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

		m.gl.Log(GameLogGameAction, actions)

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

func doTurn(turnCtx context.Context, s *sandbox.Sandbox, gl *GameLogger, label string, state *GameState, out *[]Action) error {
	gl.Log(GameLogDebug, label, "Sending state")

	if err := s.Send(state); err != nil {
		return fmt.Errorf("send error: %w", err)
	}

	gl.Log(GameLogDebug, label, "Waiting for output")

	if err := s.RecvOutput(turnCtx, out); err != nil {
		return fmt.Errorf("receive error: %w", err)
	}

	return nil
}

func streamErrors(ctx context.Context, s *sandbox.Sandbox, gl *GameLogger, label string) {
	for {
		data, err := s.RecvError(ctx)
		if err != nil {
			return
		}
		if strings.TrimSpace(string(data)) != "" {
			gl.Log(GameLogError, label, string(data))
		}
	}
}
