package manager

import (
	"fmt"
	"io"
	"net/http"
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

func (gm *GameManager) NewMatch(ID, p1, p2, p1Code, p2Code string) error {
	tStart := time.Now()

	p1Dir := path.Join(gm.cfg.HostSubmissionPath, ID, "p1")
	p2Dir := path.Join(gm.cfg.HostSubmissionPath, ID, "p2")
	logFile := path.Join(gm.cfg.HostSubmissionPath, ID, "log.txt")

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

	if err := savePlayerCode(p1Code, path.Join(p1Dir, "submission.py")); err != nil {
		return fmt.Errorf("save p1 code: %w", err)
	}

	if err := savePlayerCode(p2Code, path.Join(p2Dir, "submission.py")); err != nil {
		return fmt.Errorf("save p2 code: %w", err)
	}

	gl := engine.NewGameLogger(logF)

	m := engine.NewMatch(ID, p1, p2, p1Dir, p2Dir, gl)

	gm.mu.Lock()
	gm.matches[ID] = m
	gm.mu.Unlock()

	defer func() {
		gm.mu.Lock()
		delete(gm.matches, ID)
		gm.mu.Unlock()
	}()

	gl.Log(engine.GameLogDebug, fmt.Sprintf("New match %s (at %s)", ID, tStart.Format(time.RFC3339)))

	gl.Log(engine.GameLogDebug, fmt.Sprintf("Completed Setup (elapsed %s)", time.Since(tStart)))
	tStart = time.Now()

	err = m.Start(gm.cfg)
	if err != nil {
		gl.Log(engine.GameLogError, fmt.Sprintf("Match %s failed: %v", ID, err))
		return err
	}

	gl.Log(engine.GameLogDebug, fmt.Sprintf("Completed Match (elapsed %s)", time.Since(tStart)))

	// game logs is stored in logFile
	if err := UploadFile(logFile); err != nil {
		return fmt.Errorf("upload log file: %w", err)
	}

	return nil
}

func savePlayerCode(s string, dst string) error {
	f, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer f.Close()

	var src io.ReadCloser

	if strings.HasPrefix(s, "http://") || strings.HasPrefix(s, "https://") {
		resp, err := http.Get(s)
		if err != nil {
			return err
		}
		src = resp.Body
	} else {
		src = io.NopCloser(strings.NewReader(s))
	}
	defer src.Close()

	_, err = io.Copy(f, src)
	return err
}
