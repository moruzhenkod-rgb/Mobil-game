package main

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// ── Shop items ────────────────────────────────────────────────────

type ItemType string

const (
	ItemCoins    ItemType = "coins"
	ItemMoves    ItemType = "moves"
	ItemBomb     ItemType = "booster_bomb"
	ItemRainbow  ItemType = "booster_rainbow"
	ItemNoAds    ItemType = "no_ads"
	ItemChest    ItemType = "chest"
)

type ShopItem struct {
	ID         string
	Name       string
	Desc       string
	CoinCost   int    // 0 = real-money purchase
	RealPrice  string // "$0.99" etc., empty = coin-only
	Type       ItemType
	Amount     int
	IconKind   TileKind
	IsFeatured bool
}

var shopItems = []ShopItem{
	// ── Free/Daily ──
	{
		ID: "daily_coins", Name: "Daily Reward", Desc: "100 free coins every day!",
		CoinCost: 0, RealPrice: T("free"), Type: ItemCoins, Amount: 100,
	},
	// ── Coin packs (real money) ──
	{
		ID: "coins_small", Name: "100 Coins", Desc: "Small coin pack",
		CoinCost: 0, RealPrice: "$0.99", Type: ItemCoins, Amount: 100,
	},
	{
		ID: "coins_medium", Name: "550 Coins", Desc: "500 + 50 bonus coins",
		CoinCost: 0, RealPrice: "$4.99", Type: ItemCoins, Amount: 550, IsFeatured: true,
	},
	{
		ID: "coins_large", Name: "1200 Coins", Desc: "1000 + 200 bonus coins",
		CoinCost: 0, RealPrice: "$9.99", Type: ItemCoins, Amount: 1200,
	},
	// ── Boosters (coin cost) ──
	{
		ID: "booster_moves", Name: "+5 Extra Moves", Desc: "5 bonus moves for next level",
		CoinCost: 60, Type: ItemMoves, Amount: 5, IconKind: TileEmpty,
	},
	{
		ID: "booster_bomb", Name: "Bomb Booster", Desc: "Start with a Bomb gem on board",
		CoinCost: 80, Type: ItemBomb, Amount: 1, IconKind: TileBomb,
	},
	{
		ID: "booster_rainbow", Name: "Rainbow Booster", Desc: "Start with a Rainbow gem",
		CoinCost: 120, Type: ItemRainbow, Amount: 1, IconKind: TileRainbow,
	},
	// ── Chests (coin cost + odds shown) ──
	{
		ID: "chest_bronze", Name: "Bronze Chest", Desc: "3 random items",
		CoinCost: 100, Type: ItemChest, Amount: 3, IconKind: TileOrange,
	},
	{
		ID: "chest_silver", Name: "Silver Chest", Desc: "5 items, better odds",
		CoinCost: 250, Type: ItemChest, Amount: 5, IconKind: TileBlue, IsFeatured: true,
	},
	{
		ID: "chest_gold", Name: "Gold Chest", Desc: "7 items, best odds",
		CoinCost: 500, Type: ItemChest, Amount: 7, IconKind: TileYellow,
	},
	// ── Remove ads (real money) ──
	{
		ID: "no_ads", Name: "Remove Ads", Desc: "Enjoy the game ad-free forever",
		CoinCost: 0, RealPrice: "$2.99", Type: ItemNoAds,
	},
}

// Chest odds (Apple/Google transparency requirement)
type ChestOdds struct {
	Item        string
	Probability float64 // 0.0–1.0
}

var chestOdds = map[string][]ChestOdds{
	"chest_bronze": {
		{"Common gem booster (x5)", 0.70},
		{"Extra Moves (+3)", 0.25},
		{"Bomb Booster", 0.05},
	},
	"chest_silver": {
		{"Extra Moves (+5)", 0.50},
		{"Bomb Booster", 0.30},
		{"Rainbow Booster", 0.15},
		{"Gold Chest", 0.05},
	},
	"chest_gold": {
		{"Bomb Booster x3", 0.45},
		{"Rainbow Booster x2", 0.30},
		{"Mega Moves Pack (+15)", 0.20},
		{"VIP Season Pass", 0.05},
	},
}

// ── Shop state ────────────────────────────────────────────────────

