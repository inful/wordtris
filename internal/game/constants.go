package game

// Block generation
const (
	SingleBlockProbability = 0.7  // 70% single blocks
	MaxBlocksPerGame       = 1000 // Reasonable game limit
)

// Scoring multipliers by word length
const (
	ScoreShort  = 100 // 2-3 characters: 100x
	ScoreMedium = 150 // 4-5 characters: 150x
	ScoreLong   = 200 // 6+ characters: 200x
)
