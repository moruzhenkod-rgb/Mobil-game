package main

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

type CharState int

const (
	CharIdle      CharState = iota
	CharCelebrate           // match made
	CharSad                 // out of moves
	CharDance               // level complete
	CharHit                 // got a bad swap (brief shake)
)

type Character struct {
	x, y    float64
	tick    int
	state   CharState
	stateTick int
}

func NewCharacter() *Character {
	return &Character{
		x: float64(ScreenW - 52),
		y: float64(BoardOffsetY) + float64(Rows*TileOuter)/2 - 50,
	}
}

func (c *Character) SetState(s CharState) {
	if s == CharCelebrate && c.state == CharDance {
		return
	}
	c.state = s
	c.stateTick = 0
}

func (c *Character) Update() {
	c.tick++
	c.stateTick++
	switch c.state {
	case CharCelebrate:
		if c.stateTick > 60 {
			c.state = CharIdle
		}
	case CharSad:
		// stays until level restarts
	case CharDance:
		// stays until level changes
	case CharHit:
		if c.stateTick > 25 {
			c.state = CharIdle
		}
	}
}

func (c *Character) Draw(screen *ebiten.Image) {
	bob := math.Sin(float64(c.tick) * 0.08) * 3

	switch c.state {
	case CharCelebrate:
		bob = math.Sin(float64(c.tick)*0.25) * 10
	case CharDance:
		bob = math.Sin(float64(c.tick)*0.3) * 12
	case CharHit:
		bob = math.Sin(float64(c.tick)*0.5) * 5
	case CharSad:
		bob = math.Sin(float64(c.tick)*0.04) * 2
	}

	cx := float32(c.x)
	cy := float32(c.y + bob)

	// Colours by state
	bodyC := color.RGBA{100, 180, 255, 255}
	skinC := color.RGBA{255, 220, 160, 255}
	eyeC := color.RGBA{40, 40, 80, 255}

	switch c.state {
	case CharCelebrate, CharDance:
		bodyC = color.RGBA{255, 200, 60, 255}
	case CharSad:
		bodyC = color.RGBA{110, 110, 160, 255}
		skinC = color.RGBA{200, 180, 140, 255}
	case CharHit:
		bodyC = color.RGBA{255, 100, 100, 255}
	}

	// ── Body ──
	vector.DrawFilledRect(screen, cx-16, cy+2, 32, 38, bodyC, false)
	// Collar / neck
	vector.DrawFilledRect(screen, cx-8, cy-2, 16, 8, bodyC, false)

	// ── Head ──
	vector.DrawFilledCircle(screen, cx, cy-6, 18, skinC, false)

	// ── Eyes ──
	if c.state == CharSad {
		// Angled down (sad)
		vector.DrawFilledRect(screen, cx-12, cy-8, 7, 3, eyeC, false)
		vector.DrawFilledRect(screen, cx+5, cy-8, 7, 3, eyeC, false)
		// Tear
		vector.DrawFilledCircle(screen, cx-9, cy-3, 2, color.RGBA{100, 180, 255, 200}, false)
	} else {
		// Normal eyes
		vector.DrawFilledCircle(screen, cx-7, cy-8, 3, eyeC, false)
		vector.DrawFilledCircle(screen, cx+7, cy-8, 3, eyeC, false)
		// Shine in eye
		vector.DrawFilledCircle(screen, cx-6, cy-9, 1, color.RGBA{255, 255, 255, 200}, false)
		vector.DrawFilledCircle(screen, cx+8, cy-9, 1, color.RGBA{255, 255, 255, 200}, false)
	}

	// ── Mouth ──
	switch c.state {
	case CharCelebrate, CharDance:
		// Big smile
		vector.DrawFilledRect(screen, cx-8, cy, 16, 4, color.RGBA{200, 80, 60, 255}, false)
		vector.DrawFilledRect(screen, cx-8, cy, 4, 7, color.RGBA{200, 80, 60, 255}, false)
		vector.DrawFilledRect(screen, cx+4, cy, 4, 7, color.RGBA{200, 80, 60, 255}, false)
	case CharSad:
		// Frown
		vector.DrawFilledRect(screen, cx-7, cy+3, 14, 3, color.RGBA{160, 80, 60, 255}, false)
		vector.DrawFilledRect(screen, cx-7, cy, 3, 6, color.RGBA{160, 80, 60, 255}, false)
		vector.DrawFilledRect(screen, cx+4, cy, 3, 6, color.RGBA{160, 80, 60, 255}, false)
	default:
		// Neutral smile
		vector.DrawFilledRect(screen, cx-6, cy+1, 12, 3, color.RGBA{180, 80, 60, 255}, false)
	}

	// ── Arms ──
	armY := cy + 10
	if c.state == CharCelebrate || c.state == CharDance {
		// Arms up + wave
		wave := float32(math.Sin(float64(c.tick)*0.3) * 5)
		vector.DrawFilledRect(screen, cx-32, armY-14+wave, 18, 7, bodyC, false)
		vector.DrawFilledRect(screen, cx+14, armY-14-wave, 18, 7, bodyC, false)
	} else if c.state == CharSad {
		// Arms drooped
		vector.DrawFilledRect(screen, cx-30, armY+10, 16, 6, bodyC, false)
		vector.DrawFilledRect(screen, cx+14, armY+10, 16, 6, bodyC, false)
	} else {
		// Normal idle arms
		idleSway := float32(math.Sin(float64(c.tick)*0.08) * 2)
		vector.DrawFilledRect(screen, cx-30, armY+idleSway, 16, 6, bodyC, false)
		vector.DrawFilledRect(screen, cx+14, armY-idleSway, 16, 6, bodyC, false)
	}

	// ── Legs ──
	legSway := float32(0)
	if c.state == CharDance {
		legSway = float32(math.Sin(float64(c.tick)*0.3) * 6)
	}
	vector.DrawFilledRect(screen, cx-14, cy+40, 10, 22, bodyC, false)
	vector.DrawFilledRect(screen, cx+4, cy+40, 10, 22, bodyC, false)
	// Feet
	vector.DrawFilledRect(screen, cx-18+legSway, cy+58, 16, 6, color.RGBA{60, 40, 80, 255}, false)
	vector.DrawFilledRect(screen, cx+2-legSway, cy+58, 16, 6, color.RGBA{60, 40, 80, 255}, false)

	// ── Hat (wizard hat) ──
	hatC := color.RGBA{80, 30, 140, 255}
	hatBounce := float32(0)
	if c.state == CharCelebrate || c.state == CharDance {
		hatC = color.RGBA{200, 160, 30, 255}
		hatBounce = float32(math.Sin(float64(c.tick)*0.35)) * 4
	}
	// Brim
	vector.DrawFilledRect(screen, cx-20, cy-22+hatBounce, 40, 5, hatC, false)
	// Cone segments
	vector.DrawFilledRect(screen, cx-10, cy-44+hatBounce, 20, 23, hatC, false)
	vector.DrawFilledRect(screen, cx-5, cy-54+hatBounce, 10, 12, hatC, false)
	vector.DrawFilledRect(screen, cx-2, cy-60+hatBounce, 4, 8, hatC, false)
	// Star on hat
	starGlow := uint8(180 + 60*math.Sin(float64(c.tick)*0.15))
	vector.DrawFilledCircle(screen, cx, cy-46+hatBounce, 3, color.RGBA{255, 220, 80, starGlow}, false)

	// ── Magic sparkles (celebrate / dance) ──────────────────────────
	if c.state == CharCelebrate || c.state == CharDance {
		for i := 0; i < 3; i++ {
			angle := float64(c.tick)*0.2 + float64(i)*2.09
			sr := float32(28 + i*8)
			sx := cx + float32(math.Cos(angle))*sr
			sy := cy - 20 + float32(math.Sin(angle))*sr*0.5
			spA := uint8(160 + 80*math.Sin(float64(c.tick)*0.3+float64(i)))
			vector.DrawFilledCircle(screen, sx, sy, 3, color.RGBA{255, 220, 100, spA}, false)
		}
	}
}

