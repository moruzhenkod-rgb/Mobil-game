package main

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// ── Game ──────────────────────────────────────────────────────────

type Game struct {
	state     GameState
	prevState GameState // for back-navigation

	// Playing
	level      Level
	levelNum   int
	board      *Board
	char       *Character
	ps         ParticleSystem

	score      int
	movesLeft  int
	cleared    [GemCount]int
	clearedAny int

	// Chapter/Level select
	selectedBiome int
	hoverLevel    int

	// Dialogue
	dlg           *DialoguePlayer
	afterDlg      GameState
	afterDlgLevel int

	// Biome intro
	biomeIntroN int

	// Win/Lose
	winTick int

	tick  int
	stars []starDot

	// Stage 5: visual effects
	offscreen *ebiten.Image
	shake     screenShake

	// prevent replaying one-shot sounds
	winSoundPlayed  bool
	loseSoundPlayed bool
}

type starDot struct {
	x, y  float32
	r, speed float32
	alpha uint8
	dir  int8
}

// ── Init ──────────────────────────────────────────────────────────

func NewGame() *Game {
	loadProgress()
	InitFirebase()
	InitAudio()
	InitFonts()
	ads.LoadRewardedAd()
	g := &Game{
		state:    StateAgeGate,
		levelNum: 1,
		stars:    makeStars(90),
	}
	if progress.AgeVerified {
		g.state = StateMainMenu
	}
	return g
}

func makeStars(n int) []starDot {
	s := make([]starDot, n)
	for i := range s {
		s[i] = starDot{
			x: float32(randFloat(ScreenW)), y: float32(randFloat(ScreenH)),
			r: float32(randFloat(1.5) + 0.5), speed: float32(randFloat(0.2) + 0.05),
			alpha: uint8(randIntn(140) + 80), dir: int8(randIntn(2)*2 - 1),
		}
	}
	return s
}

// ── navigation helpers ────────────────────────────────────────────

func (g *Game) goToMainMenu() {
	g.state = StateMainMenu
}

func (g *Game) goToChapterSelect() {
	g.prevState = g.state
	g.state = StateChapterSelect
}

func (g *Game) goToLevelSelect(biome int) {
	g.selectedBiome = biome
	g.prevState = g.state
	g.state = StateLevelSelect
}

func (g *Game) startLevel(n int) {
	g.levelNum = n
	g.level = GetLevel(n)
	g.board = NewBoard(g.level)
	g.char = NewCharacter()
	g.ps = ParticleSystem{}
	g.score = 0
	g.movesLeft = g.level.MaxMoves
	g.cleared = [GemCount]int{}
	g.clearedAny = 0
	g.winTick = 0
	g.winSoundPlayed = false
	g.loseSoundPlayed = false
	g.state = StatePlaying
	LogLevelStart(n)
}

func (g *Game) launchLevel(n int) {
	if sceneTransition.active {
		return // prevent double-trigger during fade
	}
	biome := BiomeOf(n)
	isFirstInBiome := (n-1)%LevelsPerBiome == 0

	doStart := func() {
		// Check for dialogue
		if lines, ok := dialogueTable[n]; ok {
			g.dlg = NewDialogue(lines)
			g.afterDlg = StatePlaying
			g.afterDlgLevel = n
			g.biomeIntroN = 0
			g.state = StateDialogue
			FadeIn(14, nil)
			return
		}
		// Biome intro for first level of new biome
		if isFirstInBiome && biome >= 2 && !isLevelUnlocked(n) {
			intro := biomeIntroTable[biome]
			if len(intro) > 0 {
				g.dlg = NewDialogue(intro)
				g.afterDlg = StatePlaying
				g.afterDlgLevel = n
				g.biomeIntroN = biome
				g.state = StateDialogue
				FadeIn(14, nil)
				return
			}
		}
		g.startLevel(n)
		FadeIn(14, nil)
	}

	FadeOut(14, doStart)
}

// ── Update ────────────────────────────────────────────────────────

func (g *Game) Update() error {
	g.tick++
	g.updateStars()
	sceneTransition.Update()
	g.shake.Update()

	switch g.state {
	case StateAgeGate:
		g.updateAgeGate()
	case StateMainMenu:
		g.updateMainMenu()
	case StateChapterSelect:
		g.updateChapterSelect()
	case StateLevelSelect:
		g.updateLevelSelect()
	case StateDialogue:
		g.updateDialogue()
	case StatePlaying:
		g.updatePlaying()
	case StateWin:
		g.updateWin()
	case StateLose:
		g.updateLose()
	case StateSettings:
		g.updateSettings()
	case StateShop:
		g.updateShop()
	case StateLeaderboard:
		g.updateLeaderboard()
	}
	return nil
}

