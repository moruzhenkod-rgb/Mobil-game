package main

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// ── Settings screen ───────────────────────────────────────────────

type SettingsSection int

const (
	SecGeneral  SettingsSection = iota
	SecAccount
	SecParental
	SecGDPR
	SecAbout
)

type SettingsState struct {
	section SettingsSection
}

var settingsState = &SettingsState{}

func DrawSettings(screen *ebiten.Image) {
	// Header
	vector.DrawFilledRect(screen, 0, 0, ScreenW, 56, color.RGBA{10, 5, 28, 245}, false)
	ebitenutil.DebugPrintAt(screen, T("settings"), ScreenW/2-21, 10)
	drawSettingsBtn(screen, T("back"), 12, 10, 70, 28, color.RGBA{40, 25, 70, 220})
	vector.DrawFilledRect(screen, 0, 52, ScreenW, 3, color.RGBA{120, 60, 200, 160}, false)

	// Tab buttons
	tabs := []string{"General", "Account", "Parental", "GDPR", "About"}
	for i, tab := range tabs {
		bx := 10 + i*104
		bg := color.RGBA{25, 14, 55, 220}
		if SettingsSection(i) == settingsState.section {
			bg = color.RGBA{60, 35, 100, 255}
		}
		vector.DrawFilledRect(screen, float32(bx), 60, 100, 32, bg, false)
		ebitenutil.DebugPrintAt(screen, tab, bx+50-len(tab)*3, 72)
	}
	vector.DrawFilledRect(screen, 0, 90, ScreenW, 2, color.RGBA{80, 50, 140, 160}, false)

	switch settingsState.section {
	case SecGeneral:
		drawGeneralSettings(screen)
	case SecAccount:
		drawAccountSettings(screen)
	case SecParental:
		drawParentalSettings(screen)
	case SecGDPR:
		drawGDPRSettings(screen)
	case SecAbout:
		drawAboutSettings(screen)
	}
}

func drawGeneralSettings(screen *ebiten.Image) {
	y := 108
	// Language
	drawRow(screen, T("lang_label"), "Language", y)
	// Display current language name
	langName := LangNames[currentLang]
	drawToggleRow(screen, "< "+langName+" >", y, true)
	y += 60

	// Sound
	drawRow(screen, T("sound"), "", y)
	soundLabel := T("on")
	if !progress.SoundOn {
		soundLabel = T("off")
	}
	drawToggleRow(screen, soundLabel, y, progress.SoundOn)
	y += 60

	// Music
	drawRow(screen, T("music"), "", y)
	musicLabel := T("on")
	if !progress.MusicOn {
		musicLabel = T("off")
	}
	drawToggleRow(screen, musicLabel, y, progress.MusicOn)
	y += 60

	// Notifications
	drawRow(screen, T("notifs"), "", y)
	notifLabel := T("on")
	if !progress.NotifsOn {
		notifLabel = T("off")
	}
	drawToggleRow(screen, notifLabel, y, progress.NotifsOn)
	y += 60

	// Firebase status
	status := "Cloud: " + T("sync")
	if fb.IsOffline() {
		status = "Cloud: " + T("offline")
	} else if fb.IsReady() {
		status = "Cloud: Connected"
	}
	ebitenutil.DebugPrintAt(screen, status, 24, y+16)
}

func drawAccountSettings(screen *ebiten.Image) {
	y := 108
	if progress.UserEmail != "" {
		ebitenutil.DebugPrintAt(screen, "Signed in as:", 24, y)
		ebitenutil.DebugPrintAt(screen, progress.UserEmail, 24, y+18)
		y += 60
		drawSettingsBtn(screen, T("logout"), 24, y, 180, 44, color.RGBA{80, 40, 100, 255})
		y += 60
	} else {
		ebitenutil.DebugPrintAt(screen, "Not signed in.", 24, y)
		y += 30
		drawSettingsBtn(screen, T("login_google"), 24, y, 240, 44, color.RGBA{50, 80, 200, 255})
		y += 56
		drawSettingsBtn(screen, T("login_apple"), 24, y, 240, 44, color.RGBA{30, 30, 50, 255})
		y += 56
		ebitenutil.DebugPrintAt(screen, "Sign in to sync progress across devices.", 24, y)
	}
	y += 60
	// Leaderboard
	drawSettingsBtn(screen, T("leaderboard"), 24, y, 180, 44, color.RGBA{60, 100, 160, 255})
	y += 60
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Best Score: %d", progress.BestScore), 24, y)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Total Stars: %d/%d", totalStars(), MaxLevels*3), 24, y+18)
}

