package main

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// ── Menu background GIF (animated, looping) ──────────────────────

var menuBgGIF *GIFPlayer
var menuBgLoaded bool

func ensureMenuBg() {
	if menuBgLoaded {
		return
	}
	menuBgLoaded = true
	gp, err := LoadGIF("home.gif")
	if err != nil || gp == nil || len(gp.frames) == 0 {
		return
	}
	menuBgGIF = gp
}

// UpdateMenuBg advances the looping background animation.
func UpdateMenuBg() {
	if menuBgGIF != nil {
		menuBgGIF.UpdateLoop()
	}
}

// ── Main Menu ─────────────────────────────────────────────────────

func DrawMainMenu(screen *ebiten.Image, tick int) {
	ensureMenuBg()
	ensureUIAssets()

	if menuBgGIF != nil {
		menuBgGIF.Draw(screen)
	} else {
		screen.Fill(color.RGBA{8, 5, 22, 255})
	}

	drawMenuHUDOverlay(screen)
}

// drawBtnText draws localised text centred at (cx, cy) with a strong drop shadow
// so it reads clearly on any GIF background. No image backgrounds are drawn.
func drawBtnText(screen *ebiten.Image, label string, cx, cy float64, large bool) {
	face := FontBtn
	if large {
		face = FontXL
	}
	if face == nil || label == "" {
		return
	}
	// Multi-layer shadow for readability
	for _, d := range [][2]float64{{-1, -1}, {1, -1}, {-1, 1}, {1, 1}, {0, 2}} {
		FCenter(screen, label, cx+d[0]*2, cy+d[1]*2, face, color.RGBA{0, 0, 0, 160})
	}
	FCenter(screen, label, cx, cy, face, ColWhite)
}

func drawMenuButtonsOverlay(screen *ebiten.Image) {
	// PLAY — centred, y≈329
	drawBtnText(screen, T("play"), float64(ScreenW)/2, 329, true)

	// Row 1 — ADVENTURE | LEVELS
	drawBtnText(screen, T("adventure"), 135, 651, false)
	drawBtnText(screen, T("levels"),    405, 651, false)

	// Row 2 — SHOP | COLLECTION
	drawBtnText(screen, T("shop"),       135, 747, false)
	drawBtnText(screen, T("collection"), 405, 747, false)

	// Row 3 — QUESTS | SETTINGS | MESSAGES
	drawBtnText(screen, T("quests"),   92,  841, false)
	drawBtnText(screen, T("settings"), 272, 841, false)
	drawBtnText(screen, T("messages"), 452, 841, false)
}

