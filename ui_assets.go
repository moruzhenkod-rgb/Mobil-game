package main

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

var (
	// Button images — preprocessed to display size (cover crop, no distortion)
	btnPlay      *ebiten.Image // 280 × 70
	btnAdventure *ebiten.Image // 264 × 93
	btnLevels    *ebiten.Image // 264 × 93
	btnShop      *ebiten.Image // 264 × 91
	btnQuests    *ebiten.Image // 176 × 88
	btnSettings  *ebiten.Image // 176 × 88
	btnMessages  *ebiten.Image // 176 × 88

	// HUD icons — preprocessed square
	icoCoin        *ebiten.Image // 24 × 24
	icoGem         *ebiten.Image // 24 × 24
	imgAvatarRound *ebiten.Image // 58 × 58 circular

	uiAssetsLoaded bool
)

func rawImg(path string) *ebiten.Image {
	img, _, err := ebitenutil.NewImageFromFile(path)
	if err != nil {
		return nil
	}
	return img
}

func ensureUIAssets() {
	if uiAssetsLoaded {
		return
	}
	uiAssetsLoaded = true

	btnPlay      = imgCoverCrop(rawImg("playbtn.jpg"),     280, 70)
	btnAdventure = imgCoverCrop(rawImg("ADVENTURE.jpg"),   264, 93)
	btnLevels    = imgCoverCrop(rawImg("LEVELS.jpg"),      264, 93)
	btnShop      = imgCoverCrop(rawImg("SHOP.jpg"),        264, 91)
	btnQuests    = imgCoverCrop(rawImg("QUESTS.jpg"),      176, 88)
	btnSettings  = imgCoverCrop(rawImg("SETTINGS.jpg"),    176, 88)
	btnMessages  = imgCoverCrop(rawImg("MESSAGES.jpg"),    176, 88)

	icoCoin = imgCoverCrop(rawImg("coin.jpg"), 24, 24)
	icoGem  = imgCoverCrop(rawImg("gem.jpg"),  24, 24)

	imgAvatarRound = makeCircularImage(rawImg("avatar.jpg"), 58)
}

// imgCoverCrop scales src to fill (w×h) from center, clipping excess — no distortion.
func imgCoverCrop(src *ebiten.Image, w, h int) *ebiten.Image {
	dst := ebiten.NewImage(w, h)
	if src == nil {
		return dst
	}
	iw, ih := src.Bounds().Dx(), src.Bounds().Dy()
	scaleX := float64(w) / float64(iw)
	scaleY := float64(h) / float64(ih)
	scale := scaleX
	if scaleY > scale {
		scale = scaleY
	}
	offX := (float64(w) - float64(iw)*scale) / 2
	offY := (float64(h) - float64(ih)*scale) / 2
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(scale, scale)
	op.GeoM.Translate(offX, offY)
	dst.DrawImage(src, op)
	return dst
}

// makeCircularImage clips src into a circle of the given diameter.
func makeCircularImage(src *ebiten.Image, size int) *ebiten.Image {
	dst := ebiten.NewImage(size, size)
	if src == nil {
		return dst
	}
	vector.DrawFilledCircle(dst,
		float32(size)/2, float32(size)/2, float32(size)/2,
		color.White, true)
	iw, ih := src.Bounds().Dx(), src.Bounds().Dy()
	scaleX := float64(size) / float64(iw)
	scaleY := float64(size) / float64(ih)
	scale := scaleX
	if scaleY > scale {
		scale = scaleY
	}
	offX := (float64(size) - float64(iw)*scale) / 2
	offY := (float64(size) - float64(ih)*scale) / 2
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(scale, scale)
	op.GeoM.Translate(offX, offY)
	op.Blend = ebiten.BlendSourceIn
	dst.DrawImage(src, op)
	return dst
}

// drawImgAt draws a precomputed image at exact pixel position.
func drawImgAt(screen, img *ebiten.Image, x, y int) {
	if img == nil {
		return
	}
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(x), float64(y))
	screen.DrawImage(img, op)
}

// drawImgAtAlpha draws a precomputed image at position with given alpha (0–1).
func drawImgAtAlpha(screen, img *ebiten.Image, x, y int, alpha float32) {
	if img == nil {
		return
	}
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(x), float64(y))
	op.ColorScale.ScaleAlpha(alpha)
	screen.DrawImage(img, op)
}
