package main

import (
	"bytes"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	textv2 "github.com/hajimehoshi/ebiten/v2/text/v2"
	"golang.org/x/image/font/gofont/gobold"
	"golang.org/x/image/font/gofont/goregular"
)

// ── Font faces ────────────────────────────────────────────────────

var (
	fontSrcR *textv2.GoTextFaceSource // regular
	fontSrcB *textv2.GoTextFaceSource // bold

	FontSm  *textv2.GoTextFace // 13 px — captions, labels
	FontMd  *textv2.GoTextFace // 20 px — body, buttons
	FontLg  *textv2.GoTextFace // 30 px — section titles
	FontXL  *textv2.GoTextFace // 52 px — main title (bold)
	FontBtn *textv2.GoTextFace // 17 px bold — button labels
)

func InitFonts() {
	var err error
	fontSrcR, err = textv2.NewGoTextFaceSource(bytes.NewReader(goregular.TTF))
	if err != nil {
		panic("font init: " + err.Error())
	}
	fontSrcB, err = textv2.NewGoTextFaceSource(bytes.NewReader(gobold.TTF))
	if err != nil {
		panic("font init bold: " + err.Error())
	}
	FontSm  = &textv2.GoTextFace{Source: fontSrcR, Size: 13}
	FontMd  = &textv2.GoTextFace{Source: fontSrcR, Size: 20}
	FontLg  = &textv2.GoTextFace{Source: fontSrcR, Size: 30}
	FontXL  = &textv2.GoTextFace{Source: fontSrcB, Size: 52}
	FontBtn = &textv2.GoTextFace{Source: fontSrcB, Size: 17}
}

// ── Draw helpers ──────────────────────────────────────────────────

// FDraw draws text at (x, y) — top-left origin.
func FDraw(dst *ebiten.Image, s string, x, y float64, face *textv2.GoTextFace, clr color.Color) {
	op := &textv2.DrawOptions{}
	op.GeoM.Translate(x, y)
	op.ColorScale.ScaleWithColor(clr)
	textv2.Draw(dst, s, face, op)
}

// FCenter draws text horizontally centered around cx.
func FCenter(dst *ebiten.Image, s string, cx, y float64, face *textv2.GoTextFace, clr color.Color) {
	w, _ := textv2.Measure(s, face, 0)
	FDraw(dst, s, cx-w/2, y, face, clr)
}

// FMeasureW returns the pixel width of string s in face.
func FMeasureW(s string, face *textv2.GoTextFace) float64 {
	w, _ := textv2.Measure(s, face, 0)
	return w
}

// Palette — frequently used colours
var (
	ColWhite  = color.RGBA{255, 255, 255, 255}
	ColGold   = color.RGBA{255, 215, 60, 255}
	ColPurple = color.RGBA{180, 120, 255, 255}
	ColGray   = color.RGBA{160, 150, 180, 255}
	ColGreen  = color.RGBA{80, 220, 100, 255}
	ColRed    = color.RGBA{240, 80, 80, 255}
	ColDim    = color.RGBA{120, 110, 140, 255}
)
