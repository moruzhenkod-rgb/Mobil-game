package main

import (
	"image/color"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// ── Screen fade transition ────────────────────────────────────────

type Transition struct {
	alpha  float32
	speed  float32
	dir    float32 // +1 fade to black, -1 fade to clear
	active bool
	onDone func()
}

var sceneTransition = &Transition{}

// FadeOut fades to black, then calls onDone.
func FadeOut(speed float32, onDone func()) {
	sceneTransition.alpha = 0
	sceneTransition.speed = speed
	sceneTransition.dir = 1
	sceneTransition.active = true
	sceneTransition.onDone = onDone
}

// FadeIn fades from black to clear, then calls onDone.
func FadeIn(speed float32, onDone func()) {
	sceneTransition.alpha = 255
	sceneTransition.speed = speed
	sceneTransition.dir = -1
	sceneTransition.active = true
	sceneTransition.onDone = onDone
}

func (t *Transition) Update() {
	if !t.active {
		return
	}
	t.alpha += t.dir * t.speed
	done := false
	if t.dir > 0 && t.alpha >= 255 {
		t.alpha = 255
		done = true
	} else if t.dir < 0 && t.alpha <= 0 {
		t.alpha = 0
		done = true
	}
	if done {
		t.active = false
		if t.onDone != nil {
			cb := t.onDone
			t.onDone = nil
			cb()
		}
	}
}

func (t *Transition) Draw(screen *ebiten.Image) {
	if t.alpha <= 0 {
		return
	}
	vector.DrawFilledRect(screen, 0, 0, ScreenW, ScreenH,
		color.RGBA{0, 0, 0, uint8(t.alpha)}, false)
}

// ── Screen shake ──────────────────────────────────────────────────

type screenShake struct {
	amp   float32
	ticks int
}

func (s *screenShake) Trigger(amp float32, ticks int) {
	if amp > s.amp {
		s.amp = amp
	}
	if ticks > s.ticks {
		s.ticks = ticks
	}
}

func (s *screenShake) Update() {
	if s.ticks > 0 {
		s.ticks--
		if s.ticks == 0 {
			s.amp = 0
		}
	}
}

func (s *screenShake) Offset() (float64, float64) {
	if s.ticks == 0 {
		return 0, 0
	}
	decay := float32(s.ticks) / 18.0
	if decay > 1 {
		decay = 1
	}
	a := s.amp * decay
	return float64((rand.Float32()*2 - 1) * a),
		float64((rand.Float32()*2 - 1) * a)
}
