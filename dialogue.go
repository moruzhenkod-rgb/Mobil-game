package main

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// ── Data ──────────────────────────────────────────────────────────

type DialogueLine struct {
	Speaker string
	Text    string
}

// dialogueTable maps level number to lines shown before that level.
var dialogueTable = map[int][]DialogueLine{
	// Biome 1 — Forest Ruins
	1: {
		{"???", "Who dares disturb the ruins of the Runic Forest?"},
		{"Aiden", "I'm Aiden, apprentice wizard. I seek the lost runes!"},
		{"???", "Then prove your worth. Match the gems to unlock the path."},
	},
	5: {
		{"Aiden", "These gems pulse with strange energy..."},
		{"Elder Mira", "Match 4 in a line — a power gem will form!"},
		{"Elder Mira", "Match 5 and you shall summon a rainbow gem!"},
	},
	10: {
		{"Forest Guardian", "Apprentice! You have proven worthy of the first seal."},
		{"Aiden", "What lies beyond the ruins?"},
		{"Forest Guardian", "Crystal Caves... where gems sleep frozen in time."},
	},
	// Biome 2 — Crystal Caves
	11: {
		{"Aiden", "The air is freezing! The gems are... encased in ice!"},
		{"Elder Mira", "Match adjacent gems to crack the ice. Two hits to shatter!"},
		{"Aiden", "Got it. Let's warm things up!"},
	},
	15: {
		{"Aiden", "Some ice is double-layered. It's much harder to crack!"},
		{"Elder Mira", "Use bonus gems — a bomb will shatter it instantly."},
	},
	20: {
		{"Ice Witch", "You've broken my eternal frost! The caves weep..."},
		{"Aiden", "I'm sorry, but I must find the runes."},
		{"Ice Witch", "The Volcanic Forge awaits. May you survive its heat."},
	},
	// Biome 3 — Volcanic Forge
	21: {
		{"Aiden", "Stone barriers everywhere! My magic barely scratches them!"},
		{"Elder Mira", "Only a bomb gem can smash stone. Create power tiles!"},
		{"Aiden", "Right — match 4 or more for bombs!"},
	},
	30: {
		{"Forge Master", "The flames respect your power, young wizard."},
		{"Aiden", "What is this place guarding?"},
		{"Forge Master", "Ancient tomes... chained by forgotten spells."},
	},
	// Biome 4 — Arcane Library
	31: {
		{"Aiden", "These gems are wrapped in chains!"},
		{"Elder Mira", "An adjacent match will break one chain."},
		{"Elder Mira", "The gem beneath is then free to match normally."},
	},
	40: {
		{"Librarian", "The chains are broken. The knowledge is yours."},
		{"Aiden", "I can feel the runes... they are above. In the sky!"},
		{"Librarian", "The Sky Citadel. Your final test awaits."},
	},
	// Biome 5 — Sky Citadel
	41: {
		{"Aiden", "Ice, stone, chains — all at once! This is impossible!"},
		{"Elder Mira", "Nothing is impossible for a true Runic Master."},
		{"Elder Mira", "Plan your moves. Save bombs for stone. Patience!"},
	},
	50: {
		{"Shadow Lord", "You dare challenge my citadel?"},
		{"Aiden", "I seek the lost runes. Stand aside!"},
		{"Shadow Lord", "Then face my ultimate puzzle... and despair."},
	},
	55: {
		{"Aiden", "I can see the final rune! Just a few more steps!"},
		{"Elder Mira", "You've come so far, Aiden. I believe in you."},
	},
	100: {
		{"Aiden", "The rune is mine! The curse is broken!"},
		{"Elder Mira", "Well done, Runic Master. The forest is saved."},
		{"Aiden", "This is only the beginning of my journey..."},
		{"Shadow Lord", "...We shall meet again."},
	},
}

