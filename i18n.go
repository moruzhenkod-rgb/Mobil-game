package main

// Lang identifies a supported UI language.
type Lang string

const (
	LangEN Lang = "en"
	LangRU Lang = "ru"
	LangDE Lang = "de"
	LangFR Lang = "fr"
	LangES Lang = "es"
	LangZH Lang = "zh"
)

var AllLangs = []Lang{LangEN, LangRU, LangDE, LangFR, LangES, LangZH}

var LangNames = map[Lang]string{
	LangEN: "English",
	LangRU: "Русский",
	LangDE: "Deutsch",
	LangFR: "Français",
	LangES: "Español",
	LangZH: "中文",
}

// currentLang is set from SaveData.Language at startup.
var currentLang = LangEN

func SetLang(l Lang) {
	currentLang = l
	progress.Language = string(l)
	saveProgress()
}

// T returns a localised string for key.
// Falls back to English, then to the key itself.
func T(key string) string {
	if m, ok := translations[currentLang]; ok {
		if v, ok2 := m[key]; ok2 {
			return v
		}
	}
	if m, ok := translations[LangEN]; ok {
		if v, ok2 := m[key]; ok2 {
			return v
		}
	}
	return key
}

// ── Translation table ─────────────────────────────────────────────
// Note: ebitenutil.DebugPrint only renders ASCII; non-ASCII chars are
// shown as '?' until Stage 5 adds a proper TTF renderer.
// Keys stay English-named for code clarity.