func (g *Game) updateStars() {
	for i := range g.stars {
		s := &g.stars[i]
		s.y += s.speed
		if s.y > ScreenH {
			s.y = 0
			s.x = float32(randFloat(ScreenW))
		}
		na := int(s.alpha) + int(s.dir)*2
		if na > 230 {
			na = 230
			s.dir = -1
		} else if na < 60 {
			na = 60
			s.dir = 1
		}
		s.alpha = uint8(na)
	}
}

func (g *Game) updateAgeGate() {
	mx, my, clicked := g.getClick()
	if !clicked {
		return
	}
	if mx >= 100 && mx <= 240 && my >= 420 && my <= 472 {
		progress.AgeVerified = true
		saveProgress()
		g.state = StateMainMenu
	}
}

func (g *Game) updateMainMenu() {
	mx, my, clicked := g.getClick()
	if !clicked {
		return
	}
	PlaySound(SndButton)
	// PLAY
	if mx >= ScreenW/2-100 && mx <= ScreenW/2+100 && my >= 340 && my <= 396 {
		g.goToChapterSelect()
		return
	}
	// CONTINUE
	if progress.UnlockedLevel > 1 && mx >= ScreenW/2-120 && mx <= ScreenW/2+120 && my >= 420 && my <= 470 {
		g.launchLevel(progress.UnlockedLevel)
		return
	}
	// SETTINGS
	if mx >= ScreenW/2-80 && mx <= ScreenW/2+80 && my >= 490 && my <= 534 {
		g.prevState = g.state
		g.state = StateSettings
		return
	}
	// SHOP
	if mx >= ScreenW/2-80 && mx <= ScreenW/2+80 && my >= 548 && my <= 592 {
		shopState.Reset()
		g.prevState = g.state
		g.state = StateShop
		return
	}
}

func (g *Game) updateChapterSelect() {
	mx, my, clicked := g.getClick()
	if !clicked {
		return
	}
	// Back
	if mx >= 20 && mx <= 100 && my >= 20 && my <= 48 {
		g.state = StateMainMenu
		return
	}
	for b := 1; b <= BiomeCount; b++ {
		y0, y1 := ChapterSelectBounds(b)
		if my >= y0 && my <= y1 {
			firstLevel := (b-1)*LevelsPerBiome + 1
			if isLevelUnlocked(firstLevel) {
				g.goToLevelSelect(b)
			}
			return
		}
	}
}

func (g *Game) updateLevelSelect() {
	mx, my, clicked := g.getClick()
	if !clicked {
		return
	}
	// Back
	if mx >= 20 && mx <= 100 && my >= 20 && my <= 48 {
		g.state = StateChapterSelect
		return
	}
	n := LevelSelectHit(g.selectedBiome, mx, my)
	if n > 0 && isLevelUnlocked(n) {
		g.launchLevel(n)
	}
}

func (g *Game) updateDialogue() {
	_, _, clicked := g.getClick()
	if g.dlg == nil {
		g.state = g.afterDlg
		if g.afterDlg == StatePlaying {
			g.startLevel(g.afterDlgLevel)
		}
		return
	}
	g.dlg.Update()
	if clicked {
		done := g.dlg.Tap()
		if done {
			g.dlg = nil
			if g.afterDlg == StatePlaying {
				g.startLevel(g.afterDlgLevel)
			} else {
				g.state = g.afterDlg
			}
		}
	}
}

