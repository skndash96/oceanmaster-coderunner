package manager

import (
	"io"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/delta/code-runner/internal/config"
)

type GameManager struct {
	cfg     *config.Config
	matches map[string]*match
}

func NewGameManager(cfg *config.Config) *GameManager {
	return &GameManager{
		cfg:     cfg,
		matches: map[string]*match{},
	}
}

func (gm *GameManager) NewMatch(ID, p1, p2, p1Code, p2Code string) error {
	p1Dir := path.Join(gm.cfg.HostSubmissionPath, ID, "p1")
	p2Dir := path.Join(gm.cfg.HostSubmissionPath, ID, "p2")

	if err := os.MkdirAll(p1Dir, 0700); err != nil {
		return err
	}
	defer os.RemoveAll(p1Dir)

	if err := os.MkdirAll(p2Dir, 0700); err != nil {
		return err
	}
	defer os.RemoveAll(p2Dir)

	if err := savePlayerCode(p1Code, path.Join(p1Dir, "submission.py")); err != nil {
		return err
	}

	if err := savePlayerCode(p2Code, path.Join(p2Dir, "submission.py")); err != nil {
		return err
	}

	m := &match{
		ID:    ID,
		p1:    p1,
		p2:    p2,
		p1Dir: p1Dir,
		p2Dir: p2Dir,
	}

	gm.matches[ID] = m
	defer func() {
		delete(gm.matches, ID)
	}()

	err := m.Start(gm.cfg)

	if err != nil {
		return err
	}

	return nil
}

func savePlayerCode(s string, path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	var content io.ReadCloser

	if strings.HasPrefix(s, "http") {
		resp, err := http.Get(s)
		if err != nil {
			return err
		}

		content = resp.Body
		defer content.Close()
	} else {
		content = io.NopCloser(strings.NewReader(s))
	}

	_, err = io.Copy(f, content)
	if err != nil {
		return err
	}

	return nil
}