func drawParentalSettings(screen *ebiten.Image) {
	y := 108
	ebitenutil.DebugPrintAt(screen, "Parental Controls", 24, y)
	y += 30

	pinStatus := "PIN: Not set"
	if progress.ParentalPIN != "" {
		pinStatus = "PIN: Set (****)"
	}
	ebitenutil.DebugPrintAt(screen, pinStatus, 24, y)
	drawSettingsBtn(screen, "Change PIN", ScreenW-180, y-6, 156, 36, color.RGBA{60, 40, 100, 255})
	y += 52

	limitStr := "No limit"
	if progress.MonthlyLimit > 0 {
		limitStr = fmt.Sprintf("$%.2f / month", float64(progress.MonthlyLimit)/100)
	}
	ebitenutil.DebugPrintAt(screen, T("spend_limit")+": "+limitStr, 24, y)
	drawSettingsBtn(screen, "Set Limit", ScreenW-160, y-6, 136, 36, color.RGBA{60, 40, 100, 255})
	y += 52

	spentStr := fmt.Sprintf("This month: $%.2f", float64(progress.MonthlySpent)/100)
	ebitenutil.DebugPrintAt(screen, spentStr, 24, y)
	y += 40

	ebitenutil.DebugPrintAt(screen, "When PIN is set, all purchases", 24, y)
	ebitenutil.DebugPrintAt(screen, "require PIN confirmation.", 24, y+18)
	y += 52

	ebitenutil.DebugPrintAt(screen, "Require PIN for in-app purchases:", 24, y)
	reqLabel := T("on")
	if !progress.RequirePIN {
		reqLabel = T("off")
	}
	drawToggleRow(screen, reqLabel, y, progress.RequirePIN)
}

func drawGDPRSettings(screen *ebiten.Image) {
	y := 108
	ebitenutil.DebugPrintAt(screen, "Privacy & Data (GDPR)", 24, y)
	y += 30

	ebitenutil.DebugPrintAt(screen, "We collect gameplay data to improve", 24, y)
	ebitenutil.DebugPrintAt(screen, "your experience. No personal data", 24, y+18)
	ebitenutil.DebugPrintAt(screen, "is sold to third parties.", 24, y+36)
	y += 72

	drawSettingsBtn(screen, T("privacy"), 24, y, 200, 44, color.RGBA{50, 35, 90, 255})
	y += 56
	drawSettingsBtn(screen, T("tos"), 24, y, 200, 44, color.RGBA{50, 35, 90, 255})
	y += 56
	drawSettingsBtn(screen, T("gdpr_export"), 24, y, 200, 44, color.RGBA{40, 80, 120, 255})
	y += 56
	drawSettingsBtn(screen, T("gdpr_delete"), 24, y, 200, 44, color.RGBA{140, 40, 40, 255})
	y += 56

	ebitenutil.DebugPrintAt(screen, "Analytics: ON (helps us improve the game)", 24, y)
	ebitenutil.DebugPrintAt(screen, "Data region: EU (GDPR compliant)", 24, y+18)
}

func drawAboutSettings(screen *ebiten.Image) {
	y := 108
	ebitenutil.DebugPrintAt(screen, "RUNIC CRUSH", ScreenW/2-33, y)
	y += 24
	ebitenutil.DebugPrintAt(screen, "Version 0.4.0 (Stage 4)", 24, y)
	y += 24
	ebitenutil.DebugPrintAt(screen, "Built with Ebitengine (Go)", 24, y)
	y += 40
	ebitenutil.DebugPrintAt(screen, "Licensed under MIT.", 24, y)
	y += 24
	ebitenutil.DebugPrintAt(screen, "Firebase (Google LLC) for cloud services.", 24, y)
	y += 24
	ebitenutil.DebugPrintAt(screen, "AdMob (Google LLC) for advertising.", 24, y)
	y += 40
	ebitenutil.DebugPrintAt(screen, "Support: support@runiccrush.game", 24, y)
	y += 24
	ebitenutil.DebugPrintAt(screen, "For GDPR requests: privacy@runiccrush.game", 24, y)
}