func (g *Game) updatePlaying() {
	g.char.Update()
	g.ps.Update()

	gained, matched := g.board.Update()
	if gained > 0 {
		g.score += gained
		// Screen shake proportional to explosion size
		if gained >= 300 {
			g.shake.Trigger(7, 18) // bomb / large combo
		} else if gained >= 80 {
			g.shake.Trigger(3, 8)
		}
	}

	if matched {
		// Sync cleared counts
		for k := TileKind(TileRed); int(k) < int(GemCount); k++ {
			delta := g.board.ClearedByKind[k] - g.cleared[k]
			if delta > 0 {
				g.cleared[k] += delta
				g.clearedAny += delta
			}
		}
		g.char.SetState(CharCelebrate)
		g.emitMatchParticles()
		PlaySound(SndMatch)
	}

	if g.goalMet() {
		g.char.SetState(CharDance)
		markComplete(g.levelNum, g.score, g.level.MaxMoves, g.movesLeft)
		if !g.winSoundPlayed {
			PlaySound(SndWin)
			g.winSoundPlayed = true
		}
		LogLevelWin(g.levelNum, g.score)
		g.state = StateWin
		return
	}
	if g.movesLeft <= 0 && !g.board.IsBusy() {
		g.char.SetState(CharSad)
		if !g.loseSoundPlayed {
			PlaySound(SndLose)
			g.loseSoundPlayed = true
		}
		g.state = StateLose
		return
	}

	g.handleInput()
}

func (g *Game) emitMatchParticles() {
	for r := 0; r < Rows; r++ {
		for c := 0; c < Cols; c++ {
			t := g.board.Tiles(r, c)
			if t != nil && t.Exploding && t.ExplodeTick == 1 {
				tx, ty := tilePixelPos(r, c)
				clr := GemColor(t.Kind, t.BaseColor)
				g.ps.EmitBurst(float32(tx)+TileSize/2, float32(ty)+TileSize/2, clr)
			}
		}
	}
}

func (g *Game) handleInput() {
	mx, my, clicked := g.getClick()
	if clicked {
		if g.board.TrySelectPixel(mx, my) {
			g.movesLeft--
			PlaySound(SndSwap)
		}
	}
}

func (g *Game) updateWin() {
	g.winTick++
	g.ps.Update()
	g.char.Update()

	// Star rain + confetti
	g.ps.EmitStarRain(ScreenW)
	if g.winTick%2 == 0 {
		cols := []color.RGBA{
			{255, 80, 80, 220}, {80, 200, 80, 220},
			{80, 120, 255, 220}, {255, 200, 40, 220},
			{200, 80, 255, 220}, {255, 160, 40, 220},
		}
		g.ps.Emit(float32(randFloat(ScreenW)), 0, cols[randIntn(len(cols))], 2)
	}

	mx, my, clicked := g.getClick()
	if !clicked {
		return
	}
	// Next Level
	if mx >= 160 && mx <= 380 && my >= 610 && my <= 665 {
		next := g.levelNum + 1
		if next > MaxLevels {
			g.goToMainMenu()
		} else {
			g.launchLevel(next)
		}
		return
	}
	// Level Select
	if mx >= 160 && mx <= 380 && my >= 676 && my <= 720 {
		g.goToLevelSelect(BiomeOf(g.levelNum))
	}
}

func (g *Game) updateLose() {
	g.char.Update()
	mx, my, clicked := g.getClick()
	if !clicked {
		return
	}
	// Retry
	if mx >= 160 && mx <= 380 && my >= 560 && my <= 612 {
		g.launchLevel(g.levelNum)
		return
	}
	// Watch Ad for +3 moves
	if mx >= 160 && mx <= 380 && my >= 626 && my <= 671 && ads.IsReady() {
		ads.ShowRewardedAd(func() {
			g.movesLeft += 3
			g.state = StatePlaying
		})
		return
	}
	// Level Select
	if mx >= 160 && mx <= 380 && my >= 682 && my <= 727 {
		g.goToLevelSelect(BiomeOf(g.levelNum))
	}
}

func (g *Game) updateSettings() {
	mx, my, clicked := g.getClick()
	if !clicked {
		return
	}
	if pinDlg.IsOpen() {
		pinDlg.HandleClick(mx, my)
		return
	}
	if SettingsHandleClick(mx, my) == "back" {
		g.state = g.prevState
	}
}

func (g *Game) updateShop() {
	mx, my, clicked := g.getClick()
	if !clicked {
		return
	}
	if pinDlg.IsOpen() {
		pinDlg.HandleClick(mx, my)
		return
	}
	if ShopHandleClick(mx, my) == "back" {
		g.state = g.prevState
	}
}

func (g *Game) updateLeaderboard() {
	_, _, clicked := g.getClick()
	if clicked {
		g.state = g.prevState
	}
}

// ── Goal check ────────────────────────────────────────────────────