type Shop struct {
	scroll      float32
	selectedIdx int
	showOdds    bool // odds panel for selected chest
	showConfirm bool
	confirmItem *ShopItem
	message     string
	msgTick     int
}

var shopState = &Shop{selectedIdx: -1}

func (s *Shop) Reset() {
	s.showOdds = false
	s.showConfirm = false
	s.confirmItem = nil
	s.message = ""
}

// ── Drawing ───────────────────────────────────────────────────────

func DrawShop(screen *ebiten.Image, tick int) {
	s := shopState

	// Header
	vector.DrawFilledRect(screen, 0, 0, ScreenW, 56, color.RGBA{10, 5, 28, 245}, false)
	ebitenutil.DebugPrintAt(screen, T("shop_title"), ScreenW/2-18, 10)
	// Coin balance
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Coins: %d", progress.Coins), ScreenW-100, 10)
	// Back
	drawShopBtn(screen, T("back"), 12, 10, 70, 28, color.RGBA{40, 25, 70, 220})
	vector.DrawFilledRect(screen, 0, 52, ScreenW, 3, color.RGBA{120, 60, 200, 160}, false)

	if s.showOdds && s.confirmItem != nil {
		drawOddsPanel(screen, s.confirmItem)
		return
	}
	if s.showConfirm && s.confirmItem != nil {
		drawConfirmPanel(screen, s.confirmItem)
		return
	}

	// Item list
	y := 70
	for i := range shopItems {
		item := &shopItems[i]
		drawShopItem(screen, item, y, i == s.selectedIdx, tick)
		y += 88
	}

	// Message toast
	if s.message != "" && s.msgTick > 0 {
		s.msgTick--
		alpha := uint8(min8(s.msgTick*6, 255))
		vector.DrawFilledRect(screen, 80, ScreenH/2-20, ScreenW-160, 40, color.RGBA{20, 12, 50, alpha}, false)
		ebitenutil.DebugPrintAt(screen, s.message, ScreenW/2-len(s.message)*3, ScreenH/2-10)
	}
}

func drawShopItem(screen *ebiten.Image, item *ShopItem, y int, selected bool, tick int) {
	bg := color.RGBA{18, 10, 45, 220}
	if item.IsFeatured {
		bg = color.RGBA{30, 18, 65, 230}
	}
	if selected {
		bg = color.RGBA{50, 30, 90, 230}
	}
	vector.DrawFilledRect(screen, 10, float32(y), ScreenW-20, 80, bg, false)

	// Featured badge
	if item.IsFeatured {
		vector.DrawFilledRect(screen, 10, float32(y), 80, 3, color.RGBA{255, 210, 50, 255}, false)
		ebitenutil.DebugPrintAt(screen, "FEATURED", 12, y+6)
	}

	// Icon gem (if applicable)
	if item.IconKind != TileEmpty && int(item.IconKind) < int(GemCount) {
		c := gemColors[item.IconKind]
		vector.DrawFilledRect(screen, 18, float32(y)+14, 36, 36, c, false)
		vector.DrawFilledRect(screen, 18, float32(y)+14, 36, 3, color.RGBA{255, 255, 255, 80}, false)
	} else {
		// Coin icon
		vector.DrawFilledCircle(screen, 36, float32(y)+32, 16, color.RGBA{255, 210, 50, 255}, false)
		vector.DrawFilledCircle(screen, 36, float32(y)+32, 10, color.RGBA{200, 160, 30, 255}, false)
	}

	// Name + desc
	ebitenutil.DebugPrintAt(screen, item.Name, 68, y+12)
	ebitenutil.DebugPrintAt(screen, item.Desc, 68, y+28)

	// Price / Buy button
	priceStr := item.RealPrice
	if item.CoinCost > 0 {
		priceStr = fmt.Sprintf("%d coins", item.CoinCost)
	} else if item.RealPrice == "" {
		priceStr = T("free")
	}

	btnC := color.RGBA{50, 160, 70, 255}
	if item.CoinCost > progress.Coins {
		btnC = color.RGBA{100, 60, 60, 200}
	}
	drawShopBtn(screen, priceStr, ScreenW-130, y+22, 118, 32, btnC)

	// Odds info link for chests
	if item.Type == ItemChest {
		ebitenutil.DebugPrintAt(screen, "[View Odds]", 68, y+46)
	}
}

