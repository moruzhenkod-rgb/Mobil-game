package main

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// ── Age Gate ──────────────────────────────────────────────────────

func DrawAgeGate(screen *ebiten.Image) {
	screen.Fill(color.RGBA{8, 4, 22, 255})

	ebitenutil.DebugPrintAt(screen, "=== RUNIC CRUSH ===", ScreenW/2-57, 180)
	ebitenutil.DebugPrintAt(screen, "Match gems, break ancient curses!", ScreenW/2-96, 202)

	vector.DrawFilledRect(screen, 60, 295, 420, 350, color.RGBA{18, 10, 46, 240}, false)
	vector.DrawFilledRect(screen, 60, 295, 420, 3, color.RGBA{120, 60, 200, 255}, false)
	vector.DrawFilledRect(screen, 60, 642, 420, 3, color.RGBA{120, 60, 200, 255}, false)

	ebitenutil.DebugPrintAt(screen, "Age Verification Required", ScreenW/2-75, 316)
	ebitenutil.DebugPrintAt(screen, "Are you 13 years of age or older?", ScreenW/2-99, 350)

	// YES
	vector.DrawFilledRect(screen, 96, 420, 148, 52, color.RGBA{50, 170, 70, 255}, false)
	vector.DrawFilledRect(screen, 96, 420, 148, 3, color.RGBA{100, 220, 120, 200}, false)
	ebitenutil.DebugPrintAt(screen, "YES, I AM 13+", 116, 440)

	// NO
	vector.DrawFilledRect(screen, 296, 420, 148, 52, color.RGBA{180, 50, 50, 255}, false)
	vector.DrawFilledRect(screen, 296, 420, 148, 3, color.RGBA{220, 100, 100, 200}, false)
	ebitenutil.DebugPrintAt(screen, "NO, I'M YOUNGER", 302, 440)

	ebitenutil.DebugPrintAt(screen, "This game contains in-app purchases.", ScreenW/2-108, 510)
	ebitenutil.DebugPrintAt(screen, "Parental supervision is recommended for", ScreenW/2-117, 528)
	ebitenutil.DebugPrintAt(screen, "players under 18 years of age.", ScreenW/2-90, 546)
	ebitenutil.DebugPrintAt(screen, "By playing you agree to our Terms of Service", ScreenW/2-132, 580)
	ebitenutil.DebugPrintAt(screen, "and Privacy Policy (available in-game).", ScreenW/2-117, 598)
}

// ── Win ───────────────────────────────────────────────────────────

func DrawWin(screen *ebiten.Image, score, levelNum, maxMoves, movesLeft int) {
	ov := ebiten.NewImage(ScreenW, ScreenH)
	ov.Fill(color.RGBA{0, 0, 0, 140})
	screen.DrawImage(ov, nil)

	vector.DrawFilledRect(screen, 65, 240, 410, 490, color.RGBA{16, 9, 42, 248}, false)
	vector.DrawFilledRect(screen, 65, 240, 410, 3, color.RGBA{255, 220, 60, 255}, false)
	vector.DrawFilledRect(screen, 65, 727, 410, 3, color.RGBA{255, 220, 60, 255}, false)

	if FontLg != nil {
		FCenter(screen, "LEVEL COMPLETE!", float64(ScreenW)/2, 256, FontLg, ColGold)
	} else {
		ebitenutil.DebugPrintAt(screen, "LEVEL COMPLETE!", ScreenW/2-45, 262)
	}

	// Stars
	stars := calcStars(score, maxMoves, movesLeft)
	for i := 0; i < 3; i++ {
		cx := float32(ScreenW/2 - 60 + i*60)
		c := color.RGBA{50, 40, 80, 255}
		if i < stars {
			c = color.RGBA{255, 215, 50, 255}
		}
		drawStar(screen, cx, 316, 22, c)
	}

	ebitenutil.DebugPrintAt(screen, "SCORE", ScreenW/2-18, 378)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("%d", score), ScreenW/2-len(fmt.Sprintf("%d", score))*3, 400)

	// Moves remaining bonus
	if movesLeft > 0 {
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Moves remaining: +%d", movesLeft*50), ScreenW/2-78, 432)
	}

	ebitenutil.DebugPrintAt(screen, encouragement(stars), ScreenW/2-len(encouragement(stars))*3, 468)

	// Stars earned label
	prevStars := progress.Stars[levelNum]
	if stars > prevStars {
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("New best: %d stars!", stars), ScreenW/2-54, 494)
	}

	// Next Level
	vector.DrawFilledRect(screen, 160, 610, 220, 52, color.RGBA{50, 170, 70, 255}, false)
	vector.DrawFilledRect(screen, 160, 610, 220, 3, color.RGBA{100, 220, 120, 200}, false)
	ebitenutil.DebugPrintAt(screen, "NEXT LEVEL >>", ScreenW/2-39, 629)

	// Level Select
	vector.DrawFilledRect(screen, 160, 676, 220, 44, color.RGBA{50, 35, 90, 255}, false)
	vector.DrawFilledRect(screen, 160, 676, 220, 3, color.RGBA{100, 70, 160, 200}, false)
	ebitenutil.DebugPrintAt(screen, "LEVEL SELECT", ScreenW/2-36, 692)
}

func encouragement(stars int) string {
	switch stars {
	case 3:
		return "PERFECT! 3 Stars!"
	case 2:
		return "Great job! 2 Stars!"
	default:
		return "Level cleared! 1 Star."
	}
}

