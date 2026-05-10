package main

// ── Goal types ────────────────────────────────────────────────────

type GoalType int

const (
	GoalClearColor  GoalType = iota // clear N gems of specific colour
	GoalScore                       // reach score threshold
	GoalClearAny                    // clear N gems total
	GoalClearTwo                    // clear N of colour A AND M of colour B
	GoalBreakIce                    // break N ice tiles
	GoalBreakStone                  // break N stone tiles
	GoalBreakChain                  // break N chain tiles
	GoalBreakObstacle               // break any N obstacles
)

// ── Level ─────────────────────────────────────────────────────────

type Level struct {
	Number int
	Biome  int // 1-5

	MaxMoves  int
	ScoreGoal int

	// Goal
	GoalType   GoalType
	GoalKind   TileKind
	GoalCount  int
	GoalKind2  TileKind
	GoalCount2 int

	// Obstacles
	IceCount   int
	IceLayers  int // 1 or 2
	StoneCount int
	ChainCount int
}

// ── Biome metadata ────────────────────────────────────────────────

type BiomeInfo struct {
	Name        string
	Description string
	Color       [3]uint8
}

var biomes = [BiomeCount + 1]BiomeInfo{
	0: {},
	1: {"Forest Ruins", "Ancient ruins hidden in an enchanted forest", [3]uint8{60, 130, 60}},
	2: {"Crystal Caves", "Gems frozen in eternal ice deep underground", [3]uint8{80, 160, 220}},
	3: {"Volcanic Forge", "Stone barriers forged by lava and fire", [3]uint8{200, 80, 30}},
	4: {"Arcane Library", "Ancient tomes chained by forgotten magic", [3]uint8{120, 60, 200}},
	5: {"Sky Citadel", "The final fortress above the clouds", [3]uint8{200, 180, 60}},
}

// ── Level generator ───────────────────────────────────────────────

func GetLevel(n int) Level {
	if n < 1 {
		n = 1
	}
	if n > MaxLevels {
		n = MaxLevels
	}
	return generateLevel(n)
}

func BiomeOf(n int) int {
	return (n-1)/LevelsPerBiome + 1
}

func generateLevel(n int) Level {
	biome := BiomeOf(n)
	pos := (n - 1) % LevelsPerBiome // 0..19 within biome
	t := float64(pos) / 19           // 0.0 .. 1.0 difficulty within biome

	l := Level{Number: n, Biome: biome}

	// ── Moves: decreases across biome and globally ──
	baseMoves := 32 - biome*3
	l.MaxMoves = baseMoves - int(t*8)
	if l.MaxMoves < 12 {
		l.MaxMoves = 12
	}

	// ── Score goal: increases per biome ──
	l.ScoreGoal = (400 + biome*150) + int(t*float64(biome*200))

	// ── Obstacles introduced per biome ──
	switch biome {
	case 1: // No obstacles — pure gems
		l = setGoalBiome1(l, pos)

	case 2: // Ice tiles
		l.IceLayers = 1
		if pos >= 10 {
			l.IceLayers = 2
		}
		l.IceCount = 2 + pos/3
		if l.IceCount > 10 {
			l.IceCount = 10
		}
		l = setGoalBiome2(l, pos)

	case 3: // Stone tiles + some ice
		l.StoneCount = 2 + pos/4
		if l.StoneCount > 8 {
			l.StoneCount = 8
		}
		l.IceCount = pos / 5
		l.IceLayers = 1
		l = setGoalBiome3(l, pos)

	case 4: // Chain tiles + ice
		l.ChainCount = 3 + pos/3
		if l.ChainCount > 12 {
			l.ChainCount = 12
		}
		l.IceCount = pos / 5
		l.StoneCount = pos / 8
		l.IceLayers = 2
		l = setGoalBiome4(l, pos)

	case 5: // All obstacles
		l.IceCount = 2 + pos/5
		l.IceLayers = 2
		l.StoneCount = 2 + pos/6
		l.ChainCount = 2 + pos/5
		l = setGoalBiome5(l, pos)
	}

	return l
}

// ── Per-biome goal logic ──────────────────────────────────────────

