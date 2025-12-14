package engine

import (
	"encoding/json"
	"fmt"
	"io"
	"sync"
	"time"
)

type GameLogger struct {
	mu     sync.Mutex
	enc    *json.Encoder
	tStart time.Time
}

type GameLog struct {
	Typ     GameLogType `json:"typ"`
	Msg     any         `json:"msg"`
	Elapsed string      `json:"elapsed,omitempty"`
}

type GameLogType string

const (
	GameLogDebug      GameLogType = "DEBUG"
	GameLogError      GameLogType = "ERROR"
	GameLogGameState  GameLogType = "STATE"
	GameLogGameAction GameLogType = "ACTION"
)

func NewGameLogger(w io.Writer) *GameLogger {
	return &GameLogger{
		enc:    json.NewEncoder(w),
		tStart: time.Now(),
	}
}

func (gl *GameLogger) Log(typ GameLogType, msg ...any) {
	gl.mu.Lock()
	defer gl.mu.Unlock()

	_ = gl.enc.Encode(GameLog{
		Typ:     typ,
		Msg:     msg,
		Elapsed: fmt.Sprint(time.Since(gl.tStart)),
	})
}
