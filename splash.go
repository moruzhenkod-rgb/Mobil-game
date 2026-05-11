package main

import (
	"image"
	"image/color"
	"image/draw"
	"image/gif"
	"math"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// ── GIF player ────────────────────────────────────────────────────

type GIFPlayer struct {
	frames  []*ebiten.Image
	delays  []int // centiseconds per frame
	current int
	tick    int
	Done    bool
}

func LoadGIF(path string) (*GIFPlayer, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	g, err := gif.DecodeAll(f)
	if err != nil {
		return nil, err
	}
	if len(g.Image) == 0 {
		return nil, nil
	}

	p := &GIFPlayer{delays: g.Delay}

	// Use GIF config dimensions for the canvas (not first-frame bounds)
	w, h := g.Config.Width, g.Config.Height
	if w == 0 || h == 0 {
		b := g.Image[0].Bounds()
		w, h = b.Max.X, b.Max.Y
	}
	canvasRect := image.Rect(0, 0, w, h)

	// Fill canvas with GIF background colour
	canvas := image.NewRGBA(canvasRect)
	bg := color.RGBA{0, 0, 0, 255}
	if len(g.Image[0].Palette) > int(g.BackgroundIndex) {
		c := g.Image[0].Palette[g.BackgroundIndex]
		r, gr, b2, a := c.RGBA()
		bg = color.RGBA{uint8(r >> 8), uint8(gr >> 8), uint8(b2 >> 8), uint8(a >> 8)}
	}
	draw.Draw(canvas, canvasRect, &image.Uniform{bg}, image.Point{}, draw.Src)

	// Backup canvas for DisposalPrevious
	prev := image.NewRGBA(canvasRect)
	copy(prev.Pix, canvas.Pix)

	for i, frame := range g.Image {
		// Save state before drawing (for DisposalPrevious)
		copy(prev.Pix, canvas.Pix)

		// Draw frame at its offset position
		fb := frame.Bounds()
		draw.Draw(canvas, fb, frame, fb.Min, draw.Over)

		// Snapshot the composed frame
		snap := image.NewRGBA(canvasRect)
		copy(snap.Pix, canvas.Pix)
		p.frames = append(p.frames, ebiten.NewImageFromImage(snap))

		// Apply disposal method for next frame
		disposal := byte(0)
		if i < len(g.Disposal) {
			disposal = g.Disposal[i]
		}
		switch disposal {
		case 2: // RestoreBackground
			draw.Draw(canvas, fb, &image.Uniform{bg}, image.Point{}, draw.Src)
		case 3: // RestorePrevious
			copy(canvas.Pix, prev.Pix)
		// 0 (unspecified) and 1 (DoNotDispose): leave canvas as-is
		}
	}
	return p, nil
}

func (p *GIFPlayer) Update() {
	if p.Done || len(p.frames) == 0 {
		return
	}
	p.tick++
	delay := p.delays[p.current]
	if delay <= 0 {
		delay = 8 // 80 ms default
	}
	// GIF delay in centiseconds → convert to game frames (60 fps)
	threshold := delay * 60 / 100
	if threshold < 1 {
		threshold = 1
	}
	if p.tick >= threshold {
		p.tick = 0
		p.current++
		if p.current >= len(p.frames) {
			p.Done = true
			p.current = len(p.frames) - 1
		}
	}
}

// UpdateLoop advances the GIF and resets to frame 0 when finished (for looping backgrounds).
func (p *GIFPlayer) UpdateLoop() {
	if len(p.frames) == 0 {
		return
	}
	p.tick++
	delay := p.delays[p.current]
	if delay <= 0 {
		delay = 8
	}
	threshold := delay * 60 / 100
	if threshold < 1 {
		threshold = 1
	}
	if p.tick >= threshold {
		p.tick = 0
		p.current++
		if p.current >= len(p.frames) {
			p.current = 0 // loop back
		}
	}
}

func (p *GIFPlayer) Draw(screen *ebiten.Image) {
	if len(p.frames) == 0 {
		return
	}
	// Black background so transparent GIF pixels show as black
	screen.Fill(color.RGBA{0, 0, 0, 255})

	img := p.frames[p.current]
	iw, ih := img.Bounds().Dx(), img.Bounds().Dy()
	if iw == 0 || ih == 0 {
		return
	}
	// Scale maintaining aspect ratio, centered
	scaleX := float64(ScreenW) / float64(iw)
	scaleY := float64(ScreenH) / float64(ih)
	scale := scaleX
	if scaleY < scaleX {
		scale = scaleY
	}
	drawW := float64(iw) * scale
	drawH := float64(ih) * scale
	offX := (float64(ScreenW) - drawW) / 2
	offY := (float64(ScreenH) - drawH) / 2

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(scale, scale)
	op.GeoM.Translate(offX, offY)
	op.Filter = ebiten.FilterLinear
	screen.DrawImage(img, op)
}

// ── Code-drawn ATP splash (fallback when no GIF found) ────────────

type ATPSplash struct {
	tick int
	Done bool
}

const splashDuration = 180 // 3 seconds at 60 fps

func (s *ATPSplash) Update() {
	s.tick++
	if s.tick >= splashDuration {
		s.Done = true
	}
}

// alpha returns 0→255→0 fade envelope over the splash duration.
func (s *ATPSplash) alpha() uint8 {
	t := float64(s.tick)
	fadeIn := 30.0
	fadeOut := 30.0
	total := float64(splashDuration)
	var a float64
	switch {
	case t < fadeIn:
		a = t / fadeIn
	case t > total-fadeOut:
		a = (total - t) / fadeOut
	default:
		a = 1.0
	}
	return uint8(a * 255)
}

func (s *ATPSplash) Draw(screen *ebiten.Image) {
	// Black background
	screen.Fill(color.RGBA{0, 0, 0, 255})

	a := s.alpha()
	if a == 0 {
		return
	}

	cx := float32(ScreenW) / 2
	cy := float32(ScreenH) / 2

	// Scale pops in slightly during fade-in
	scaleT := math.Min(float64(s.tick)/30.0, 1.0)
	scale := float32(0.85 + 0.15*scaleT)

	drawATPLogo(screen, cx, cy, scale, a)

	// "STUDIOS" caption below logo
	if FontSm != nil && a > 60 {
		capA := color.RGBA{180, 180, 180, uint8(float32(a) * 0.7)}
		FCenter(screen, "S T U D I O S", float64(cx), float64(cy)+58*float64(scale), FontSm, capA)
	}
}

// drawATPLogo renders the ATP letters using bold rectangular geometry,
// matching the angular stencil style of the actual logo.
func drawATPLogo(screen *ebiten.Image, cx, cy, scale float32, alpha uint8) {
	w := color.RGBA{255, 255, 255, alpha}

	s := scale        // shorthand
	u := float32(9) * s // unit — base stroke width

	// Total logo width ≈ 180u, center it
	// A: x offset -90u, T: -14u, P: +42u  (rough positions)

	// ── A ────────────────────────────────────────────────────────────
	ax := cx - 88*s
	ay := cy - 30*s
	h := float32(60) * s

	// Left leg
	vector.DrawFilledRect(screen, ax, ay, u, h, w, false)
	// Right leg
	vector.DrawFilledRect(screen, ax+36*s, ay, u, h, w, false)
	// Top bar (flat — angular A has no peak)
	vector.DrawFilledRect(screen, ax, ay, 36*s+u, u, w, false)
	// Mid bar
	vector.DrawFilledRect(screen, ax, ay+h*0.45, 36*s+u, u*0.9, w, false)
	// Angular cut on top-left corner (notch)
	vector.DrawFilledRect(screen, ax, ay, u*1.8, u*1.8, color.RGBA{0, 0, 0, 255}, false)

	// ── T ────────────────────────────────────────────────────────────
	tx := cx - 26*s
	ty := ay
	tw := float32(52) * s

	// Top bar
	vector.DrawFilledRect(screen, tx, ty, tw, u, w, false)
	// Stem
	vector.DrawFilledRect(screen, tx+tw/2-u/2, ty, u, h, w, false)

	// ── P ────────────────────────────────────────────────────────────
	px := cx + 38*s
	py := ay
	pw := float32(36) * s

	// Vertical stem
	vector.DrawFilledRect(screen, px, py, u, h, w, false)
	// Top bar
	vector.DrawFilledRect(screen, px, py, pw, u, w, false)
	// Mid bar
	vector.DrawFilledRect(screen, px, py+h*0.45, pw, u, w, false)
	// Right side of bowl (top half only)
	vector.DrawFilledRect(screen, px+pw, py, u, h*0.45+u, w, false)
	// Angular cut bottom-right of P bowl
	vector.DrawFilledRect(screen, px+pw, py+h*0.45, u*1.8, u*1.8,
		color.RGBA{0, 0, 0, 255}, false)
}

// ── Splash coordinator ────────────────────────────────────────────

// Splash holds either a GIF player or the code-drawn fallback.
type Splash struct {
	gif     *GIFPlayer
	atp     *ATPSplash
	Done    bool
}

func NewSplash() *Splash {
	s := &Splash{}
	if gp, err := LoadGIF("intro.gif"); err == nil {
		s.gif = gp
	} else {
		s.atp = &ATPSplash{}
	}
	return s
}

func (s *Splash) Update() {
	if s.Done {
		return
	}
	if s.gif != nil {
		s.gif.Update()
		if s.gif.Done {
			s.Done = true
		}
	} else {
		s.atp.Update()
		if s.atp.Done {
			s.Done = true
		}
	}
}

func (s *Splash) Draw(screen *ebiten.Image) {
	if s.gif != nil {
		s.gif.Draw(screen)
	} else if s.atp != nil {
		s.atp.Draw(screen)
	}
}

// Skip immediately finishes the splash (on tap).
func (s *Splash) Skip() {
	s.Done = true
	if s.gif != nil {
		s.gif.Done = true
	}
	if s.atp != nil {
		s.atp.Done = true
	}
}
