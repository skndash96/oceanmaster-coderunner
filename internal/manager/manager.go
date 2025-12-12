package manager

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"strings"
	"sync"

	"github.com/delta/code-runner/internal/config"
)

type GameManager struct {
	cfg     *config.Config
	matches map[string]*match
	mu      sync.Mutex
}

func NewGameManager(cfg *config.Config) *GameManager {
	return &GameManager{
		cfg:     cfg,
		matches: make(map[string]*match),
		mu:      sync.Mutex{},
	}
}

func (gm *GameManager) NewMatch(ID, p1, p2, p1Code, p2Code string) error {
	p1Dir := path.Join(gm.cfg.HostSubmissionPath, ID, "p1")
	p2Dir := path.Join(gm.cfg.HostSubmissionPath, ID, "p2")

	if err := os.MkdirAll(p1Dir, 0700); err != nil {
		return fmt.Errorf("mkdir p1: %w", err)
	}
	if err := os.MkdirAll(p2Dir, 0700); err != nil {
		return fmt.Errorf("mkdir p2: %w", err)
	}

	if err := savePlayerCode(p1Code, path.Join(p1Dir, "submission.py")); err != nil {
		return fmt.Errorf("save p1 code: %w", err)
	}
	if err := savePlayerCode(p2Code, path.Join(p2Dir, "submission.py")); err != nil {
		return fmt.Errorf("save p2 code: %w", err)
	}

	m := &match{
		ID:    ID,
		p1:    p1,
		p2:    p2,
		p1Dir: p1Dir,
		p2Dir: p2Dir,
	}

	gm.mu.Lock()
	gm.matches[ID] = m
	gm.mu.Unlock()

	err := m.Start(gm.cfg)

	gm.mu.Lock()
	delete(gm.matches, ID)
	gm.mu.Unlock()

	_ = os.RemoveAll(m.p1Dir)
	_ = os.RemoveAll(m.p2Dir)

	return err
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
