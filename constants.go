package main

const (
	ScreenW = 540
	ScreenH = 960

	Cols = 7
	Rows = 9

	TileSize    = 60
	TilePadding = 4
	TileOuter   = TileSize + TilePadding

	BoardOffsetX = (ScreenW - Cols*TileOuter) / 2
	BoardOffsetY = 160

	HUDHeight = 150

	SwapDuration = 12
	ExplodeDur   = 20

	MaxLevels  = 100
	BiomeCount = 5
	LevelsPerBiome = 20
)

// ── Game States ───────────────────────────────────────────────────

type GameState int

const (
	StateIntro       GameState = iota // studio splash / GIF intro
	StateAgeGate
	StateMainMenu              // title + play/settings
	StateChapterSelect         // choose biome
	StateLevelSelect           // choose level within biome
	StateDialogue              // story dialogue overlay
	StateBiomeIntro            // biome unlock cutscene
	StatePlaying
	StateWin
	StateLose
	StateSettings              // settings screen
	StateShop                  // in-game shop
	StateLeaderboard           // online leaderboard
	StateCollection            // rune / character collection
	StateQuests                // daily quests
	StateMessages              // inbox / mail
)

// ── Tile Kinds ────────────────────────────────────────────────────

type TileKind int

const (
	TileEmpty TileKind = iota

	// Base gems
	TileRed
	TileBlue
	TileGreen
	TileYellow
	TilePurple
	TileOrange
	GemCount // = 7

	// Power-ups (created by matching 4+)
	TileRowBomb // clears full row
	TileColBomb // clears full column
	TileBomb    // 3×3 area
	TileRainbow // clears all of BaseColor

	// Obstacles (placed by level; do not participate in matches)
	TileIce   // frozen gem, cracked by adjacent match (1–2 layers)
	TileStone // stone block, only cleared by bomb explosion
	TileChain // chained gem, adjacent match breaks chain → normal gem
)

func (k TileKind) IsBonus() bool    { return k >= TileRowBomb && k <= TileRainbow }
func (k TileKind) IsGem() bool      { return k >= TileRed && k < GemCount }
func (k TileKind) IsObstacle() bool { return k >= TileIce }
