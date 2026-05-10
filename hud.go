package main

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

func DrawHUD(screen *ebiten.Image, g *Game) {
	l := g.level

	// Background
	vector.DrawFilledRect(screen, 0, 0, ScreenW, HUDHeight, color.RGBA{10, 5, 28, 245}, false)
	vector.DrawFilledRect(screen, 0, float32(HUDHeight-3), ScreenW, 3, color.RGBA{100, 50, 180, 160}, false)

	// — Level —
	if FontBtn != nil {
		FCenter(screen, fmt.Sprintf("LEVEL %d", g.levelNum), float64(ScreenW)/2, 6, FontBtn, ColGold)
	} else {
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("LEVEL %d", g.levelNum), ScreenW/2-30, 8)
	}

	// — Moves (left) —
	vector.DrawFilledRect(screen, 10, 28, 112, 116, color.RGBA{18, 10, 45, 200}, false)
	if FontSm != nil {
		FCenter(screen, "MOVES", 66, 32, FontSm, ColGray)
		FCenter(screen, fmt.Sprintf("%d", g.movesLeft), 66, 50, FontMd, ColWhite)
	} else {
		ebitenutil.DebugPrintAt(screen, "MOVES", 30, 36)
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("%d", g.movesLeft), 50, 56)
	}
	if l.MaxMoves > 0 {
		pct := float32(g.movesLeft) / float32(l.MaxMoves)
		bc := color.RGBA{80, 200, 100, 255}
		if pct < 0.35 {
			bc = color.RGBA{220, 80, 60, 255}
		} else if pct < 0.6 {
			bc = color.RGBA{220, 180, 40, 255}
		}
		vector.DrawFilledRect(screen, 18, 82, 96, 8, color.RGBA{40, 25, 65, 255}, false)
		vector.DrawFilledRect(screen, 18, 82, 96*pct, 8, bc, false)
	}
	if FontSm != nil {
		FCenter(screen, fmt.Sprintf("/%d", l.MaxMoves), 66, 94, FontSm, ColDim)
	} else {
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("/%d", l.MaxMoves), 20, 96)
	}

	// — Score (center) —
	vector.DrawFilledRect(screen, 130, 28, 160, 116, color.RGBA{18, 10, 45, 200}, false)
	if FontSm != nil {
		FCenter(screen, "SCORE", float64(ScreenW)/2, 32, FontSm, ColGray)
		scoreStr := fmt.Sprintf("%d", g.score)
		FCenter(screen, scoreStr, float64(ScreenW)/2, 50, FontMd, ColWhite)
	} else {
		ebitenutil.DebugPrintAt(screen, "SCORE", ScreenW/2-18, 36)
		scoreStr := fmt.Sprintf("%d", g.score)
		ebitenutil.DebugPrintAt(screen, scoreStr, ScreenW/2-len(scoreStr)*3, 56)
	}
	if l.ScoreGoal > 0 {
		pct := float32(g.score) / float32(l.ScoreGoal)
		if pct > 1 {
			pct = 1
		}
		vector.DrawFilledRect(screen, 138, 82, 144, 8, color.RGBA{40, 25, 65, 255}, false)
		vector.DrawFilledRect(screen, 138, 82, 144*pct, 8, color.RGBA{255, 210, 50, 255}, false)
		if FontSm != nil {
			FCenter(screen, fmt.Sprintf("/%d", l.ScoreGoal), float64(ScreenW)/2, 94, FontSm, ColDim)
		} else {
			ebitenutil.DebugPrintAt(screen, fmt.Sprintf("/%d", l.ScoreGoal), ScreenW/2-20, 96)
		}
	}

	// — Goal (right) —
	vector.DrawFilledRect(screen, 298, 28, 132, 116, color.RGBA{18, 10, 45, 200}, false)
	if FontSm != nil {
		FCenter(screen, "GOAL", float64(ScreenW-66), 32, FontSm, ColGray)
	} else {
		ebitenutil.DebugPrintAt(screen, "GOAL", ScreenW-72, 36)
	}
	drawGoal(screen, g, ScreenW-66, 56)

	// — Combo indicator —
	if g.board != nil && g.board.combo > 1 {
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("COMBO x%d!", g.board.combo), ScreenW/2-33, HUDHeight+4)
	}
}

