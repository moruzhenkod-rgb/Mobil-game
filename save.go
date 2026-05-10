package main

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// SaveData is serialised to %APPDATA%\RunicCrush\save.json
type SaveData struct {
	// Progression
	AgeVerified   bool
	UnlockedLevel int
	Stars         [MaxLevels + 1]int
	BestScore     int

	// Currency & inventory
	Coins           int
	ExtraMoves      int
	BombBoosters    int
	RainbowBoosters int
	AdsRemoved      bool

	// Settings
	Language string
	SoundOn  bool
	MusicOn  bool
	NotifsOn bool

	// Account
	UserEmail string
	UserUID   string

	// Parental controls
	ParentalPIN  string // hashed 4-digit PIN (SHA-256 hex), empty = no PIN
	MonthlyLimit int    // in cents; 0 = unlimited
	MonthlySpent int    // cents spent this calendar month
	RequirePIN   bool

	// GDPR
	GDPRConsent    bool
	GDPRConsentDate string
}

var progress SaveData

func savePath() string {
	dir := filepath.Join(os.Getenv("APPDATA"), "RunicCrush")
	os.MkdirAll(dir, 0755)
	return filepath.Join(dir, "save.json")
}

func loadProgress() {
	data, err := os.ReadFile(savePath())
	if err != nil {
		progress = defaultSave()
		return
	}
	if err := json.Unmarshal(data, &progress); err != nil {
		progress = defaultSave()
		return
	}
	// Sanitise
	if progress.UnlockedLevel < 1 {
		progress.UnlockedLevel = 1
	}
	if progress.Language != "" {
		currentLang = Lang(progress.Language)
	}
}

func defaultSave() SaveData {
	return SaveData{
		UnlockedLevel: 1,
		SoundOn:       true,
		MusicOn:       true,
		NotifsOn:      true,
		Coins:         50, // starter coins
	}
}

func saveProgress() {
	data, err := json.MarshalIndent(progress, "", "  ")
	if err != nil {
		return
	}
	os.WriteFile(savePath(), data, 0644)
}

func markComplete(levelNum, score, maxMoves, movesLeft int) {
	stars := calcStars(score, maxMoves, movesLeft)
	if stars > progress.Stars[levelNum] {
		progress.Stars[levelNum] = stars
	}
	if score > progress.BestScore {
		progress.BestScore = score
	}
	next := levelNum + 1
	if next > progress.UnlockedLevel {
		progress.UnlockedLevel = next
		if progress.UnlockedLevel > MaxLevels {
			progress.UnlockedLevel = MaxLevels
		}
	}
	saveProgress()
	go fb.PushSave()
}

func calcStars(score, maxMoves, movesLeft int) int {
	if movesLeft <= 0 || score <= 0 {
		return 1
	}
	pct := float64(movesLeft) / float64(maxMoves)
	switch {
	case pct >= 0.5:
		return 3
	case pct >= 0.25:
		return 2
	default:
		return 1
	}
}

func isLevelUnlocked(n int) bool {
	return n <= progress.UnlockedLevel
}

// AddCoins safely adds coins and persists.
func AddCoins(amount int) {
	progress.Coins += amount
	if progress.Coins < 0 {
		progress.Coins = 0
	}
	saveProgress()
}