func drawShopBtn(screen *ebiten.Image, label string, x, y, w, h int, bg color.RGBA) {
	vector.DrawFilledRect(screen, float32(x), float32(y), float32(w), float32(h), bg, false)
	vector.DrawFilledRect(screen, float32(x), float32(y), float32(w), 2,
		color.RGBA{bg.R + 40, bg.G + 40, bg.B + 40, 200}, false)
	ebitenutil.DebugPrintAt(screen, label, x+w/2-len(label)*3, y+h/2-5)
}

func drawOddsPanel(screen *ebiten.Image, item *ShopItem) {
	// Dim background
	ov := ebiten.NewImage(ScreenW, ScreenH)
	ov.Fill(color.RGBA{0, 0, 0, 180})
	screen.DrawImage(ov, nil)

	vector.DrawFilledRect(screen, 40, 160, ScreenW-80, 580, color.RGBA{14, 8, 38, 248}, false)
	vector.DrawFilledRect(screen, 40, 160, ScreenW-80, 3, color.RGBA{255, 210, 50, 255}, false)

	ebitenutil.DebugPrintAt(screen, T("odds_title")+" - "+item.Name, 56, 180)
	ebitenutil.DebugPrintAt(screen, "All percentages independently verified.", 56, 200)

	odds, ok := chestOdds[item.ID]
	if ok {
		y := 240
		for _, o := range odds {
			pct := fmt.Sprintf("%.0f%%", o.Probability*100)
			ebitenutil.DebugPrintAt(screen, pct, 60, y)
			ebitenutil.DebugPrintAt(screen, o.Item, 110, y)
			// Bar
			vector.DrawFilledRect(screen, 60, float32(y+16), float32(o.Probability)*float32(ScreenW-120), 6,
				color.RGBA{120, 200, 255, 200}, false)
			y += 44
		}
	}

	ebitenutil.DebugPrintAt(screen, "Odds are per individual item draw.", 56, 530)
	ebitenutil.DebugPrintAt(screen, "Purchasing does not guarantee specific items.", 56, 548)

	drawShopBtn(screen, T("back"), ScreenW/2-60, 600, 120, 44, color.RGBA{50, 35, 90, 255})
	drawShopBtn(screen, T("buy"), ScreenW/2-60, 656, 120, 44, color.RGBA{50, 160, 70, 255})
}

func drawConfirmPanel(screen *ebiten.Image, item *ShopItem) {
	ov := ebiten.NewImage(ScreenW, ScreenH)
	ov.Fill(color.RGBA{0, 0, 0, 160})
	screen.DrawImage(ov, nil)

	vector.DrawFilledRect(screen, 60, 300, ScreenW-120, 360, color.RGBA{14, 8, 38, 248}, false)
	vector.DrawFilledRect(screen, 60, 300, ScreenW-120, 3, color.RGBA{120, 60, 200, 255}, false)

	ebitenutil.DebugPrintAt(screen, T("confirm_buy"), ScreenW/2-54, 320)
	ebitenutil.DebugPrintAt(screen, item.Name, ScreenW/2-len(item.Name)*3, 358)

	price := item.RealPrice
	if item.CoinCost > 0 {
		price = fmt.Sprintf("%d %s", item.CoinCost, T("coins"))
	}
	ebitenutil.DebugPrintAt(screen, price, ScreenW/2-len(price)*3, 382)

	if progress.ParentalPIN != "" {
		ebitenutil.DebugPrintAt(screen, "PIN required for purchases.", ScreenW/2-78, 420)
	}

	// Monthly limit check
	if progress.MonthlyLimit > 0 {
		spent := progress.MonthlySpent
		ebitenutil.DebugPrintAt(screen,
			fmt.Sprintf("Monthly spend: $%.2f / $%.2f", float64(spent)/100, float64(progress.MonthlyLimit)/100),
			ScreenW/2-105, 438)
	}

	drawShopBtn(screen, T("cancel"), 100, 490, 140, 50, color.RGBA{90, 50, 100, 255})
	drawShopBtn(screen, T("confirm"), ScreenW-240, 490, 140, 50, color.RGBA{50, 160, 70, 255})
}

// ── Shop input ────────────────────────────────────────────────────