// biomeIntroTable shows when entering a new biome for the first time.
var biomeIntroTable = [BiomeCount + 1][]DialogueLine{
	0: {},
	1: {{"Elder Mira", "Welcome to the Forest Ruins. Your adventure begins!"}},
	2: {
		{"Elder Mira", "The Crystal Caves! Watch out for frozen gems."},
		{"Elder Mira", "Match adjacent tiles to crack the ice — twice to clear!"},
	},
	3: {
		{"Elder Mira", "The Volcanic Forge. Stone blocks can only be destroyed"},
		{"Elder Mira", "by bomb gems. Create them by matching 4 in a line!"},
	},
	4: {
		{"Elder Mira", "The Arcane Library. Chained gems are everywhere."},
		{"Elder Mira", "An adjacent match breaks the chain. Then match normally."},
	},
	5: {
		{"Elder Mira", "The Sky Citadel — all obstacles combined!"},
		{"Elder Mira", "This is the ultimate test. Use every skill you've learned."},
	},
}

// ── Dialogue player ───────────────────────────────────────────────

type DialoguePlayer struct {
	lines    []DialogueLine
	idx      int
	tick     int
	charIdx  int   // how many chars to reveal (typewriter)
	full     bool  // line fully revealed
	done     bool
}

func NewDialogue(lines []DialogueLine) *DialoguePlayer {
	return &DialoguePlayer{lines: lines}
}

func (d *DialoguePlayer) Update() {
	if d.done {
		return
	}
	d.tick++
	if !d.full {
		d.charIdx += 2
		if d.charIdx >= len(d.lines[d.idx].Text) {
			d.charIdx = len(d.lines[d.idx].Text)
			d.full = true
		}
	}
}

// Tap advances the dialogue; returns true when all lines are exhausted.
func (d *DialoguePlayer) Tap() bool {
	if d.done {
		return true
	}
	if !d.full {
		// Instantly reveal current line
		d.charIdx = len(d.lines[d.idx].Text)
		d.full = true
		return false
	}
	d.idx++
	if d.idx >= len(d.lines) {
		d.done = true
		return true
	}
	d.charIdx = 0
	d.full = false
	return false
}

func (d *DialoguePlayer) IsDone() bool { return d.done }

func (d *DialoguePlayer) Draw(screen *ebiten.Image) {
	if d.done || len(d.lines) == 0 {
		return
	}
	cur := d.lines[d.idx]

	// Panel
	py := float32(ScreenH - 220)
	vector.DrawFilledRect(screen, 20, py, ScreenW-40, 200, color.RGBA{12, 6, 30, 240}, false)
	vector.DrawFilledRect(screen, 20, py, ScreenW-40, 3, color.RGBA{160, 100, 255, 200}, false)
	vector.DrawFilledRect(screen, 20, py+197, ScreenW-40, 3, color.RGBA{160, 100, 255, 200}, false)

	// Speaker name
	ebitenutil.DebugPrintAt(screen, cur.Speaker, 36, int(py)+14)

	// Text (typewriter)
	txt := cur.Text
	if d.charIdx < len(txt) {
		txt = txt[:d.charIdx]
	}
	// Word-wrap at ~60 chars
	wrapped := wordWrap(txt, 62)
	for i, line := range wrapped {
		ebitenutil.DebugPrintAt(screen, line, 36, int(py)+38+i*16)
	}

	// Prompt
	if d.full {
		prompt := "[ Tap to continue ]"
		if d.idx >= len(d.lines)-1 {
			prompt = "[ Tap to start ]"
		}
		blinkAlpha := uint8(128 + 127*(d.tick%40)/40)
		_ = blinkAlpha
		ebitenutil.DebugPrintAt(screen, prompt, ScreenW/2-len(prompt)*3, int(py)+170)
	}
}

func wordWrap(s string, width int) []string {
	var lines []string
	for len(s) > width {
		cut := width
		for cut > 0 && s[cut] != ' ' {
			cut--
		}
		if cut == 0 {
			cut = width
		}
		lines = append(lines, s[:cut])
		s = s[cut+1:]
	}
	if len(s) > 0 {
		lines = append(lines, s)
	}
	return lines
}