// ── Settings helpers ──────────────────────────────────────────────

func drawRow(screen *ebiten.Image, label, _ string, y int) {
	vector.DrawFilledRect(screen, 0, float32(y), ScreenW, 52, color.RGBA{18, 10, 45, 200}, false)
	ebitenutil.DebugPrintAt(screen, label, 24, y+18)
}

func drawToggleRow(screen *ebiten.Image, value string, y int, on bool) {
	c := color.RGBA{50, 160, 70, 255}
	if !on {
		c = color.RGBA{100, 50, 60, 255}
	}
	drawSettingsBtn(screen, value, ScreenW-130, y+10, 110, 32, c)
}

func drawSettingsBtn(screen *ebiten.Image, label string, x, y, w, h int, bg color.RGBA) {
	vector.DrawFilledRect(screen, float32(x), float32(y), float32(w), float32(h), bg, false)
	vector.DrawFilledRect(screen, float32(x), float32(y), float32(w), 2,
		color.RGBA{bg.R + 30, bg.G + 30, bg.B + 30, 200}, false)
	ebitenutil.DebugPrintAt(screen, label, x+w/2-len(label)*3, y+h/2-5)
}

// ── Settings input ────────────────────────────────────────────────

// SettingsHandleClick returns "back" or "".
func SettingsHandleClick(mx, my int) string {
	// Back
	if mx >= 12 && mx <= 82 && my >= 10 && my <= 38 {
		return "back"
	}

	// Tab selection
	if my >= 60 && my <= 92 {
		for i := 0; i < 5; i++ {
			bx := 10 + i*104
			if mx >= bx && mx <= bx+100 {
				settingsState.section = SettingsSection(i)
				return ""
			}
		}
	}

	switch settingsState.section {
	case SecGeneral:
		settingsGeneralClick(mx, my)
	case SecParental:
		settingsParentalClick(mx, my)
	case SecGDPR:
		settingsGDPRClick(mx, my)
	}
	return ""
}

func settingsGeneralClick(mx, my int) {
	// Language toggle row: y=108
	if my >= 108 && my <= 160 && mx >= ScreenW-140 {
		cycleLanguage()
		return
	}
	// Sound row: y=168
	if my >= 168 && my <= 220 && mx >= ScreenW-140 {
		progress.SoundOn = !progress.SoundOn
		saveProgress()
		return
	}
	// Music row: y=228
	if my >= 228 && my <= 280 && mx >= ScreenW-140 {
		progress.MusicOn = !progress.MusicOn
		saveProgress()
		SetMusicEnabled(progress.MusicOn)
		return
	}
	// Notifs row: y=288
	if my >= 288 && my <= 340 && mx >= ScreenW-140 {
		progress.NotifsOn = !progress.NotifsOn
		saveProgress()
		return
	}
}

func settingsParentalClick(mx, my int) {
	// "Change PIN" button: y=132..168, x=ScreenW-180..ScreenW-24
	if my >= 132 && my <= 168 && mx >= ScreenW-180 {
		OpenPINSet(nil)
		return
	}
	// "Require PIN" toggle: y=260..296
	if my >= 260 && my <= 296 && mx >= ScreenW-160 {
		progress.RequirePIN = !progress.RequirePIN
		saveProgress()
	}
}

func settingsGDPRClick(mx, my int) {
	// Export: y ~= 390
	// Delete: y ~= 446
	// (stubs — would open browser/email in real build)
}

func cycleLanguage() {
	langs := AllLangs
	for i, l := range langs {
		if l == currentLang {
			next := langs[(i+1)%len(langs)]
			SetLang(next)
			return
		}
	}
	SetLang(LangEN)
}
