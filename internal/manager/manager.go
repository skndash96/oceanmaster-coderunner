package manager

import (
	"fmt"
	"io"
	"os"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/delta/code-runner/internal/config"
	"github.com/delta/code-runner/internal/engine"
)

type GameManager struct {
	cfg     *config.Config
	matches map[string]*engine.Match
	mu      sync.Mutex
}

func NewGameManager(cfg *config.Config) *GameManager {
	return &GameManager{
		cfg:     cfg,
		matches: make(map[string]*engine.Match),
		mu:      sync.Mutex{},
	}
}

type MatchJob struct {
	ID     string `json:"id"`
	P1     string `json:"p1"`
	P2     string `json:"p2"`
	P1Code string `json:"p1_code"`
	P2Code string `json:"p2_code"`
}

func (gm *GameManager) NewMatch(job MatchJob) error {
	p1Dir := path.Join(gm.cfg.HostSubmissionPath, job.ID, "p1")
	p2Dir := path.Join(gm.cfg.HostSubmissionPath, job.ID, "p2")
	logFile := path.Join(gm.cfg.HostSubmissionPath, job.ID, "log.txt")

	if err := os.MkdirAll(p1Dir, 0700); err != nil {
		return fmt.Errorf("mkdir p1: %w", err)
	}
	defer os.RemoveAll(p1Dir)

	if err := os.MkdirAll(p2Dir, 0700); err != nil {
		return fmt.Errorf("mkdir p2: %w", err)
	}
	defer os.RemoveAll(p2Dir)

	logF, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("create log file: %w", err)
	}
	defer logF.Close()

	gl := engine.NewGameLogger(logF)

	gl.Log(engine.GameLogDebug, fmt.Sprintf("Match ID %s at %s", job.ID, time.Now().Format(time.RFC3339)))

	if err := savePlayerCode(job.P1Code, path.Join(p1Dir, "submission.py")); err != nil {
		err = fmt.Errorf("save p1 code: %w", err)
		gl.Log(engine.GameLogError, err.Error())
		return err
	}

	if err := savePlayerCode(job.P2Code, path.Join(p2Dir, "submission.py")); err != nil {
		err = fmt.Errorf("save p2 code: %w", err)
		gl.Log(engine.GameLogError, err.Error())
		return err
	}

	m := engine.NewMatch(job.ID, job.P1, job.P2, p1Dir, p2Dir, gl)

	gm.mu.Lock()
	gm.matches[job.ID] = m
	gm.mu.Unlock()

	defer func() {
		gm.mu.Lock()
		delete(gm.matches, job.ID)
		gm.mu.Unlock()
	}()

	gl.Log(engine.GameLogDebug, "Completed Setup")

	err = m.Simulate(gm.cfg)
	if err != nil {
		err = fmt.Errorf("simulate: %w", err)
		gl.Log(engine.GameLogError, err.Error())
		return err
	}

	gl.Log(engine.GameLogDebug, "Completed Simulation")

	if err := UploadFile(m.ID, logFile); err != nil {
		err = fmt.Errorf("upload log file: %w", err)
		gl.Log(engine.GameLogError, err.Error())
		return err
	}

	gl.Log(engine.GameLogDebug, "Completed Post-Simulation Operations")

	return nil
}

func savePlayerCode(s string, dst string) error {
	f, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer f.Close()

	var src = io.NopCloser(strings.NewReader(s))

	// if strings.HasPrefix(s, "http://") || strings.HasPrefix(s, "https://") {
	// 	resp, err := http.Get(s)
	// 	if err != nil {
	// 		return err
	// 	}
	// 	src = resp.Body
	// } else {
	// 	src = io.NopCloser(strings.NewReader(s))
	// }

	defer src.Close()

	_, err = io.Copy(f, src)
	return err
}
