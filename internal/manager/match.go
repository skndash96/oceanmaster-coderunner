package manager

import (
	"context"
	"fmt"
	"slices"

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
	gameCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	s1, err := sandbox.NewSandbox(gameCtx, cfg.NsjailPath, cfg.NsjailCfgPath, m.p1Dir, cfg.JailSubmissionPath)
	if err != nil {
		return err
	}
	defer s1.Destroy()

	s2, err := sandbox.NewSandbox(gameCtx, cfg.NsjailPath, cfg.NsjailCfgPath, m.p2Dir, cfg.JailSubmissionPath)
	if err != nil {
		return err
	}
	defer s2.Destroy()

	gameState := []int{}
	isP1Turn := true

	// --------------
	// after this point, failure is probably due to the user's code
	// or the sandbox environment
	// so log properly and let the user know

	go func() {
		for {
			data, err := s1.RecvError()
			if err != nil {
				break
			}
			fmt.Printf("[ERROR] p1: %v", string(data))
		}
	}()

	go func() {
		for {
			data, err := s2.RecvError()
			if err != nil {
				break
			}
			fmt.Printf("[ERROR] p2: %v", string(data))
		}
	}()

	// its not required to start a goroutine
	// for each sandbox coz we'd be taking turns
	// so only one at a time

	if err := s1.Start(); err != nil {
		return err
	}
	if err := s2.Start(); err != nil {
		return err
	}

	tick := 0

	for {
		tick++

		fmt.Println("Sending state...")

		var actions []int
		if isP1Turn {
			err = s1.Send(gameState)
			if err != nil {
				return err
			}
			fmt.Println("Waiting for p1...")
			err = s1.RecvOutput(&actions)
		} else {
			err = s2.Send(gameState)
			if err != nil {
				return err
			}
			fmt.Println("Waiting for p2...")
			err = s2.RecvOutput(&actions)
		}

		if err != nil {
			return err
		}

		fmt.Println("Received actions:", actions)

		// TODO: Validate actions

		gameState = slices.Concat(gameState, actions)
		isP1Turn = !isP1Turn

		fmt.Println("Tick", tick, "Turn", isP1Turn, len(gameState), gameState[len(gameState)-1])

		if tick > 30 {
			fmt.Println("Game over!")
			break
		}
	}

	return nil
}
