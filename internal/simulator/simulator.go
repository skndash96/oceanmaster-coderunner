package simulator

import (
	"io"
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

func Simulate(cfg *config.Config, m *Match) {
	p1Dir := path.Join(cfg.HostSubmissionPath, m.matchID, m.player1)
	p2Dir := path.Join(cfg.HostSubmissionPath, m.matchID, m.player2)

	if err := os.MkdirAll(p1Dir, 0777); err != nil {
		panic(err)
	}
	if err := os.MkdirAll(p2Dir, 0777); err != nil {
		panic(err)
	}

	p1File, err := os.Create(path.Join(p1Dir, "submission.py"))
	if err != nil {
		panic(err)
	}
	p2File, err := os.Create(path.Join(p2Dir, "submission.py"))
	if err != nil {
		panic(err)
	}

	p1Code, err := getPlayerCode(m.player1Code)
	if err != nil {
		panic(err)
	}
	p2Code, err := getPlayerCode(m.player2Code)
	if err != nil {
		panic(err)
	}

	_, err = io.Copy(p1File, p1Code)
	if err != nil {
		panic(err)
	}
	_, err = io.Copy(p2File, p2Code)
	if err != nil {
		panic(err)
	}

	p1 := sandbox.NewSandbox(cfg, p1Dir)
	p2 := sandbox.NewSandbox(cfg, p2Dir)

	err = p1.Start()
	if err != nil {
		panic(err)
	}
	err = p2.Start()
	if err != nil {
		panic(err)
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
}

func getPlayerCode(URL string) (io.Reader, error) {
	// TODO: Implement
	return strings.NewReader("print('a')"), nil
}
