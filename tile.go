package main

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// ── Colours ───────────────────────────────────────────────────────

var gemColors = [GemCount]color.RGBA{
	TileEmpty:  {0, 0, 0, 0},
	TileRed:    {220, 60, 60, 255},
	TileBlue:   {60, 120, 220, 255},
	TileGreen:  {60, 180, 80, 255},
	TileYellow: {230, 210, 40, 255},
	TilePurple: {160, 60, 200, 255},
	TileOrange: {230, 130, 40, 255},
}

var gemDark = [GemCount]color.RGBA{
	TileEmpty:  {0, 0, 0, 0},
	TileRed:    {160, 30, 30, 255},
	TileBlue:   {30, 70, 160, 255},
	TileGreen:  {30, 120, 40, 255},
	TileYellow: {180, 160, 20, 255},
	TilePurple: {100, 20, 140, 255},
	TileOrange: {170, 80, 10, 255},
}

// tileColors alias for HUD/screens
var tileColors = gemColors

func GemColor(kind, base TileKind) color.RGBA {
	k := kind
	if kind.IsBonus() || kind.IsObstacle() {
		k = base
	}
	if int(k) > 0 && int(k) < len(gemColors) {
		return gemColors[k]
	}
	return color.RGBA{180, 180, 180, 255}
}

func GemDark(kind, base TileKind) color.RGBA {
	k := kind
	if kind.IsBonus() || kind.IsObstacle() {
		k = base
	}
	if int(k) > 0 && int(k) < len(gemDark) {
		return gemDark[k]
	}
	return color.RGBA{100, 100, 100, 255}
}

// ── Tile ──────────────────────────────────────────────────────────

type Tile struct {
	Kind      TileKind
	BaseColor TileKind // bonus tiles: their gem colour; ice/chain: gem underneath
	Layers    int      // TileIce: 2=full, 1=cracked

	X, Y            float64
	TargetX, TargetY float64
	StartX, StartY  float64

	SwapTick  int
	Swapping  bool
	VelY      float64
	ExplodeTick int
	Exploding   bool
	Selected    bool
	Age         int

	// Damage flash
	DmgFlash int
}

func easeInOut(t float64) float64 { return t * t * (3 - 2*t) }

func (t *Tile) Update() {
	t.Age++
	if t.DmgFlash > 0 {
		t.DmgFlash--
	}

	if t.Swapping {
		t.SwapTick++
		p := float64(t.SwapTick) / SwapDuration
		if p >= 1 {
			t.X, t.Y = t.TargetX, t.TargetY
			t.Swapping = false
		} else {
			e := easeInOut(p)
			t.X = t.StartX + e*(t.TargetX-t.StartX)
			t.Y = t.StartY + e*(t.TargetY-t.StartY)
		}
		return
	}
	if t.Exploding {
		t.ExplodeTick++
		return
	}
	if t.Y < t.TargetY-0.5 {
		t.VelY += 0.9
		t.Y += t.VelY
		if t.Y >= t.TargetY {
			t.Y = t.TargetY
			t.VelY = 0
		}
	}
}

func (t *Tile) IsFalling() bool  { return !t.Swapping && !t.Exploding && t.Y < t.TargetY-0.5 }
func (t *Tile) IsAnimating() bool { return t.Swapping || t.Exploding || t.IsFalling() }

// ── Draw ──────────────────────────────────────────────────────────

func (t *Tile) Draw(screen *ebiten.Image) {
	if t.Kind == TileEmpty {
		return
	}
	switch {
	case t.Kind == TileIce:
		t.drawIce(screen)
	case t.Kind == TileStone:
		t.drawStone(screen)
	case t.Kind == TileChain:
		t.drawChain(screen)
	default:
		t.drawGem(screen)
	}
}

func (t *Tile) drawGem(screen *ebiten.Image) {
	x, y := float32(t.X), float32(t.Y)
	sz := float32(TileSize)
	mainC := GemColor(t.Kind, t.BaseColor)
	darkC := GemDark(t.Kind, t.BaseColor)

	if t.Exploding {
		p := float64(t.ExplodeTick) / float64(ExplodeDur)
		if p > 1 {
			p = 1
		}
		sc := float32(1 + p*0.6)
		mainC.A = uint8(255 * (1 - p))
		cx, cy := x+sz/2, y+sz/2
		s := sz * sc
		vector.DrawFilledRect(screen, cx-s/2, cy-s/2, s, s, mainC, false)
		return
	}

	scale := float32(1.0)
	if t.Selected {
		scale = 1 + float32(math.Sin(float64(t.Age)*0.15)*0.05+0.06)
	}
	ssz := sz * scale
	cx, cy := x+sz/2, y+sz/2
	rx, ry := cx-ssz/2, cy-ssz/2

	vector.DrawFilledRect(screen, rx+3, ry+4, ssz, ssz, color.RGBA{0, 0, 0, 70}, false)
	vector.DrawFilledRect(screen, rx, ry, ssz, ssz, mainC, false)
	vector.DrawFilledRect(screen, rx, ry, ssz, 4, color.RGBA{255, 255, 255, 90}, false)
	vector.DrawFilledRect(screen, rx, ry, 4, ssz, color.RGBA{255, 255, 255, 60}, false)
	vector.DrawFilledRect(screen, rx, ry+ssz-4, ssz, 4, darkC, false)
	vector.DrawFilledRect(screen, rx+ssz-4, ry, 4, ssz, darkC, false)
	vector.DrawFilledCircle(screen, cx-ssz*0.18, cy-ssz*0.2, ssz*0.12, color.RGBA{255, 255, 255, 110}, false)

	switch t.Kind {
	case TileRowBomb:
		drawRowBombIcon(screen, rx, ry, ssz)
	case TileColBomb:
		drawColBombIcon(screen, rx, ry, ssz)
	case TileBomb:
		drawBombIcon(screen, cx, cy, ssz)
	case TileRainbow:
		drawRainbowIcon(screen, cx, cy, ssz)
	}

	if t.Selected {
		vector.StrokeRect(screen, rx-2, ry-2, ssz+4, ssz+4, 3, color.RGBA{255, 255, 255, 230}, false)
	}
}

