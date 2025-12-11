package simulator

import (
	"io"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/delta/code-runner/internal/config"
	"github.com/delta/code-runner/internal/sandbox"
)

// simulator receives a match request
// collects required info like user code
// coordinates game engine and sandbox

type Match struct {
	matchID     string
	player1     string
	player1Code string
	player2     string
	player2Code string
}

func NewMatch(matchID, player1, player2 string, player1Code, player2Code string) *Match {
	return &Match{
		matchID:     matchID,
		player1:     player1,
		player1Code: player1Code,
		player2:     player2,
		player2Code: player2Code,
	}
}

func Simulate(cfg *config.Config, m *Match) error {
	p1Dir, err := os.MkdirTemp(cfg.HostSubmissionPath, "p1-*")
	if err != nil {
		return err
	}

	p2Dir, err := os.MkdirTemp(cfg.HostSubmissionPath, "p2-*")
	if err != nil {
		return err
	}

	p1File, err := os.Create(path.Join(p1Dir, "submission.py"))
	if err != nil {
		return err
	}
	p2File, err := os.Create(path.Join(p2Dir, "submission.py"))
	if err != nil {
		return err
	}

	p1Code, err := getPlayerCode(m.player1Code)
	if err != nil {
		return err
	}
	p2Code, err := getPlayerCode(m.player2Code)
	if err != nil {
		return err
	}

	_, err = io.Copy(p1File, p1Code)
	if err != nil {
		return err
	}
	_, err = io.Copy(p2File, p2Code)
	if err != nil {
		return err
	}

	p1 := sandbox.NewSandbox(cfg, p1Dir)
	p2 := sandbox.NewSandbox(cfg, p2Dir)

	err = p1.Start()
	if err != nil {
		return err
	}
	err = p2.Start()
	if err != nil {
		return err
	}

	// TODO: Convert to Goroutine
	// TODO: Wait for goroutine

	// Cleanup
	defer func() {
		if err := p1File.Close(); err != nil {
			panic(err)
		}
		if err := p2File.Close(); err != nil {
			panic(err)
		}

		if err := os.RemoveAll(p1Dir); err != nil {
			panic(err)
		}
		if err := os.RemoveAll(p2Dir); err != nil {
			panic(err)
		}
	}()

	return nil
}

func getPlayerCode(s string) (io.Reader, error) {
	if strings.HasPrefix(s, "http") {
		resp, err := http.Get(s)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()
		return resp.Body, nil
	}

	return strings.NewReader(s), nil
}