func (g *Game) goalMet() bool {
	l := g.level
	b := g.board
	switch l.GoalType {
	case GoalClearColor:
		return g.cleared[l.GoalKind] >= l.GoalCount
	case GoalScore:
		return g.score >= l.ScoreGoal
	case GoalClearAny:
		return g.clearedAny >= l.GoalCount
	case GoalClearTwo:
		return g.cleared[l.GoalKind] >= l.GoalCount &&
			g.cleared[l.GoalKind2] >= l.GoalCount2
	case GoalBreakIce:
		return b.ClearedIce >= l.GoalCount
	case GoalBreakStone:
		return b.ClearedStone >= l.GoalCount
	case GoalBreakChain:
		return b.ClearedChain >= l.GoalCount
	case GoalBreakObstacle:
		return b.ClearedIce+b.ClearedStone+b.ClearedChain >= l.GoalCount
	}
	return false
}

// ── Draw ──────────────────────────────────────────────────────────

func (g *Game) Draw(screen *ebiten.Image) {
	// Render to offscreen first, then blit with shake offset.
	if g.offscreen == nil {
		g.offscreen = ebiten.NewImage(ScreenW, ScreenH)
	}
	canvas := g.offscreen
	canvas.Fill(color.RGBA{8, 4, 22, 255})
	g.drawStars(canvas)

	switch g.state {
	case StateAgeGate:
		DrawAgeGate(canvas)
	case StateMainMenu:
		DrawMainMenu(canvas, g.tick)
	case StateChapterSelect:
		DrawChapterSelect(canvas, g.tick)
	case StateLevelSelect:
		DrawLevelSelect(canvas, g.selectedBiome, g.hoverLevel, g.tick)
	case StateDialogue:
		if g.board != nil {
			g.drawPlaying(canvas)
		} else {
			biome := g.biomeIntroN
			if biome < 1 {
				biome = 1
			}
			DrawBiomeScene(canvas, biome, g.tick)
		}
		if g.dlg != nil {
			g.dlg.Draw(canvas)
		}
	case StatePlaying:
		g.drawPlaying(canvas)
	case StateWin:
		g.drawPlaying(canvas)
		g.ps.Draw(canvas)
		DrawWin(canvas, g.score, g.levelNum, g.level.MaxMoves, g.movesLeft)
	case StateLose:
		g.drawPlaying(canvas)
		DrawLose(canvas, g.score)
	case StateSettings:
		DrawSettings(canvas)
	case StateShop:
		DrawShop(canvas, g.tick)
	case StateLeaderboard:
		DrawLeaderboard(canvas)
	}

	// Blit offscreen → screen with shake
	screen.Fill(color.RGBA{8, 4, 22, 255})
	ox, oy := g.shake.Offset()
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(ox, oy)
	screen.DrawImage(canvas, op)

	// Transition overlay — screen-stable (not shaken)
	sceneTransition.Draw(screen)

	// PIN dialog — always on top, screen-stable
	pinDlg.Draw(screen)
}

func (g *Game) drawPlaying(screen *ebiten.Image) {
	// Subtle vignette
	for i := 0; i < 3; i++ {
		a := uint8(20 - i*6)
		vector.DrawFilledRect(screen,
			float32(i), float32(HUDHeight+i),
			float32(ScreenW-i*2), float32(ScreenH-HUDHeight-i*2),
			color.RGBA{60, 30, 100, a}, false)
	}
	if g.board != nil {
		g.board.Draw(screen)
	}
	g.ps.Draw(screen)
	if g.char != nil {
		g.char.Draw(screen)
	}
	if g.board != nil {
		DrawHUD(screen, g)
	}
}

func (g *Game) drawStars(screen *ebiten.Image) {
	for _, s := range g.stars {
		vector.DrawFilledCircle(screen, s.x, s.y, s.r, color.RGBA{200, 210, 255, s.alpha}, false)
	}
}

func (g *Game) Layout(_, _ int) (int, int) { return ScreenW, ScreenH }

// ── Input helper ──────────────────────────────────────────────────

func (g *Game) getClick() (mx, my int, clicked bool) {
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		mx, my = ebiten.CursorPosition()
		return mx, my, true
	}
	ids := inpututil.AppendJustPressedTouchIDs(nil)
	if len(ids) > 0 {
		mx, my = ebiten.TouchPosition(ids[0])
		return mx, my, true
	}
	return 0, 0, false
}