// ── Lose ──────────────────────────────────────────────────────────

func DrawLose(screen *ebiten.Image, score int) {
	ov := ebiten.NewImage(ScreenW, ScreenH)
	ov.Fill(color.RGBA{0, 0, 0, 155})
	screen.DrawImage(ov, nil)

	vector.DrawFilledRect(screen, 65, 240, 410, 510, color.RGBA{16, 9, 42, 248}, false)
	vector.DrawFilledRect(screen, 65, 240, 410, 3, color.RGBA{200, 60, 60, 255}, false)
	vector.DrawFilledRect(screen, 65, 747, 410, 3, color.RGBA{200, 60, 60, 255}, false)

	if FontLg != nil {
		FCenter(screen, "OUT OF MOVES!", float64(ScreenW)/2, 256, FontLg, ColRed)
		FCenter(screen, "So close! Give it another try.", float64(ScreenW)/2, 298, FontSm, ColGray)
	} else {
		ebitenutil.DebugPrintAt(screen, "OUT OF MOVES!", ScreenW/2-39, 262)
		ebitenutil.DebugPrintAt(screen, "So close! Give it another try!", ScreenW/2-90, 290)
	}

	ebitenutil.DebugPrintAt(screen, "SCORE", ScreenW/2-18, 360)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("%d", score), ScreenW/2-len(fmt.Sprintf("%d", score))*3, 384)

	// Retry
	vector.DrawFilledRect(screen, 160, 560, 220, 52, color.RGBA{50, 120, 200, 255}, false)
	vector.DrawFilledRect(screen, 160, 560, 220, 3, color.RGBA{100, 170, 255, 200}, false)
	ebitenutil.DebugPrintAt(screen, "RETRY LEVEL", ScreenW/2-33, 579)

	// Watch Ad for +3 Moves
	if ads.IsReady() {
		vector.DrawFilledRect(screen, 160, 626, 220, 45, color.RGBA{180, 120, 20, 255}, false)
		vector.DrawFilledRect(screen, 160, 626, 220, 3, color.RGBA{255, 200, 60, 200}, false)
		ebitenutil.DebugPrintAt(screen, "WATCH AD: +3 MOVES", ScreenW/2-54, 642)
	}

	// Level Select
	vector.DrawFilledRect(screen, 160, 682, 220, 45, color.RGBA{50, 35, 90, 255}, false)
	vector.DrawFilledRect(screen, 160, 682, 220, 3, color.RGBA{100, 70, 160, 200}, false)
	ebitenutil.DebugPrintAt(screen, "LEVEL SELECT", ScreenW/2-36, 698)
}

// ── Leaderboard ───────────────────────────────────────────────────

func DrawLeaderboard(screen *ebiten.Image) {
	vector.DrawFilledRect(screen, 0, 0, ScreenW, 56, color.RGBA{10, 5, 28, 245}, false)
	ebitenutil.DebugPrintAt(screen, T("leaderboard"), ScreenW/2-30, 10)
	vector.DrawFilledRect(screen, 12, 10, 70, 28, color.RGBA{40, 25, 70, 220}, false)
	ebitenutil.DebugPrintAt(screen, T("back"), 20, 18)
	vector.DrawFilledRect(screen, 0, 52, ScreenW, 3, color.RGBA{120, 60, 200, 160}, false)

	entries := GetLeaderboard()
	y := 80
	ebitenutil.DebugPrintAt(screen, "TOP PLAYERS - ALL TIME", ScreenW/2-66, y)
	y += 30

	medalColors := []color.RGBA{
		{255, 215, 50, 255},  // gold
		{190, 190, 190, 255}, // silver
		{180, 100, 40, 255},  // bronze
	}
	for _, e := range entries {
		bg := color.RGBA{18, 10, 45, 200}
		if e.Name == "You" {
			bg = color.RGBA{30, 20, 70, 240}
		}
		vector.DrawFilledRect(screen, 20, float32(y), ScreenW-40, 44, bg, false)
		rankStr := fmt.Sprintf("#%d", e.Rank)
		if e.Rank <= 3 {
			mc := medalColors[e.Rank-1]
			vector.DrawFilledRect(screen, 24, float32(y+10), 24, 24, mc, false)
			ebitenutil.DebugPrintAt(screen, rankStr, 26, y+14)
		} else {
			ebitenutil.DebugPrintAt(screen, rankStr, 26, y+14)
		}
		ebitenutil.DebugPrintAt(screen, e.Name, 60, y+14)
		scoreStr := fmt.Sprintf("%d", e.Score)
		ebitenutil.DebugPrintAt(screen, scoreStr, ScreenW-20-len(scoreStr)*6, y+14)
		y += 52
	}

	y += 10
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Your best: %d", progress.BestScore), ScreenW/2-54, y)
	y += 24
	ebitenutil.DebugPrintAt(screen, "Tap anywhere to go back", ScreenW/2-66, y)
}

// ── Shared helpers ────────────────────────────────────────────────

func drawStar(screen *ebiten.Image, cx, cy, r float32, c color.RGBA) {
	vector.DrawFilledRect(screen, cx-r, cy-r*0.28, r*2, r*0.56, c, false)
	vector.DrawFilledRect(screen, cx-r*0.28, cy-r, r*0.56, r*2, c, false)
	d := r * 0.72
	dp := d * 0.24
	vector.DrawFilledRect(screen, cx-d, cy-dp, d*2, dp*2, c, false)
}