func (t *Tile) drawIce(screen *ebiten.Image) {
	x, y := float32(t.X), float32(t.Y)
	sz := float32(TileSize)

	if t.Exploding {
		p := float64(t.ExplodeTick) / float64(ExplodeDur)
		sc := float32(1 + p*0.5)
		c := color.RGBA{160, 220, 255, uint8(200 * (1 - p))}
		cx, cy := x+sz/2, y+sz/2
		s := sz * sc
		vector.DrawFilledRect(screen, cx-s/2, cy-s/2, s, s, c, false)
		return
	}

	// Draw gem underneath (dimmed)
	gemC := GemColor(TileIce, t.BaseColor)
	gemC.A = 160
	vector.DrawFilledRect(screen, x, y, sz, sz, gemC, false)

	// Ice overlay
	if t.Layers >= 2 {
		// Full ice — opaque blue overlay
		vector.DrawFilledRect(screen, x, y, sz, sz, color.RGBA{140, 210, 255, 200}, false)
		// Grid cracks (just lines)
		vector.DrawFilledRect(screen, x+sz*0.33, y, 2, sz, color.RGBA{200, 240, 255, 180}, false)
		vector.DrawFilledRect(screen, x+sz*0.66, y, 2, sz, color.RGBA{200, 240, 255, 180}, false)
		vector.DrawFilledRect(screen, x, y+sz*0.33, sz, 2, color.RGBA{200, 240, 255, 180}, false)
		vector.DrawFilledRect(screen, x, y+sz*0.66, sz, 2, color.RGBA{200, 240, 255, 180}, false)
	} else {
		// Cracked ice — semi-transparent
		vector.DrawFilledRect(screen, x, y, sz, sz, color.RGBA{160, 220, 255, 130}, false)
		// Crack lines (diagonal)
		vector.DrawFilledRect(screen, x, y, sz, 3, color.RGBA{255, 255, 255, 160}, false)
		vector.DrawFilledRect(screen, x, y+sz*0.5, sz, 3, color.RGBA{200, 240, 255, 140}, false)
		vector.DrawFilledRect(screen, x+sz*0.5, y, 3, sz, color.RGBA{200, 240, 255, 140}, false)
	}

	// Shine
	vector.DrawFilledCircle(screen, x+sz*0.25, y+sz*0.2, sz*0.1, color.RGBA{255, 255, 255, 120}, false)

	if t.DmgFlash > 0 {
		alpha := uint8(180 * t.DmgFlash / 8)
		vector.DrawFilledRect(screen, x, y, sz, sz, color.RGBA{255, 255, 255, alpha}, false)
	}
}

func (t *Tile) drawStone(screen *ebiten.Image) {
	x, y := float32(t.X), float32(t.Y)
	sz := float32(TileSize)

	if t.Exploding {
		p := float64(t.ExplodeTick) / float64(ExplodeDur)
		sc := float32(1 + p*0.5)
		c := color.RGBA{120, 120, 120, uint8(200 * (1 - p))}
		cx, cy := x+sz/2, y+sz/2
		s := sz * sc
		vector.DrawFilledRect(screen, cx-s/2, cy-s/2, s, s, c, false)
		return
	}

	// Stone body
	vector.DrawFilledRect(screen, x+2, y+3, sz, sz, color.RGBA{60, 55, 65, 180}, false) // shadow
	vector.DrawFilledRect(screen, x, y, sz, sz, color.RGBA{110, 105, 120, 255}, false)
	// Highlight top
	vector.DrawFilledRect(screen, x, y, sz, 4, color.RGBA{160, 155, 170, 255}, false)
	vector.DrawFilledRect(screen, x, y, 4, sz, color.RGBA{160, 155, 170, 255}, false)
	// Dark bottom
	vector.DrawFilledRect(screen, x, y+sz-4, sz, 4, color.RGBA{70, 65, 80, 255}, false)
	vector.DrawFilledRect(screen, x+sz-4, y, 4, sz, color.RGBA{70, 65, 80, 255}, false)
	// Crack pattern
	vector.DrawFilledRect(screen, x+sz*0.3, y+sz*0.1, 3, sz*0.4, color.RGBA{70, 65, 80, 200}, false)
	vector.DrawFilledRect(screen, x+sz*0.3, y+sz*0.5, sz*0.4, 3, color.RGBA{70, 65, 80, 200}, false)
	vector.DrawFilledRect(screen, x+sz*0.6, y+sz*0.5, 3, sz*0.4, color.RGBA{70, 65, 80, 200}, false)

	if t.DmgFlash > 0 {
		alpha := uint8(180 * t.DmgFlash / 8)
		vector.DrawFilledRect(screen, x, y, sz, sz, color.RGBA{255, 100, 30, alpha}, false)
	}
}