var translations = map[Lang]map[string]string{
	LangEN: {
		"level":          "LEVEL",
		"score":          "SCORE",
		"moves":          "MOVES",
		"goal":           "GOAL",
		"combo":          "COMBO",
		"play":           "PLAY",
		"continue":       "CONTINUE",
		"settings":       "SETTINGS",
		"shop":           "SHOP",
		"back":           "< BACK",
		"retry":          "RETRY LEVEL",
		"next_level":     "NEXT LEVEL >>",
		"level_select":   "LEVEL SELECT",
		"win_title":      "LEVEL COMPLETE!",
		"lose_title":     "OUT OF MOVES!",
		"lose_sub":       "So close! Try again!",
		"any":            "ANY",
		"done":           "DONE!",
		"locked":         "[LOCKED]",
		"stars":          "Stars",
		"chapter_select": "SELECT CHAPTER",
		"lang_label":     "Language",
		"sound":          "Sound",
		"music":          "Music",
		"notifs":         "Notifications",
		"parental":       "Parental Controls",
		"privacy":        "Privacy Policy",
		"tos":            "Terms of Service",
		"gdpr_export":    "Export My Data",
		"gdpr_delete":    "Delete My Data",
		"account":        "Account",
		"login_google":   "Sign in with Google",
		"login_apple":    "Sign in with Apple",
		"logout":         "Sign Out",
		"shop_title":     "SHOP",
		"buy":            "BUY",
		"coins":          "Coins",
		"free":           "FREE",
		"confirm_buy":    "Confirm Purchase",
		"cancel":         "CANCEL",
		"confirm":        "CONFIRM",
		"insufficient":   "Not enough coins!",
		"pin_setup":      "Set Parental PIN",
		"pin_enter":      "Enter PIN",
		"spend_limit":    "Monthly Spend Limit",
		"age_yes":        "YES, I AM 13+",
		"age_no":         "NO, I'M YOUNGER",
		"purchase_limit": "Spend Limit Reached",
		"watch_ad":       "Watch Ad for +3 Moves",
		"no_ads":         "Remove Ads",
		"odds_title":     "Drop Rates",
		"leaderboard":    "Leaderboard",
		"rank":           "Rank",
		"sync":           "Syncing...",
		"offline":        "Offline",
		"on":             "ON",
		"off":            "OFF",
	},
	LangRU: {
		"level":          "UROVENY",
		"score":          "OCHKI",
		"moves":          "HODY",
		"goal":           "TSEL",
		"combo":          "KOMBO",
		"play":           "IGRAT",
		"continue":       "PRODOLZHAT",
		"settings":       "NASTROJKI",
		"shop":           "MAGAZIN",
		"back":           "< NAZAD",
		"retry":          "SNOVA",
		"next_level":     "SLEDUJUSCHIJ >>",
		"level_select":   "VYBOR UROVNJA",
		"win_title":      "UROVEN PROJDEN!",
		"lose_title":     "HODY KONCHILIS!",
		"lose_sub":       "Pochti! Eshche raz!",
		"any":            "LUBOJ",
		"done":           "GOTOVO!",
		"locked":         "[ZABLOK]",
		"stars":          "Zvezdy",
		"chapter_select": "VYBOR BIOMA",
		"lang_label":     "Jazyk",
		"sound":          "Zvuk",
		"music":          "Muzyka",
		"notifs":         "Uvedomlenija",
		"parental":       "Rod. kontrol",
		"privacy":        "Konf-ost",
		"tos":            "Polz. soglaschenie",
		"gdpr_export":    "Eksport dannyh",
		"gdpr_delete":    "Udalit dannye",
		"account":        "Akkaunt",
		"login_google":   "Vojti cherez Google",
		"login_apple":    "Vojti cherez Apple",
		"logout":         "Vyjti",
		"shop_title":     "MAGAZIN",
		"buy":            "KUPIT",
		"coins":          "Monety",
		"free":           "BESPLATNO",
		"confirm_buy":    "Podtverdit pokupku",
		"cancel":         "OTMENA",
		"confirm":        "PODTVERDIT",
		"insufficient":   "Nedostatochno monet!",
		"pin_setup":      "Nastroit PIN",
		"pin_enter":      "Vvesti PIN",
		"spend_limit":    "Limit trat v mesyaz",
		"age_yes":        "DA, MNE 13+",
		"age_no":         "NET, JA MOLZHE",
		"purchase_limit": "Limit trat dostignut",
		"watch_ad":       "Reklama: +3 hoda",
		"no_ads":         "Otkl. reklamu",
		"odds_title":     "Shansy",
		"leaderboard":    "Rejting",
		"rank":           "Mesto",
		"sync":           "Sinhronizacija...",
		"offline":        "Offlajn",
		"on":             "VKL",
		"off":            "OTKL",
	},
	LangDE: {
		"level": "LEVEL", "score": "PUNKTE", "moves": "ZÜGE", "goal": "ZIEL",
		"play": "SPIELEN", "continue": "WEITER", "settings": "EINSTELLUNGEN",
		"shop": "SHOP", "back": "< ZURÜCK", "retry": "NOCHMAL",
		"next_level": "WEITER >>", "win_title": "LEVEL ABGESCHLOSSEN!",
		"lose_title": "KEINE ZÜGE!", "buy": "KAUFEN", "cancel": "ABBRECHEN",
		"confirm": "BESTÄTIGEN", "on": "AN", "off": "AUS",
	},
	LangFR: {
		"level": "NIVEAU", "score": "SCORE", "moves": "COUPS", "goal": "OBJECTIF",
		"play": "JOUER", "continue": "CONTINUER", "settings": "PARAMÈTRES",
		"shop": "BOUTIQUE", "back": "< RETOUR", "retry": "RÉESSAYER",
		"next_level": "SUIVANT >>", "win_title": "NIVEAU RÉUSSI!",
		"lose_title": "PLUS DE COUPS!", "buy": "ACHETER", "cancel": "ANNULER",
		"confirm": "CONFIRMER", "on": "OUI", "off": "NON",
	},
	LangES: {
		"level": "NIVEL", "score": "PUNTOS", "moves": "MOVS", "goal": "META",
		"play": "JUGAR", "continue": "CONTINUAR", "settings": "AJUSTES",
		"shop": "TIENDA", "back": "< ATRÁS", "retry": "REINTENTAR",
		"next_level": "SIGUIENTE >>", "win_title": "¡NIVEL COMPLETADO!",
		"lose_title": "¡SIN MOVIMIENTOS!", "buy": "COMPRAR", "cancel": "CANCELAR",
		"confirm": "CONFIRMAR", "on": "SÍ", "off": "NO",
	},
	LangZH: {
		"level": "LEVEL", "score": "SCORE", "moves": "MOVES", "goal": "GOAL",
		"play": "PLAY", "continue": "CONTINUE", "settings": "SETTINGS",
		"shop": "SHOP", "back": "< BACK", "on": "ON", "off": "OFF",
	},
}