// drawMenuHUDOverlay renders the top player-info bar.
func drawMenuHUDOverlay(screen *ebiten.Image) {
	const barH = 76

	// ── Background: two-layer gradient simulation ─────────────────
	vector.DrawFilledRect(screen, 0, 0, ScreenW, barH, color.RGBA{0, 0, 0, 200}, false)
	vector.DrawFilledRect(screen, 0, 0, ScreenW, barH/2, color.RGBA{20, 12, 40, 60}, false)

	// Gold top line + bottom line
	vector.DrawFilledRect(screen, 0, 0, ScreenW, 2, color.RGBA{200, 162, 28, 200}, false)
	vector.DrawFilledRect(screen, 0, barH-2, ScreenW, 2, color.RGBA{200, 162, 28, 160}, false)

	// ── Avatar ────────────────────────────────────────────────────
	const avD = 60
	const avX, avY = 8, 8
	const avCX, avCY = avX + avD/2, avY + avD/2

	// Outer glow ring
	vector.DrawFilledCircle(screen, avCX, avCY, avD/2+4, color.RGBA{200, 162, 28, 80}, false)
	// Gold ring
	vector.DrawFilledCircle(screen, avCX, avCY, avD/2+2, color.RGBA{200, 162, 28, 255}, false)
	// Avatar image or fallback
	if imgAvatarRound != nil {
		drawImgAt(screen, imgAvatarRound, avX+1, avY+1)
	} else {
		vector.DrawFilledCircle(screen, avCX, avCY, avD/2, color.RGBA{30, 60, 140, 255}, false)
		nm := progress.UserName
		if nm == "" {
			nm = "P"
		}
		if FontLg != nil {
			FCenter(screen, string([]rune(nm)[0:1]), avCX, float64(avCY)-16, FontLg, ColWhite)
		}
	}

	// ── Name + Level badge ────────────────────────────────────────
	name := progress.UserName
	if name == "" {
		name = "PLAYER"
	}
	lv := progress.UnlockedLevel
	if lv > MaxLevels {
		lv = MaxLevels
	}
	tx := float64(avX + avD + 10)

	if FontBtn != nil {
		// Name shadow
		FDraw(screen, name, tx+1, 10, FontBtn, color.RGBA{0, 0, 0, 180})
		FDraw(screen, name, tx, 9, FontBtn, ColWhite)
	}

	// Level badge pill
	lvStr := fmt.Sprintf("Lv %d", lv)
	vector.DrawFilledRect(screen, float32(tx)-2, 30, 52, 18, color.RGBA{160, 120, 10, 220}, false)
	vector.DrawFilledRect(screen, float32(tx)-2, 30, 52, 2, color.RGBA{255, 215, 50, 255}, false)
	if FontSm != nil {
		FDraw(screen, lvStr, tx+2, 32, FontSm, color.RGBA{255, 240, 180, 255})
	}

	// ── Vertical divider ──────────────────────────────────────────
	divX := float32(tx + 82)
	vector.DrawFilledRect(screen, divX, 10, 1, barH-20, color.RGBA{200, 162, 28, 100}, false)

	// ── Coins ─────────────────────────────────────────────────────
	icX := float32(tx + 90)
	// Gold coin circle
	vector.DrawFilledCircle(screen, icX+10, 24, 11, color.RGBA{180, 140, 10, 255}, false)
	vector.DrawFilledCircle(screen, icX+10, 24, 8, color.RGBA{255, 210, 40, 255}, false)
	vector.DrawFilledCircle(screen, icX+8, 22, 3, color.RGBA{255, 250, 180, 200}, false)
	if FontBtn != nil {
		FDraw(screen, fmt.Sprintf("%d", progress.Coins), float64(icX+24), 16, FontBtn, ColGold)
	}

	// ── Gems ──────────────────────────────────────────────────────
	gemIconX := float32(tx + 90)
	gx, gy := gemIconX, float32(42)
	// Diamond shape
	vector.DrawFilledRect(screen, gx+4, gy, 12, 5, color.RGBA{60, 160, 255, 255}, false)
	vector.DrawFilledRect(screen, gx, gy+4, 20, 10, color.RGBA{40, 130, 240, 255}, false)
	vector.DrawFilledRect(screen, gx+4, gy+13, 12, 5, color.RGBA{30, 100, 200, 255}, false)
	vector.DrawFilledRect(screen, gx+7, gy+17, 6, 4, color.RGBA{20, 80, 170, 255}, false)
	// Shine
	vector.DrawFilledRect(screen, gx+5, gy+1, 5, 4, color.RGBA{180, 230, 255, 180}, false)
	if FontBtn != nil {
		FDraw(screen, fmt.Sprintf("%d", progress.Gems), float64(gx+24), 46, FontBtn,
			color.RGBA{130, 210, 255, 255})
	}

	// ── Stars (right side) ────────────────────────────────────────
	if FontSm != nil {
		stars := fmt.Sprintf("★ %d", totalStars())
		FDraw(screen, stars, float64(ScreenW)-70, 12, FontSm, color.RGBA{255, 215, 60, 230})
	}

	// ── Settings button (top-right) ───────────────────────────────
	const gearX, gearY = float32(ScreenW - 34), float32(8)
	const gearS = float32(28)
	// Gear background
	vector.DrawFilledCircle(screen, gearX+gearS/2, gearY+gearS/2, gearS/2+2,
		color.RGBA{160, 128, 18, 220}, false)
	vector.DrawFilledCircle(screen, gearX+gearS/2, gearY+gearS/2, gearS/2,
		color.RGBA{30, 24, 60, 240}, false)
	// Gear teeth (4 directions)
	gcx, gcy := gearX+gearS/2, gearY+gearS/2
	vector.DrawFilledCircle(screen, gcx, gcy, 6, color.RGBA{180, 175, 210, 255}, false)
	vector.DrawFilledCircle(screen, gcx, gcy, 3, color.RGBA{30, 24, 60, 255}, false)
	for _, d := range [][2]float32{{0, -1}, {0, 1}, {-1, 0}, {1, 0}} {
		vector.DrawFilledRect(screen, gcx+d[0]*9-2, gcy+d[1]*9-2, 5, 5,
			color.RGBA{180, 175, 210, 255}, false)
	}
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
			stars := biomeStars(b)
			maxStars := LevelsPerBiome * 3
			ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Stars: %d/%d", stars, maxStars), 46, int(by)+54)
			lastLevel := b * LevelsPerBiome
			ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Levels %d - %d", firstLevel, lastLevel), 46, int(by)+74)
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
			for s := 0; s < 3; s++ {
				sc := color.RGBA{60, 50, 80, 255}
				if s < stars {
					sc = color.RGBA{255, 210, 50, 255}
				}
				sx := float32(bx + 16 + s*28)
				sy := float32(by + 42)
				drawStarSmall(screen, sx, sy, 10, sc)
			}
			lvl := GetLevel(n)
			drawDifficulty(screen, lvl, bx+8, by+72)
		} else {
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
	vector.DrawFilledRect(screen, 0, 0, ScreenW, ScreenH,
		color.RGBA{uint8(r * 0.1), uint8(g * 0.1), uint8(b * 0.1), 255}, false)
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

// ── Shared button helper ──────────────────────────────────────────

func drawMenuButton(screen *ebiten.Image, label string, x, y, w, h int, bg, highlight color.RGBA, tick int) {
	fx, fy, fw, fh := float32(x), float32(y), float32(w), float32(h)
	vector.DrawFilledRect(screen, fx+3, fy+5, fw, fh, color.RGBA{0, 0, 0, 100}, false)
	vector.DrawFilledRect(screen, fx, fy, fw, fh, bg, false)
	vector.DrawFilledRect(screen, fx, fy, fw, 3, highlight, false)
	vector.DrawFilledRect(screen, fx, fy, fw, fh*0.4, color.RGBA{255, 255, 255, 20}, false)
	dark := color.RGBA{bg.R / 3, bg.G / 3, bg.B / 3, 255}
	vector.DrawFilledRect(screen, fx, fy+fh-5, fw, 5, dark, false)
	if FontBtn != nil {
		FCenter(screen, label, float64(fx+fw/2), float64(fy)+float64(fh)/2-10, FontBtn, ColWhite)
	}
}
