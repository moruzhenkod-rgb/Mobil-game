package main

import (
	"fmt"
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// ── Main Menu ─────────────────────────────────────────────────────

func DrawMainMenu(screen *ebiten.Image, tick int) {
	// Animated title gems
	gemX := []float32{80, 160, 240, 320, 400, 460}
	gemKinds := []TileKind{TileRed, TileBlue, TileGreen, TileYellow, TilePurple, TileOrange}
	for i, gx := range gemX {
		bob := float32(math.Sin(float64(tick)*0.05+float64(i)*0.7)) * 8
		gy := float32(130) + bob
		c := gemColors[gemKinds[i]]
		dc := gemDark[gemKinds[i]]
		vector.DrawFilledRect(screen, gx, gy, 44, 44, c, false)
		vector.DrawFilledRect(screen, gx, gy, 44, 4, color.RGBA{255, 255, 255, 80}, false)
		vector.DrawFilledRect(screen, gx, gy+40, 44, 4, dc, false)
		vector.DrawFilledCircle(screen, gx+10, gy+10, 6, color.RGBA{255, 255, 255, 100}, false)
	}

	// Title
	if FontXL != nil {
		FCenter(screen, "RUNIC CRUSH", float64(ScreenW)/2, 186, FontXL, ColGold)
		FCenter(screen, "~ Match Gems, Break Curses ~", float64(ScreenW)/2, 248, FontSm, ColPurple)
	} else {
		title := "RUNIC  CRUSH"
		ebitenutil.DebugPrintAt(screen, title, ScreenW/2-len(title)*3, 200)
	}

	// Glow line under title
	glow := uint8(160 + 80*math.Sin(float64(tick)*0.04))
	vector.DrawFilledRect(screen, 80, 238, ScreenW-160, 2, color.RGBA{160, 100, 255, glow}, false)

	// PLAY button
	drawMenuButton(screen, "PLAY", ScreenW/2-100, 340, 200, 56, color.RGBA{50, 170, 70, 255}, color.RGBA{80, 220, 100, 200}, tick)

	// Continue button (if progress exists)
	if progress.UnlockedLevel > 1 {
		label := fmt.Sprintf("CONTINUE (Lv.%d)", progress.UnlockedLevel)
		drawMenuButton(screen, label, ScreenW/2-120, 420, 240, 50, color.RGBA{60, 100, 200, 255}, color.RGBA{90, 140, 255, 200}, 0)
	}

	// Settings / Shop
	drawMenuButton(screen, "SETTINGS", ScreenW/2-80, 490, 160, 44, color.RGBA{50, 35, 90, 255}, color.RGBA{100, 70, 160, 200}, 0)
	drawMenuButton(screen, "SHOP", ScreenW/2-80, 548, 160, 44, color.RGBA{160, 100, 20, 255}, color.RGBA{220, 160, 40, 200}, 0)

	// Version / legal
	ebitenutil.DebugPrintAt(screen, "v0.5.0  |  Privacy Policy  |  Terms of Service", 60, ScreenH-30)

	// Stars count
	total := totalStars()
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Stars: %d / %d", total, MaxLevels*3), ScreenW/2-54, ScreenH-55)
}

func drawMenuButton(screen *ebiten.Image, label string, x, y, w, h int, bg, border color.RGBA, tick int) {
	pulse := float32(0)
	if tick > 0 {
		pulse = float32(math.Sin(float64(tick)*0.05)) * 2
	}
	vector.DrawFilledRect(screen, float32(x)-pulse, float32(y)-pulse, float32(w)+pulse*2, float32(h)+pulse*2, bg, false)
	vector.DrawFilledRect(screen, float32(x), float32(y), float32(w), 3, border, false)
	ebitenutil.DebugPrintAt(screen, label, x+w/2-len(label)*3, y+h/2-5)
}

func totalStars() int {
	total := 0
	for _, s := range progress.Stars {
		total += s
	}
	return total
}

// ── Chapter Select ────────────────────────────────────────────────

func DrawChapterSelect(screen *ebiten.Image, tick int) {
	ebitenutil.DebugPrintAt(screen, "SELECT CHAPTER", ScreenW/2-42, 30)
	vector.DrawFilledRect(screen, 60, 48, ScreenW-120, 2, color.RGBA{120, 60, 200, 180}, false)

	// Back button
	drawMenuButton(screen, "< BACK", 20, 20, 80, 28, color.RGBA{40, 25, 70, 220}, color.RGBA{100, 60, 180, 180}, 0)

	for b := 1; b <= BiomeCount; b++ {
		info := biomes[b]
		firstLevel := (b-1)*LevelsPerBiome + 1
		unlocked := isLevelUnlocked(firstLevel)

		by := float32(80 + (b-1)*148)
		bc := color.RGBA{info.Color[0] / 3, info.Color[1] / 3, info.Color[2] / 3, 220}
		if !unlocked {
			bc = color.RGBA{20, 15, 35, 220}
		}
		vector.DrawFilledRect(screen, 30, by, ScreenW-60, 136, bc, false)
		bdc := color.RGBA{info.Color[0], info.Color[1], info.Color[2], 200}
		if !unlocked {
			bdc = color.RGBA{60, 50, 80, 180}
		}
		vector.DrawFilledRect(screen, 30, by, ScreenW-60, 3, bdc, false)

		if unlocked {
			ebitenutil.DebugPrintAt(screen, fmt.Sprintf("BIOME %d: %s", b, info.Name), 46, int(by)+14)
			ebitenutil.DebugPrintAt(screen, info.Description, 46, int(by)+34)
			// Stars for this biome
			stars := biomeStars(b)
			maxStars := LevelsPerBiome * 3
			ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Stars: %d/%d", stars, maxStars), 46, int(by)+54)
			// Level range
			lastLevel := b * LevelsPerBiome
			ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Levels %d - %d", firstLevel, lastLevel), 46, int(by)+74)
			// Progress bar
			pct := float32(biomeCompleted(b)) / LevelsPerBiome
			vector.DrawFilledRect(screen, 46, by+96, float32(ScreenW-92), 10, color.RGBA{30, 20, 55, 255}, false)
			vector.DrawFilledRect(screen, 46, by+96, float32(ScreenW-92)*pct, 10, bdc, false)
		} else {
			ebitenutil.DebugPrintAt(screen, fmt.Sprintf("BIOME %d: [LOCKED]", b), 46, int(by)+14)
			ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Complete level %d to unlock", firstLevel-1), 46, int(by)+34)
		}
	}
}

func biomeStars(b int) int {
	total := 0
	start := (b-1)*LevelsPerBiome + 1
	end := b * LevelsPerBiome
	for i := start; i <= end; i++ {
		total += progress.Stars[i]
	}
	return total
}

func biomeCompleted(b int) int {
	count := 0
	start := (b-1)*LevelsPerBiome + 1
	end := b * LevelsPerBiome
	for i := start; i <= end; i++ {
		if progress.Stars[i] > 0 {
			count++
		}
	}
	return count
}

// ChapterSelectBounds returns the y-range for each biome card (for hit testing).
func ChapterSelectBounds(b int) (int, int) {
	y := 80 + (b-1)*148
	return y, y + 136
}

// ── Level Select ──────────────────────────────────────────────────

func DrawLevelSelect(screen *ebiten.Image, biome, hoverLevel, tick int) {
	info := biomes[biome]
	title := fmt.Sprintf("BIOME %d: %s", biome, info.Name)
	ebitenutil.DebugPrintAt(screen, title, ScreenW/2-len(title)*3, 30)
	vector.DrawFilledRect(screen, 60, 48, ScreenW-120, 2,
		color.RGBA{info.Color[0], info.Color[1], info.Color[2], 180}, false)

	drawMenuButton(screen, "< BACK", 20, 20, 80, 28, color.RGBA{40, 25, 70, 220}, color.RGBA{100, 60, 180, 180}, 0)

	// 4 columns × 5 rows = 20 levels
	const cols = 4
	bw, bh := 110, 110
	padX := (ScreenW - cols*bw - (cols-1)*8) / 2
	padY := 65

	start := (biome-1)*LevelsPerBiome + 1
	for i := 0; i < LevelsPerBiome; i++ {
		n := start + i
		row := i / cols
		col := i % cols
		bx := padX + col*(bw+8)
		by := padY + row*(bh+8)

		unlocked := isLevelUnlocked(n)
		stars := progress.Stars[n]

		// Card background
		bg := color.RGBA{20, 12, 42, 220}
		if !unlocked {
			bg = color.RGBA{15, 10, 28, 200}
		} else if n == hoverLevel {
			bg = color.RGBA{50, 30, 90, 220}
		}
		vector.DrawFilledRect(screen, float32(bx), float32(by), float32(bw), float32(bh), bg, false)

		topC := color.RGBA{info.Color[0], info.Color[1], info.Color[2], 180}
		if !unlocked {
			topC = color.RGBA{50, 40, 70, 160}
		}
		vector.DrawFilledRect(screen, float32(bx), float32(by), float32(bw), 3, topC, false)

		if unlocked {
			ebitenutil.DebugPrintAt(screen, fmt.Sprintf("%d", n), bx+bw/2-len(fmt.Sprintf("%d", n))*3, by+18)
			// Stars
			for s := 0; s < 3; s++ {
				sc := color.RGBA{60, 50, 80, 255}
				if s < stars {
					sc = color.RGBA{255, 210, 50, 255}
				}
				sx := float32(bx + 16 + s*28)
				sy := float32(by + 42)
				drawStarSmall(screen, sx, sy, 10, sc)
			}
			// Difficulty dots
			lvl := GetLevel(n)
			drawDifficulty(screen, lvl, bx+8, by+72)
		} else {
			// Lock icon
			vector.DrawFilledRect(screen, float32(bx+bw/2-8), float32(by+24), 16, 20, color.RGBA{80, 70, 100, 255}, false)
			vector.DrawFilledCircle(screen, float32(bx+bw/2), float32(by+22), 10, color.RGBA{80, 70, 100, 255}, false)
			vector.DrawFilledCircle(screen, float32(bx+bw/2), float32(by+22), 6, color.RGBA{20, 15, 35, 255}, false)
		}
	}
}

func drawStarSmall(screen *ebiten.Image, cx, cy, r float32, c color.RGBA) {
	vector.DrawFilledRect(screen, cx-r, cy-r*0.3, r*2, r*0.6, c, false)
	vector.DrawFilledRect(screen, cx-r*0.3, cy-r, r*0.6, r*2, c, false)
}

func drawDifficulty(screen *ebiten.Image, l Level, x, y int) {
	// Obstacle icons (tiny coloured dots)
	ox := x
	if l.IceCount > 0 {
		vector.DrawFilledRect(screen, float32(ox), float32(y), 10, 10, color.RGBA{140, 210, 255, 220}, false)
		ox += 14
	}
	if l.StoneCount > 0 {
		vector.DrawFilledRect(screen, float32(ox), float32(y), 10, 10, color.RGBA{110, 105, 120, 220}, false)
		ox += 14
	}
	if l.ChainCount > 0 {
		vector.DrawFilledRect(screen, float32(ox), float32(y), 10, 10, color.RGBA{200, 170, 80, 220}, false)
	}
}

// DrawBiomeScene draws an animated backdrop for biome intro dialogues.
func DrawBiomeScene(screen *ebiten.Image, biome, tick int) {
	info := biomes[biome]
	r, g, b := float32(info.Color[0]), float32(info.Color[1]), float32(info.Color[2])
	// Gradient sky
	vector.DrawFilledRect(screen, 0, 0, ScreenW, ScreenH,
		color.RGBA{uint8(r * 0.1), uint8(g * 0.1), uint8(b * 0.1), 255}, false)

	switch biome {
	case 1: // Forest Ruins — floating leaves and tree silhouettes
		for i := 0; i < 6; i++ {
			tx := float32(80+i*66) + float32(math.Sin(float64(tick)*0.02+float64(i)))*8
			vector.DrawFilledRect(screen, tx, ScreenH-200, 30, 200, color.RGBA{30, 80, 30, 200}, false)
			vector.DrawFilledCircle(screen, tx+15, ScreenH-210, 40, color.RGBA{40, 100, 40, 200}, false)
		}
		for i := 0; i < 12; i++ {
			lx := float32((tick*2+i*44)%ScreenW)
			ly := float32(100+i*70) + float32(math.Sin(float64(tick)*0.04+float64(i)))*20
			vector.DrawFilledRect(screen, lx, ly, 8, 14, color.RGBA{60, 140, 60, 180}, false)
		}
	case 2: // Crystal Caves — icicles and snowflakes
		for i := 0; i < 14; i++ {
			ix := float32(20 + i*38)
			ih := float32(40 + (i*17)%90)
			vector.DrawFilledRect(screen, ix, 0, 14, ih, color.RGBA{140, 200, 255, 200}, false)
		}
		for i := 0; i < 10; i++ {
			fx := float32((tick+i*52)%ScreenW)
			fy := float32((tick*2+i*80)%ScreenH)
			vector.DrawFilledCircle(screen, fx, fy, 4, color.RGBA{200, 230, 255, 160}, false)
		}
	case 3: // Volcanic Forge — lava rocks and ember particles
		for i := 0; i < 8; i++ {
			rx := float32(30+i*66) + float32(math.Sin(float64(i)*1.3))*15
			vector.DrawFilledRect(screen, rx, ScreenH-140, 50, 140, color.RGBA{80, 50, 30, 220}, false)
			glow := uint8(100 + 80*math.Sin(float64(tick)*0.06+float64(i)))
			vector.DrawFilledCircle(screen, rx+25, ScreenH-140, 20, color.RGBA{220, glow, 20, 180}, false)
		}
		for i := 0; i < 8; i++ {
			ex := float32((tick*3+i*68)%ScreenW)
			ey := ScreenH - float32((tick*2+i*110)%int(ScreenH))
			vector.DrawFilledCircle(screen, ex, ey, 3, color.RGBA{255, 140, 20, 200}, false)
		}
	case 4: // Arcane Library — floating books and chain links
		for i := 0; i < 5; i++ {
			by := float32(100+i*140) + float32(math.Sin(float64(tick)*0.04+float64(i)*0.8))*12
			vector.DrawFilledRect(screen, float32(60+i*80), by, 40, 55, color.RGBA{80, 50, 130, 200}, false)
			vector.DrawFilledRect(screen, float32(60+i*80), by, 6, 55, color.RGBA{100, 70, 160, 220}, false)
		}
		for i := 0; i < 6; i++ {
			cx := float32(40+i*84) + float32(math.Sin(float64(tick)*0.03+float64(i)))*10
			cy := float32(ScreenH/2) + float32(math.Cos(float64(tick)*0.03+float64(i)*0.5))*60
			vector.DrawFilledCircle(screen, cx, cy, 10, color.RGBA{160, 130, 200, 160}, false)
			vector.DrawFilledCircle(screen, cx, cy, 6, color.RGBA{20, 10, 40, 200}, false)
		}
	case 5: // Sky Citadel — clouds and lightning flashes
		for i := 0; i < 5; i++ {
			clx := float32((tick+i*110)%ScreenW) - 50
			cly := float32(80 + i*70)
			vector.DrawFilledRect(screen, clx, cly, 120, 40, color.RGBA{200, 210, 255, 120}, false)
			vector.DrawFilledCircle(screen, clx+30, cly, 30, color.RGBA{200, 210, 255, 120}, false)
			vector.DrawFilledCircle(screen, clx+70, cly-10, 25, color.RGBA{200, 210, 255, 120}, false)
		}
		// Lightning flash occasionally
		if tick%120 < 3 {
			vector.DrawFilledRect(screen, ScreenW/2-2, 0, 4, ScreenH/2, color.RGBA{255, 255, 200, 200}, false)
		}
	}
}

// LevelSelectHit returns the level number clicked, or 0.
func LevelSelectHit(biome, mx, my int) int {
	const cols = 4
	bw, bh := 110, 110
	padX := (ScreenW - cols*bw - (cols-1)*8) / 2
	padY := 65

	start := (biome-1)*LevelsPerBiome + 1
	for i := 0; i < LevelsPerBiome; i++ {
		n := start + i
		row := i / cols
		col := i % cols
		bx := padX + col*(bw+8)
		by := padY + row*(bh+8)
		if mx >= bx && mx <= bx+bw && my >= by && my <= by+bh {
			return n
		}
	}
	return 0
}
