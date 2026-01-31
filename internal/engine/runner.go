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

func (m *Match) Simulate(cfg *config.Config) error {
	m.gl.Log(GameLogDebug, "Starting sandbox")

	matchCtx, cancelCtx := context.WithTimeout(context.Background(), time.Duration(cfg.JailWallTimeoutMS)*time.Millisecond)
	defer cancelCtx()

	s1, err := sandbox.NewSandbox(matchCtx, cfg.NsjailPath, cfg.NsjailCfgPath, m.Player1Dir, cfg.JailSubmissionPath)
	if err != nil {
		return fmt.Errorf("create p1 sandbox: %w", err)
	}
	defer s1.Destroy()

	s2, err := sandbox.NewSandbox(matchCtx, cfg.NsjailPath, cfg.NsjailCfgPath, m.Player2Dir, cfg.JailSubmissionPath)
	if err != nil {
		return fmt.Errorf("create p2 sandbox: %w", err)
	}
	defer s2.Destroy()

	go streamErrors(matchCtx, s1, m.gl, "p1")
	go streamErrors(matchCtx, s2, m.gl, "p2")

	if err := s1.Start(); err != nil {
		return fmt.Errorf("start p1 sandbox: %w", err)
	}
	if err := s2.Start(); err != nil {
		return fmt.Errorf("start p2 sandbox: %w", err)
	}

	if err := handshakeSandbox(matchCtx, s1, cfg.JailHandshakeTimeoutMS); err != nil {
		return fmt.Errorf("p1 handshake: %w", err)
	}

	if err := handshakeSandbox(matchCtx, s2, cfg.JailHandshakeTimeoutMS); err != nil {
		return fmt.Errorf("p2 handshake: %w", err)
	}

	m.gl.Log(GameLogDebug, "Completed Handshakes")

	var (
		ge = InitGameEngine(m.gl)
		isP1Turn  = true
	)

	m.gl.Log(GameLogGameState, ge)

	for {
		turnCtx, cancelTurn := context.WithTimeout(matchCtx, time.Duration(cfg.JailTickTimeoutMS)*time.Millisecond)

		move := PlayerMoves{}
		var turnErr error

		if isP1Turn {
			turnErr = doTurn(turnCtx, s1, m.gl, "p1", ge.getState(PlayerOne), &move)
			m.gl.Log(GameLogDebug, "Completed Turn")
		} else {
			turnErr = doTurn(turnCtx, s2, m.gl, "p2", ge.getState(PlayerTwo), &move)
			m.gl.Log(GameLogDebug, "Completed Turn")
		}

		cancelTurn()

		if turnErr != nil {
			// should i end the match as failed?
			// or skip turn and wait till N consecutive turn errors to end the match?
			if isP1Turn {
				return fmt.Errorf("p1 turn: %w", turnErr)
			} else {
				return fmt.Errorf("p2 turn: %w", turnErr)
			}
		}

		m.gl.Log(GameLogGameAction, move)

		// argpass by value
		ge.UpdateState(move)

		isP1Turn = !isP1Turn

		if ge.Winner != -1 {
			break
		}
	}

	return nil
}

func handshakeSandbox(mCtx context.Context, s *sandbox.Sandbox, timeoutMS uint32) error {
	ctx, cancel := context.WithTimeout(mCtx, time.Duration(timeoutMS) * time.Millisecond)
	defer cancel()

	data := ""

	err := s.RecvOutput(ctx, &data)
	if err != nil {
		return err
	}

	if strings.TrimSpace(data) != "__READY__" {
		return fmt.Errorf("Invalid handshake")
	}

	return nil
}

func doTurn(turnCtx context.Context, s *sandbox.Sandbox, gl *GameLogger, label string, playerView PlayerView, out *PlayerMoves) error {
	gl.Log(GameLogDebug, label, "Sending state")

	if err := s.Send(playerView); err != nil {
		return fmt.Errorf("send state: %w", err)
	}

	gl.Log(GameLogDebug, label, "Waiting for output")

	if err := s.RecvOutput(turnCtx, out); err != nil {
		return fmt.Errorf("receive actions: %w", err)
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