func (t *Tile) drawChain(screen *ebiten.Image) {
	x, y := float32(t.X), float32(t.Y)
	sz := float32(TileSize)

	if t.Exploding {
		p := float64(t.ExplodeTick) / float64(ExplodeDur)
		sc := float32(1 + p*0.5)
		mainC := GemColor(TileChain, t.BaseColor)
		mainC.A = uint8(200 * (1 - p))
		cx, cy := x+sz/2, y+sz/2
		s := sz * sc
		vector.DrawFilledRect(screen, cx-s/2, cy-s/2, s, s, mainC, false)
		return
	}

	// Draw underlying gem (dimmed)
	gemC := GemColor(TileChain, t.BaseColor)
	vector.DrawFilledRect(screen, x, y, sz, sz, gemC, false)
	vector.DrawFilledRect(screen, x, y, sz, 4, color.RGBA{255, 255, 255, 60}, false)

	// Chain overlay — horizontal and vertical bars
	chainC := color.RGBA{200, 170, 80, 230}
	chainD := color.RGBA{140, 110, 30, 255}
	// Horizontal chains
	vector.DrawFilledRect(screen, x, y+sz*0.3, sz, 10, chainC, false)
	vector.DrawFilledRect(screen, x, y+sz*0.6, sz, 10, chainC, false)
	// Vertical chains
	vector.DrawFilledRect(screen, x+sz*0.3, y, 10, sz, chainC, false)
	vector.DrawFilledRect(screen, x+sz*0.6, y, 10, sz, chainC, false)
	// Chain links (darker squares at intersections)
	for _, px := range []float32{x + sz*0.3, x + sz*0.6} {
		for _, py := range []float32{y + sz*0.3, y + sz*0.6} {
			vector.DrawFilledRect(screen, px, py, 10, 10, chainD, false)
		}
	}

	if t.DmgFlash > 0 {
		alpha := uint8(180 * t.DmgFlash / 8)
		vector.DrawFilledRect(screen, x, y, sz, sz, color.RGBA{255, 255, 255, alpha}, false)
	}
}

// ── Bonus icons ───────────────────────────────────────────────────

func drawRowBombIcon(screen *ebiten.Image, rx, ry, sz float32) {
	mid := ry + sz/2 - 3
	vector.DrawFilledRect(screen, rx+8, mid, sz-16, 6, color.RGBA{255, 255, 255, 210}, false)
	vector.DrawFilledRect(screen, rx+8, mid-5, 5, 16, color.RGBA{255, 255, 255, 210}, false)
	vector.DrawFilledRect(screen, rx+sz-13, mid-5, 5, 16, color.RGBA{255, 255, 255, 210}, false)
}

func drawColBombIcon(screen *ebiten.Image, rx, ry, sz float32) {
	mid := rx + sz/2 - 3
	vector.DrawFilledRect(screen, mid, ry+8, 6, sz-16, color.RGBA{255, 255, 255, 210}, false)
	vector.DrawFilledRect(screen, mid-5, ry+8, 16, 5, color.RGBA{255, 255, 255, 210}, false)
	vector.DrawFilledRect(screen, mid-5, ry+sz-13, 16, 5, color.RGBA{255, 255, 255, 210}, false)
}

func drawBombIcon(screen *ebiten.Image, cx, cy, sz float32) {
	vector.DrawFilledCircle(screen, cx, cy, sz*0.27, color.RGBA{0, 0, 0, 110}, false)
	vector.DrawFilledRect(screen, cx-sz*0.17, cy-3, sz*0.34, 6, color.RGBA{255, 255, 255, 210}, false)
	vector.DrawFilledRect(screen, cx-3, cy-sz*0.17, 6, sz*0.34, color.RGBA{255, 255, 255, 210}, false)
}

func drawRainbowIcon(screen *ebiten.Image, cx, cy, sz float32) {
	cols := []color.RGBA{
		{255, 60, 60, 170}, {255, 160, 30, 170}, {255, 220, 30, 170},
		{60, 200, 80, 170}, {60, 120, 220, 170}, {160, 60, 200, 170},
	}
	for i, c := range cols {
		r := sz*0.1 + float32(i)*sz*0.025
		vector.DrawFilledCircle(screen, cx, cy, r, c, false)
	}
}
