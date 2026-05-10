package main

import (
	"crypto/sha256"
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// ── PIN dialog ────────────────────────────────────────────────────
// Provides a numpad overlay for parental-controls PIN entry.
// Supports: verify existing PIN, set new PIN, confirm new PIN.

type PINMode int

const (
	PINNone    PINMode = iota
	PINVerify          // validate against stored PIN → call onSuccess
	PINSetNew          // first entry for a new PIN
	PINConfirm         // re-entry confirmation for new PIN
)

type PINDialog struct {
	mode      PINMode
	digits    string // currently entered digits (max 4)
	firstPass string // stored first entry during PINSetNew
	onSuccess func()
	onCancel  func()
	errMsg    string
	errTick   int
}

var pinDlg = &PINDialog{}

// OpenPINVerify opens the dialog to verify the existing PIN.
// onSuccess is called if the PIN matches (or no PIN is set).
func OpenPINVerify(onSuccess func()) {
	if progress.ParentalPIN == "" || !progress.RequirePIN {
		onSuccess()
		return
	}
	pinDlg.open(PINVerify, onSuccess, nil)
}

// OpenPINSet opens the dialog to set a new (or change) PIN.
func OpenPINSet(onDone func()) {
	pinDlg.open(PINSetNew, onDone, onDone)
}

func (p *PINDialog) open(mode PINMode, onSuccess, onCancel func()) {
	p.mode = mode
	p.digits = ""
	p.firstPass = ""
	p.onSuccess = onSuccess
	p.onCancel = onCancel
	p.errMsg = ""
	p.errTick = 0
}

func (p *PINDialog) IsOpen() bool { return p.mode != PINNone }

func (p *PINDialog) close() {
	p.mode = PINNone
	p.digits = ""
	p.firstPass = ""
	p.errMsg = ""
}

func hashPIN(pin string) string {
	h := sha256.Sum256([]byte(pin))
	return fmt.Sprintf("%x", h)
}

// ── Update / click handler ────────────────────────────────────────

// HandleClick returns true if the dialog consumed the click.
func (p *PINDialog) HandleClick(mx, my int) bool {
	if !p.IsOpen() {
		return false
	}
	if p.errTick > 0 {
		p.errTick--
	}

	// Cancel button
	if mx >= 60 && mx <= 220 && my >= 750 && my <= 800 {
		p.close()
		if p.onCancel != nil {
			p.onCancel()
		}
		return true
	}

	// Numpad: 3 columns × 4 rows starting at (60, 440)
	numX0, numY0 := 60, 440
	btnW, btnH := 120, 68
	gap := 8

	keys := []string{
		"1", "2", "3",
		"4", "5", "6",
		"7", "8", "9",
		"←", "0", "OK",
	}
	for i, k := range keys {
		row := i / 3
		col := i % 3
		bx := numX0 + col*(btnW+gap)
		by := numY0 + row*(btnH+gap)
		if mx >= bx && mx <= bx+btnW && my >= by && my <= by+btnH {
			p.pressKey(k)
			PlaySound(SndButton)
			return true
		}
	}
	return true // swallow all clicks while dialog is open
}

func (p *PINDialog) pressKey(k string) {
	switch k {
	case "←":
		if len(p.digits) > 0 {
			p.digits = p.digits[:len(p.digits)-1]
		}
	case "OK":
		p.confirm()
	default:
		if len(p.digits) < 4 {
			p.digits += k
		}
		if len(p.digits) == 4 {
			p.confirm()
		}
	}
}

func (p *PINDialog) confirm() {
	if len(p.digits) < 4 {
		p.errMsg = "Enter 4 digits"
		p.errTick = 90
		return
	}
	switch p.mode {
	case PINVerify:
		if hashPIN(p.digits) == progress.ParentalPIN {
			cb := p.onSuccess
			p.close()
			if cb != nil {
				cb()
			}
		} else {
			p.digits = ""
			p.errMsg = "Wrong PIN"
			p.errTick = 90
		}
	case PINSetNew:
		p.firstPass = p.digits
		p.digits = ""
		p.mode = PINConfirm
	case PINConfirm:
		if p.digits == p.firstPass {
			progress.ParentalPIN = hashPIN(p.digits)
			progress.RequirePIN = true
			saveProgress()
			cb := p.onSuccess
			p.close()
			if cb != nil {
				cb()
			}
		} else {
			p.digits = ""
			p.firstPass = ""
			p.mode = PINSetNew
			p.errMsg = "PINs don't match — try again"
			p.errTick = 90
		}
	}
}

// ── Draw ──────────────────────────────────────────────────────────

func (p *PINDialog) Draw(screen *ebiten.Image) {
	if !p.IsOpen() {
		return
	}

	// Dim overlay
	ov := ebiten.NewImage(ScreenW, ScreenH)
	ov.Fill(color.RGBA{0, 0, 0, 190})
	screen.DrawImage(ov, nil)

	// Panel
	vector.DrawFilledRect(screen, 40, 200, ScreenW-80, 620, color.RGBA{14, 8, 38, 252}, false)
	vector.DrawFilledRect(screen, 40, 200, ScreenW-80, 3, color.RGBA{140, 80, 220, 255}, false)
	vector.DrawFilledRect(screen, 40, 817, ScreenW-80, 3, color.RGBA{140, 80, 220, 255}, false)

	// Title
	title := "Parental PIN"
	switch p.mode {
	case PINVerify:
		title = "Enter PIN"
	case PINSetNew:
		title = "Set New PIN"
	case PINConfirm:
		title = "Confirm PIN"
	}
	FCenter(screen, title, ScreenW/2, 220, FontLg, ColGold)

	// Dot display
	for i := 0; i < 4; i++ {
		dx := float32(ScreenW/2 - 60 + i*40)
		filled := i < len(p.digits)
		c := color.RGBA{120, 80, 200, 255}
		if filled {
			c = color.RGBA{220, 180, 255, 255}
		}
		vector.DrawFilledCircle(screen, dx, 310, 14, c, false)
	}

	// Error message
	if p.errTick > 0 {
		FCenter(screen, p.errMsg, ScreenW/2, 350, FontSm, ColRed)
	}

	// Numpad
	numX0, numY0 := 60, 440
	btnW, btnH := 120, 68
	gap := 8
	labels := []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "←", "0", "OK"}

	for i, lbl := range labels {
		row := i / 3
		col := i % 3
		bx := float32(numX0 + col*(btnW+gap))
		by := float32(numY0 + row*(btnH+gap))

		bg := color.RGBA{30, 18, 65, 255}
		tc := ColWhite
		if lbl == "OK" {
			bg = color.RGBA{40, 140, 60, 255}
			tc = ColGreen
		} else if lbl == "←" {
			bg = color.RGBA{80, 35, 35, 255}
			tc = ColRed
		}
		vector.DrawFilledRect(screen, bx, by, float32(btnW), float32(btnH), bg, false)
		vector.DrawFilledRect(screen, bx, by, float32(btnW), 2, color.RGBA{80, 60, 120, 200}, false)
		FCenter(screen, lbl, float64(bx)+float64(btnW)/2, float64(by)+float64(btnH)/2-12, FontBtn, tc)
	}

	// Cancel
	vector.DrawFilledRect(screen, 60, 750, 380, 50, color.RGBA{60, 35, 35, 240}, false)
	vector.DrawFilledRect(screen, 60, 750, 380, 2, color.RGBA{180, 60, 60, 200}, false)
	FCenter(screen, "Cancel", ScreenW/2, 761, FontBtn, ColRed)
}
