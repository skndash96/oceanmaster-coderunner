package queue

type MatchJob struct {
	ID       string `json:"id"`
	P1       string `json:"p1"`
	P2       string `json:"p2"`
	P1Code   string `json:"p1_code"`
	P2Code   string `json:"p2_code"`
	Priority int    `json:"priority"` // Higher value = higher priority (default manager = 100)
}

type MatchResult struct {
	ID     string `json:"id"`
	Winner int    `json:"winner"`
}