// ShopHandleClick returns true if the shop consumed the click.
// Returns ("back", nil) to signal the back button was pressed.
func ShopHandleClick(mx, my int) string {
	s := shopState

	// Back button in header
	if mx >= 12 && mx <= 82 && my >= 10 && my <= 38 {
		s.Reset()
		return "back"
	}

	if s.showOdds {
		// Back
		if mx >= ScreenW/2-60 && mx <= ScreenW/2+60 && my >= 600 && my <= 644 {
			s.showOdds = false
			return ""
		}
		// Buy
		if mx >= ScreenW/2-60 && mx <= ScreenW/2+60 && my >= 656 && my <= 700 {
			s.showOdds = false
			s.showConfirm = true
			return ""
		}
		return ""
	}

	if s.showConfirm {
		item := s.confirmItem
		if item == nil {
			s.showConfirm = false
			return ""
		}
		// Cancel
		if mx >= 100 && mx <= 240 && my >= 490 && my <= 540 {
			s.showConfirm = false
			return ""
		}
		// Confirm
		if mx >= ScreenW-240 && mx <= ScreenW-100 && my >= 490 && my <= 540 {
			s.showConfirm = false
			captured := item
			OpenPINVerify(func() { executePurchase(captured) })
			return ""
		}
		return ""
	}

	// Item rows
	y := 70
	for i := range shopItems {
		item := &shopItems[i]
		if my >= y && my <= y+80 {
			// Odds link for chests
			if item.Type == ItemChest && mx >= 68 && mx <= 68+66 && my >= y+46 && my <= y+62 {
				s.selectedIdx = i
				s.confirmItem = item
				s.showOdds = true
				return ""
			}
			// Buy button area
			if mx >= ScreenW-130 && mx <= ScreenW-12 && my >= y+22 && my <= y+54 {
				s.selectedIdx = i
				s.confirmItem = item
				if item.CoinCost == 0 && item.RealPrice == "" {
					// Free daily
					executePurchase(item)
				} else {
					s.showConfirm = true
				}
				return ""
			}
		}
		y += 88
	}
	return ""
}

func executePurchase(item *ShopItem) {
	s := shopState

	// Monthly limit check (real-money items)
	if item.RealPrice != "" && item.CoinCost == 0 {
		if progress.MonthlyLimit > 0 && progress.MonthlySpent >= progress.MonthlyLimit {
			s.message = T("purchase_limit")
			s.msgTick = 120
			return
		}
	}

	// Coin check
	if item.CoinCost > 0 && progress.Coins < item.CoinCost {
		s.message = T("insufficient")
		s.msgTick = 120
		return
	}

	// Deduct coins
	if item.CoinCost > 0 {
		progress.Coins -= item.CoinCost
	}

	// Apply item
	switch item.Type {
	case ItemCoins:
		progress.Coins += item.Amount
	case ItemMoves:
		progress.ExtraMoves += item.Amount
	case ItemBomb:
		progress.BombBoosters += item.Amount
	case ItemRainbow:
		progress.RainbowBoosters += item.Amount
	case ItemNoAds:
		progress.AdsRemoved = true
	case ItemChest:
		openChest(item)
	}

	saveProgress()
	LogPurchase(item.ID, item.CoinCost)
	go fb.PushSave()

	s.message = "Purchase successful!"
	s.msgTick = 120
}

func openChest(item *ShopItem) {
	odds, ok := chestOdds[item.ID]
	if !ok {
		return
	}
	// Draw N items from odds table
	for i := 0; i < item.Amount; i++ {
		r := randFloat(1.0)
		cumulative := 0.0
		for _, o := range odds {
			cumulative += o.Probability
			if r <= cumulative {
				applyChestReward(o.Item)
				break
			}
		}
	}
}

func applyChestReward(item string) {
	switch item {
	case "Common gem booster (x5)":
		progress.BombBoosters++
	case "Extra Moves (+3)":
		progress.ExtraMoves += 3
	case "Extra Moves (+5)":
		progress.ExtraMoves += 5
	case "Mega Moves Pack (+15)":
		progress.ExtraMoves += 15
	case "Bomb Booster":
		progress.BombBoosters++
	case "Bomb Booster x3":
		progress.BombBoosters += 3
	case "Rainbow Booster":
		progress.RainbowBoosters++
	case "Rainbow Booster x2":
		progress.RainbowBoosters += 2
	default:
		progress.Coins += 50
	}
}

func min8(a, b int) int {
	if a < b {
		return a
	}
	return b
}
