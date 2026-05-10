package main

import (
	"bytes"
	"math"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2/audio"
)

const sampleRate = 44100

// ── Sound IDs ─────────────────────────────────────────────────────

type SoundID int

const (
	SndSwap    SoundID = iota // gem swap attempt
	SndMatch                  // successful match
	SndBonus                  // bonus tile created
	SndBomb                   // bomb/row/col explodes
	SndRainbow                // rainbow tile activates
	SndIce                    // ice cracked
	SndWin                    // level complete fanfare
	SndLose                   // out of moves
	SndCoin                   // coin collected / purchase
	SndButton                 // UI tap
)

// ── Manager ───────────────────────────────────────────────────────

type AudioManager struct {
	ctx    *audio.Context
	sfx    map[SoundID][]byte
	music  *audio.Player
}

var audioMgr *AudioManager

func InitAudio() {
	ctx := audio.NewContext(sampleRate)
	audioMgr = &AudioManager{
		ctx: ctx,
		sfx: make(map[SoundID][]byte),
	}
	audioMgr.pregenEffects()
	if progress.MusicOn {
		audioMgr.startMusic()
	}
}

func (a *AudioManager) pregenEffects() {
	a.sfx[SndSwap]    = genTone(440, 0.06, 0.28)
	a.sfx[SndMatch]   = genArpeggio([]float64{523, 659, 784}, 0.07, 0.30)
	a.sfx[SndBonus]   = genArpeggio([]float64{784, 1047, 1319}, 0.06, 0.28)
	a.sfx[SndBomb]    = genNoise(0.18, 0.50)
	a.sfx[SndRainbow] = genArpeggio([]float64{392, 494, 587, 740, 880}, 0.06, 0.25)
	a.sfx[SndIce]     = genTone(1400, 0.04, 0.22)
	a.sfx[SndWin]     = genArpeggio([]float64{523, 659, 784, 1047}, 0.13, 0.45)
	a.sfx[SndLose]    = genArpeggio([]float64{440, 370, 330, 262}, 0.15, 0.38)
	a.sfx[SndCoin]    = genArpeggio([]float64{880, 1109, 1319}, 0.05, 0.20)
	a.sfx[SndButton]  = genTone(660, 0.04, 0.18)
}

// PlaySound plays a sound effect (fire-and-forget).
func PlaySound(id SoundID) {
	if audioMgr == nil || !progress.SoundOn {
		return
	}
	data, ok := audioMgr.sfx[id]
	if !ok {
		return
	}
	p, err := audioMgr.ctx.NewPlayer(bytes.NewReader(data))
	if err != nil {
		return
	}
	p.Play()
}

// SetMusicEnabled starts or pauses the background music track.
func SetMusicEnabled(on bool) {
	if audioMgr == nil {
		return
	}
	if on {
		if audioMgr.music == nil || !audioMgr.music.IsPlaying() {
			audioMgr.startMusic()
		}
	} else if audioMgr.music != nil {
		audioMgr.music.Pause()
	}
}

func (a *AudioManager) startMusic() {
	if a.music != nil {
		a.music.Close()
		a.music = nil
	}
	data := genMusicLoop()
	loop := audio.NewInfiniteLoop(bytes.NewReader(data), int64(len(data)))
	p, err := a.ctx.NewPlayer(loop)
	if err != nil {
		return
	}
	p.SetVolume(0.32)
	p.Play()
	a.music = p
}

// ── PCM helpers ───────────────────────────────────────────────────

func writeStereo(buf []byte, pos int, v float64) {
	if v > 1 {
		v = 1
	} else if v < -1 {
		v = -1
	}
	s := int16(v * 32767)
	buf[pos*4+0] = byte(s)
	buf[pos*4+1] = byte(s >> 8)
	buf[pos*4+2] = byte(s)
	buf[pos*4+3] = byte(s >> 8)
}

func env(t, dur float64) float64 {
	atk := 0.008
	rel := dur * 0.25
	switch {
	case t < atk:
		return t / atk
	case t > dur-rel:
		return (dur - t) / rel
	default:
		return 1.0
	}
}

// genTone creates a mono sine tone converted to stereo PCM.
func genTone(freq, dur, amp float64) []byte {
	n := int(sampleRate * dur)
	buf := make([]byte, n*4)
	for i := 0; i < n; i++ {
		t := float64(i) / sampleRate
		writeStereo(buf, i, math.Sin(2*math.Pi*freq*t)*amp*env(t, dur))
	}
	return buf
}

// genArpeggio concatenates a sequence of tones.
func genArpeggio(freqs []float64, noteDur, amp float64) []byte {
	n := int(sampleRate * noteDur)
	buf := make([]byte, len(freqs)*n*4)
	for ni, f := range freqs {
		off := ni * n
		for i := 0; i < n; i++ {
			t := float64(i) / sampleRate
			v := (math.Sin(2*math.Pi*f*t)*0.75 +
				math.Sin(4*math.Pi*f*t)*0.25) * amp * env(t, noteDur)
			writeStereo(buf, off+i, v)
		}
	}
	return buf
}

// genNoise creates a lowpass-filtered noise burst (explosion sound).
func genNoise(dur, amp float64) []byte {
	n := int(sampleRate * dur)
	buf := make([]byte, n*4)
	prev := 0.0
	for i := 0; i < n; i++ {
		t := float64(i) / sampleRate
		raw := rand.Float64()*2 - 1
		prev = prev*0.80 + raw*0.20
		writeStereo(buf, i, prev*amp*env(t, dur))
	}
	return buf
}

// genMusicLoop generates a ~7-second mystical arpeggio loop in A-minor.
func genMusicLoop() []byte {
	type note struct{ f, d float64 }
	beat := 60.0 / 108.0
	h := beat / 2

	seq := []note{
		{220, beat}, {261.63, h}, {329.63, h},
		{392, beat}, {329.63, h}, {261.63, h},
		{220, beat}, {174.61, h}, {220, h},
		{261.63, beat}, {329.63, beat},
		{392, beat}, {440, h}, {392, h},
		{349.23, beat}, {329.63, beat},
	}

	total := 0
	for _, p := range seq {
		total += int(sampleRate * p.d)
	}
	buf := make([]byte, total*4)
	pos := 0
	for _, p := range seq {
		n := int(sampleRate * p.d)
		for i := 0; i < n; i++ {
			t := float64(i) / sampleRate
			v := (math.Sin(2*math.Pi*p.f*t)*0.65 +
				math.Sin(4*math.Pi*p.f*t)*0.20 +
				math.Sin(6*math.Pi*p.f*t)*0.10) * 0.20 * env(t, p.d)
			writeStereo(buf, pos+i, v)
		}
		pos += n
	}
	return buf
}
