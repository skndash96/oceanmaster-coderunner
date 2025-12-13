package engine

import (
	"encoding/json"
	"io"
	"sync"
)

type GameLogger struct {
	mu  sync.Mutex
	enc *json.Encoder
}

type GameLog struct {
	Typ GameLogType `json:"typ"`
	Msg any         `json:"msg"`
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
		enc: json.NewEncoder(w),
	}
}

func (gl *GameLogger) Log(typ GameLogType, msg ...any) {
	gl.mu.Lock()
	defer gl.mu.Unlock()

	_ = gl.enc.Encode(GameLog{
		Typ: typ,
		Msg: msg,
	})
}