func drawGoal(screen *ebiten.Image, g *Game, cx, y int) {
	l := g.level
	switch l.GoalType {
	case GoalClearColor:
		drawGemGoal(screen, l.GoalKind, g.cleared[l.GoalKind], l.GoalCount, cx, y)

	case GoalClearTwo:
		drawGemGoal(screen, l.GoalKind, g.cleared[l.GoalKind], l.GoalCount, cx-30, y)
		drawGemGoal(screen, l.GoalKind2, g.cleared[l.GoalKind2], l.GoalCount2, cx+30, y)

	case GoalClearAny:
		rem := l.GoalCount - g.clearedAny
		if rem < 0 {
			rem = 0
		}
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("ANY x%d", rem), cx-24, y)
		drawGoalBar(screen, float32(g.clearedAny), float32(l.GoalCount), cx, y+20)

	case GoalScore:
		if g.score >= l.ScoreGoal {
			ebitenutil.DebugPrintAt(screen, "DONE!", cx-15, y)
		} else {
			ebitenutil.DebugPrintAt(screen, "SCORE", cx-15, y)
			drawGoalBar(screen, float32(g.score), float32(l.ScoreGoal), cx, y+18)
		}

	case GoalBreakIce:
		rem := l.GoalCount - g.board.ClearedIce
		if rem < 0 {
			rem = 0
		}
		vector.DrawFilledRect(screen, float32(cx-11), float32(y), 22, 22, color.RGBA{140, 210, 255, 220}, false)
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("x%d", rem), cx-9, y+26)
		drawGoalBar(screen, float32(g.board.ClearedIce), float32(l.GoalCount), cx, y+44)

	case GoalBreakStone:
		rem := l.GoalCount - g.board.ClearedStone
		if rem < 0 {
			rem = 0
		}
		vector.DrawFilledRect(screen, float32(cx-11), float32(y), 22, 22, color.RGBA{110, 105, 120, 220}, false)
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("x%d", rem), cx-9, y+26)
		drawGoalBar(screen, float32(g.board.ClearedStone), float32(l.GoalCount), cx, y+44)

	case GoalBreakChain:
		rem := l.GoalCount - g.board.ClearedChain
		if rem < 0 {
			rem = 0
		}
		vector.DrawFilledRect(screen, float32(cx-11), float32(y), 22, 22, color.RGBA{200, 170, 80, 220}, false)
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("x%d", rem), cx-9, y+26)
		drawGoalBar(screen, float32(g.board.ClearedChain), float32(l.GoalCount), cx, y+44)

	case GoalBreakObstacle:
		cleared := g.board.ClearedIce + g.board.ClearedStone + g.board.ClearedChain
		rem := l.GoalCount - cleared
		if rem < 0 {
			rem = 0
		}
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("OBSTx%d", rem), cx-21, y)
		drawGoalBar(screen, float32(cleared), float32(l.GoalCount), cx, y+20)
	}
}

func drawGemGoal(screen *ebiten.Image, kind TileKind, cleared, total int, cx, y int) {
	rem := total - cleared
	if rem < 0 {
		rem = 0
	}
	sz := float32(22)
	gx := float32(cx) - sz/2
	if int(kind) > 0 && int(kind) < len(gemColors) {
		vector.DrawFilledRect(screen, gx, float32(y), sz, sz, gemColors[kind], false)
		vector.DrawFilledRect(screen, gx, float32(y), sz, 3, color.RGBA{255, 255, 255, 80}, false)
	}
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("x%d", rem), cx-12, y+26)
	drawGoalBar(screen, float32(cleared), float32(total), cx, y+42)
}

func drawGoalBar(screen *ebiten.Image, cur, total float32, cx, y int) {
	if total <= 0 {
		return
	}
	pct := cur / total
	if pct > 1 {
		pct = 1
	}
	bw := float32(100)
	x := float32(cx) - bw/2
	vector.DrawFilledRect(screen, x, float32(y), bw, 6, color.RGBA{40, 25, 65, 255}, false)
	vector.DrawFilledRect(screen, x, float32(y), bw*pct, 6, color.RGBA{100, 220, 120, 255}, false)
}
