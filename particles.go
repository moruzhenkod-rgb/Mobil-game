package main

import (
	"image/color"
	"math"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

type Particle struct {
	x, y   float32
	vx, vy float32
	life   float32 // 1→0
	decay  float32
	r      float32
	color  color.RGBA
	square bool // draw as square instead of circle
}

type ParticleSystem struct {
	particles []Particle
}

// Emit spawns count random-direction particles at (x, y).
func (ps *ParticleSystem) Emit(x, y float32, clr color.RGBA, count int) {
	for i := 0; i < count; i++ {
		angle := rand.Float64() * math.Pi * 2
		speed := float32(rand.Float64()*3 + 1)
		ps.particles = append(ps.particles, Particle{
			x: x, y: y,
			vx:    float32(math.Cos(angle)) * speed,
			vy:    float32(math.Sin(angle)) * speed,
			life:  1.0,
			decay: float32(rand.Float64()*0.03 + 0.02),
			r:     float32(rand.Float64()*4 + 2),
			color: clr,
		})
	}
}

// EmitBurst fires a full burst + sparkle ring (used for gem matches).
func (ps *ParticleSystem) EmitBurst(x, y float32, clr color.RGBA) {
	ps.Emit(x, y, clr, 14)
	// Sparkle ring
	for i := 0; i < 8; i++ {
		angle := float64(i) / 8 * math.Pi * 2
		px := x + float32(math.Cos(angle))*22
		py := y + float32(math.Sin(angle))*22
		ps.Emit(px, py, color.RGBA{255, 255, 200, 220}, 2)
	}
}

// EmitTrail emits a few fading particles used for swap animation trails.
func (ps *ParticleSystem) EmitTrail(x, y float32, clr color.RGBA) {
	for i := 0; i < 4; i++ {
		speed := float32(rand.Float64()*1.5 + 0.5)
		angle := rand.Float64() * math.Pi * 2
		ps.particles = append(ps.particles, Particle{
			x: x + (rand.Float32()-0.5)*8,
			y: y + (rand.Float32()-0.5)*8,
			vx:    float32(math.Cos(angle)) * speed,
			vy:    float32(math.Sin(angle)) * speed,
			life:  1.0,
			decay: float32(rand.Float64()*0.06 + 0.05),
			r:     float32(rand.Float64()*3 + 1),
			color: clr,
		})
	}
}

// EmitShockwave fires a fast-expanding ring of square particles (bomb explosions).
func (ps *ParticleSystem) EmitShockwave(x, y float32, clr color.RGBA) {
	rings := 16
	for i := 0; i < rings; i++ {
		angle := float64(i) / float64(rings) * math.Pi * 2
		speed := float32(rand.Float64()*5 + 3)
		ps.particles = append(ps.particles, Particle{
			x: x, y: y,
			vx:    float32(math.Cos(angle)) * speed,
			vy:    float32(math.Sin(angle)) * speed,
			life:  1.0,
			decay: float32(rand.Float64()*0.04 + 0.03),
			r:     float32(rand.Float64()*5 + 3),
			color: clr,
			square: true,
		})
	}
	// Inner burst
	ps.Emit(x, y, clr, 10)
}

// EmitStarRain drops glittering star-like particles from the top (win screen).
func (ps *ParticleSystem) EmitStarRain(screenW float32) {
	starColors := []color.RGBA{
		{255, 220, 60, 220}, {255, 100, 180, 200},
		{100, 200, 255, 200}, {200, 255, 100, 200},
		{255, 160, 60, 200},
	}
	for i := 0; i < 3; i++ {
		clr := starColors[rand.Intn(len(starColors))]
		ps.particles = append(ps.particles, Particle{
			x:     rand.Float32() * screenW,
			y:     -4,
			vx:    (rand.Float32()-0.5) * 1.5,
			vy:    rand.Float32()*2 + 1.5,
			life:  1.0,
			decay: float32(rand.Float64()*0.008 + 0.006),
			r:     float32(rand.Float64()*5 + 3),
			color: clr,
		})
	}
}

func (ps *ParticleSystem) Update() {
	alive := ps.particles[:0]
	for i := range ps.particles {
		p := &ps.particles[i]
		p.x += p.vx
		p.y += p.vy
		p.vy += 0.12 // gravity
		p.vx *= 0.96 // drag
		p.life -= p.decay
		if p.life > 0 {
			alive = append(alive, *p)
		}
	}
	ps.particles = alive
}

func (ps *ParticleSystem) Draw(screen *ebiten.Image) {
	for _, p := range ps.particles {
		c := p.color
		c.A = uint8(float32(c.A) * p.life)
		if p.square {
			s := p.r * p.life
			vector.DrawFilledRect(screen, p.x-s/2, p.y-s/2, s, s, c, false)
		} else {
			vector.DrawFilledCircle(screen, p.x, p.y, p.r*p.life, c, false)
		}
	}
}
