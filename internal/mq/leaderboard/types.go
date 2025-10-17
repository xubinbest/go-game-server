package leaderboard

type GameScoreMessage struct {
	UserID string  `json:"user_id"`
	Score  float64 `json:"score"`
}