var b1GoalKinds = []TileKind{TileRed, TileBlue, TileGreen, TileYellow, TilePurple, TileOrange}

func setGoalBiome1(l Level, pos int) Level {
	switch {
	case pos < 4:
		l.GoalType = GoalClearColor
		l.GoalKind = b1GoalKinds[pos%len(b1GoalKinds)]
		l.GoalCount = 10 + pos*2
	case pos < 8:
		l.GoalType = GoalClearTwo
		l.GoalKind = b1GoalKinds[pos%len(b1GoalKinds)]
		l.GoalCount = 8 + pos
		l.GoalKind2 = b1GoalKinds[(pos+2)%len(b1GoalKinds)]
		l.GoalCount2 = 8 + pos
	case pos < 12:
		l.GoalType = GoalClearAny
		l.GoalCount = 30 + pos*3
	case pos < 16:
		l.GoalType = GoalScore
		l.ScoreGoal = 600 + pos*80
	default:
		// Boss-ish: score + colour
		l.GoalType = GoalClearTwo
		l.GoalKind = TileRed
		l.GoalCount = 15 + (pos-16)*3
		l.GoalKind2 = TilePurple
		l.GoalCount2 = 15 + (pos-16)*3
		l.ScoreGoal = 800 + pos*100
	}
	return l
}

func setGoalBiome2(l Level, pos int) Level {
	switch {
	case pos < 5:
		l.GoalType = GoalBreakIce
		l.GoalCount = l.IceCount
	case pos < 10:
		l.GoalType = GoalClearColor
		l.GoalKind = b1GoalKinds[pos%len(b1GoalKinds)]
		l.GoalCount = 12 + pos*2
	case pos < 15:
		l.GoalType = GoalBreakIce
		l.GoalCount = l.IceCount
		l.ScoreGoal += 300
	default:
		l.GoalType = GoalClearAny
		l.GoalCount = 40 + pos*2
	}
	return l
}

func setGoalBiome3(l Level, pos int) Level {
	switch {
	case pos < 5:
		l.GoalType = GoalBreakStone
		l.GoalCount = l.StoneCount
	case pos < 10:
		l.GoalType = GoalClearColor
		l.GoalKind = b1GoalKinds[pos%len(b1GoalKinds)]
		l.GoalCount = 14 + pos*2
	case pos < 15:
		l.GoalType = GoalBreakObstacle
		l.GoalCount = l.StoneCount + l.IceCount
	default:
		l.GoalType = GoalScore
		l.ScoreGoal += 400
	}
	return l
}

func setGoalBiome4(l Level, pos int) Level {
	switch {
	case pos < 5:
		l.GoalType = GoalBreakChain
		l.GoalCount = l.ChainCount
	case pos < 10:
		l.GoalType = GoalClearTwo
		l.GoalKind = b1GoalKinds[pos%len(b1GoalKinds)]
		l.GoalCount = 12 + pos
		l.GoalKind2 = b1GoalKinds[(pos+3)%len(b1GoalKinds)]
		l.GoalCount2 = 10 + pos
	case pos < 15:
		l.GoalType = GoalBreakObstacle
		l.GoalCount = l.ChainCount + l.IceCount + l.StoneCount
	default:
		l.GoalType = GoalScore
		l.ScoreGoal += 600
	}
	return l
}

func setGoalBiome5(l Level, pos int) Level {
	switch {
	case pos < 5:
		l.GoalType = GoalBreakObstacle
		l.GoalCount = l.IceCount + l.StoneCount + l.ChainCount
	case pos < 10:
		l.GoalType = GoalClearTwo
		l.GoalKind = TileRed
		l.GoalCount = 15 + pos
		l.GoalKind2 = TileBlue
		l.GoalCount2 = 15 + pos
	case pos < 15:
		l.GoalType = GoalScore
		l.ScoreGoal += 800
	default: // 15-19 boss levels
		l.GoalType = GoalBreakObstacle
		l.GoalCount = l.IceCount + l.StoneCount + l.ChainCount
		l.ScoreGoal += 1000
		l.MaxMoves -= 2
		if l.MaxMoves < 12 {
			l.MaxMoves = 12
		}
	}
	return l
}
