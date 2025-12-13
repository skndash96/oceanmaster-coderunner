package engine

type GameState struct {
	Tick   int   `json:"tick"`
	Series []int `json:"series"`
}

type Action struct {
	Element int `json:"element"`
}

func NewGameState() GameState {
	return GameState{
		Tick:   0,
		Series: []int{},
	}
}

// gs passed by value
func UpdateGameState(gs GameState, a []Action) (GameState, bool, bool) {
	gs.Tick++

	for _, action := range a {
		gs.Series = append(gs.Series, action.Element)
	}

	isValid := true
	hasEnded := gs.Tick >= 30

	return gs, isValid, hasEnded
}
