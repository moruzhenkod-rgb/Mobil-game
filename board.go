package main

import (
	"image/color"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// ── Board ─────────────────────────────────────────────────────────

type Board struct {
	tiles [Rows][Cols]*Tile

	selRow, selCol int
	hasSel         bool

	swapR1, swapC1 int
	swapR2, swapC2 int
	swapping        bool

	exploding bool
	falling   bool
	combo     int

	pendingSpawns []matchGroup

	// Stats
	MovesUsed     int
	Score         int
	ClearedByKind [GemCount]int
	ClearedIce    int
	ClearedStone  int
	ClearedChain  int
}

func NewBoard(l Level) *Board {
	b := &Board{selRow: -1, selCol: -1}
	b.spawnAll()
	b.clearInitialMatches()
	b.placeObstacles(l)
	return b
}

// ── tile factory ──────────────────────────────────────────────────

func (b *Board) spawnAll() {
	for r := 0; r < Rows; r++ {
		for c := 0; c < Cols; c++ {
			if b.tiles[r][c] == nil {
				k := b.safeGem(r, c)
				b.tiles[r][c] = makeTile(k, TileEmpty, r, c, false)
			}
		}
	}
}

func makeTile(kind, base TileKind, row, col int, fromAbove bool) *Tile {
	tx, ty := tilePixelPos(row, col)
	sy := float64(ty)
	if fromAbove {
		sy = float64(BoardOffsetY - TileOuter*(Rows+1))
	}
	return &Tile{
		Kind: kind, BaseColor: base, Layers: 0,
		X: float64(tx), Y: sy,
		TargetX: float64(tx), TargetY: float64(ty),
	}
}

func tilePixelPos(row, col int) (int, int) {
	return BoardOffsetX + col*TileOuter, BoardOffsetY + row*TileOuter
}

func (b *Board) safeGem(r, c int) TileKind {
	for {
		k := TileKind(rand.Intn(int(GemCount-1)) + 1)
		if c >= 2 && b.matchKindAt(r, c-1) == k && b.matchKindAt(r, c-2) == k {
			continue
		}
		if r >= 2 && b.matchKindAt(r-1, c) == k && b.matchKindAt(r-2, c) == k {
			continue
		}
		return k
	}
}

func (b *Board) clearInitialMatches() {
	for {
		groups := b.findMatchGroups()
		if len(groups) == 0 {
			break
		}
		for _, g := range groups {
			for _, p := range g.cells {
				b.tiles[p[0]][p[1]].Kind = TileKind(rand.Intn(int(GemCount-1)) + 1)
				b.tiles[p[0]][p[1]].BaseColor = TileEmpty
			}
		}
	}
}

// ── obstacle placement ────────────────────────────────────────────

func (b *Board) placeObstacles(l Level) {
	placed := 0
	tries := 0
	positions := b.shuffledPositions()

	for _, p := range positions {
		r, c := p[0], p[1]
		if placed >= l.IceCount {
			break
		}
		t := b.tiles[r][c]
		if t == nil || t.Kind.IsObstacle() {
			continue
		}
		layers := l.IceLayers
		if layers < 1 {
			layers = 1
		}
		base := t.Kind
		b.tiles[r][c] = makeTile(TileIce, base, r, c, false)
		b.tiles[r][c].Layers = layers
		placed++
		tries++
	}

	placed = 0
	for _, p := range positions {
		r, c := p[0], p[1]
		if placed >= l.StoneCount {
			break
		}
		if b.tiles[r][c] != nil && b.tiles[r][c].Kind.IsObstacle() {
			continue
		}
		b.tiles[r][c] = makeTile(TileStone, TileEmpty, r, c, false)
		placed++
	}

	placed = 0
	for _, p := range b.shuffledPositions() {
		r, c := p[0], p[1]
		if placed >= l.ChainCount {
			break
		}
		t := b.tiles[r][c]
		if t == nil || t.Kind.IsObstacle() {
			continue
		}
		base := t.Kind
		b.tiles[r][c] = makeTile(TileChain, base, r, c, false)
		placed++
	}
	_ = tries
}

func (b *Board) shuffledPositions() [][2]int {
	all := make([][2]int, 0, Rows*Cols)
	for r := 0; r < Rows; r++ {
		for c := 0; c < Cols; c++ {
			all = append(all, [2]int{r, c})
		}
	}
	rand.Shuffle(len(all), func(i, j int) { all[i], all[j] = all[j], all[i] })
	return all
}

// ── helpers ───────────────────────────────────────────────────────

func (b *Board) matchKindAt(r, c int) TileKind {
	if r < 0 || r >= Rows || c < 0 || c >= Cols || b.tiles[r][c] == nil {
		return TileEmpty
	}
	t := b.tiles[r][c]
	// Obstacles don't participate in matches
	if t.Kind.IsObstacle() {
		return TileEmpty
	}
	if t.Kind == TileRainbow {
		return t.BaseColor
	}
	if t.Kind.IsBonus() {
		return t.BaseColor
	}
	return t.Kind
}

func (b *Board) IsBusy() bool {
	if b.swapping || b.exploding || b.falling {
		return true
	}
	for r := 0; r < Rows; r++ {
		for c := 0; c < Cols; c++ {
			if b.tiles[r][c] != nil && b.tiles[r][c].IsAnimating() {
				return true
			}
		}
	}
	return false
}

func (b *Board) Tiles(r, c int) *Tile {
	return b.tiles[r][c]
}

// ── Update ────────────────────────────────────────────────────────

func (b *Board) Update() (scoreGained int, matchFired bool) {
	for r := 0; r < Rows; r++ {
		for c := 0; c < Cols; c++ {
			if b.tiles[r][c] != nil {
				b.tiles[r][c].Update()
			}
		}
	}

	if b.swapping {
		t1 := b.tiles[b.swapR1][b.swapC1]
		t2 := b.tiles[b.swapR2][b.swapC2]
		if (t1 == nil || !t1.Swapping) && (t2 == nil || !t2.Swapping) {
			b.swapping = false
			groups := b.findMatchGroups()
			if len(groups) == 0 {
				b.swapNoAnim(b.swapR1, b.swapC1, b.swapR2, b.swapC2)
				b.combo = 0
			} else {
				b.combo = 1
				gained := b.startExplode(groups)
				return gained, true
			}
		}
	}

	if b.exploding {
		done := true
		for r := 0; r < Rows; r++ {
			for c := 0; c < Cols; c++ {
				t := b.tiles[r][c]
				if t != nil && t.Exploding && t.ExplodeTick < ExplodeDur {
					done = false
				}
			}
		}
		if done {
			b.exploding = false
			b.removeExploded()
			b.applyGravity()
			b.falling = true
		}
	}

	if b.falling {
		settled := true
		for r := 0; r < Rows; r++ {
			for c := 0; c < Cols; c++ {
				if b.tiles[r][c] != nil && b.tiles[r][c].IsFalling() {
					settled = false
				}
			}
		}
		if settled {
			b.falling = false
			b.spawnMissing()
			groups := b.findMatchGroups()
			if len(groups) > 0 {
				b.combo++
				gained := b.startExplode(groups)
				return gained, true
			}
			b.combo = 0
		}
	}

	return 0, false
}

// ── Input ─────────────────────────────────────────────────────────

func (b *Board) TrySelectPixel(px, py int) bool {
	if b.IsBusy() {
		return false
	}
	col := (px - BoardOffsetX) / TileOuter
	row := (py - BoardOffsetY) / TileOuter
	if col < 0 || col >= Cols || row < 0 || row >= Rows {
		b.clearSel()
		return false
	}
	return b.trySelect(row, col)
}

func (b *Board) trySelect(row, col int) bool {
	t := b.tiles[row][col]
	if t == nil || t.Kind == TileEmpty || t.Kind.IsObstacle() {
		b.clearSel()
		return false
	}
	if !b.hasSel {
		b.hasSel = true
		b.selRow, b.selCol = row, col
		t.Selected = true
		return false
	}
	if b.selRow == row && b.selCol == col {
		b.clearSel()
		return false
	}
	dr, dc := row-b.selRow, col-b.selCol
	if (dr == 0 && (dc == 1 || dc == -1)) || (dc == 0 && (dr == 1 || dr == -1)) {
		b.tiles[b.selRow][b.selCol].Selected = false
		b.hasSel = false
		b.MovesUsed++
		b.initiateSwap(b.selRow, b.selCol, row, col)
		return true
	}
	b.tiles[b.selRow][b.selCol].Selected = false
	b.selRow, b.selCol = row, col
	t.Selected = true
	return false
}

func (b *Board) clearSel() {
	if b.hasSel && b.selRow >= 0 {
		if t := b.tiles[b.selRow][b.selCol]; t != nil {
			t.Selected = false
		}
	}
	b.hasSel = false
	b.selRow, b.selCol = -1, -1
}

// ── Swap ──────────────────────────────────────────────────────────

func (b *Board) initiateSwap(r1, c1, r2, c2 int) {
	b.swapR1, b.swapC1 = r1, c1
	b.swapR2, b.swapC2 = r2, c2
	b.swapping = true
	t1, t2 := b.tiles[r1][c1], b.tiles[r2][c2]
	b.tiles[r1][c1], b.tiles[r2][c2] = t2, t1
	tx1, ty1 := tilePixelPos(r1, c1)
	tx2, ty2 := tilePixelPos(r2, c2)
	anim := func(t *Tile, toX, toY int) {
		if t == nil {
			return
		}
		t.StartX, t.StartY = t.X, t.Y
		t.TargetX, t.TargetY = float64(toX), float64(toY)
		t.Swapping = true
		t.SwapTick = 0
	}
	anim(t1, tx2, ty2)
	anim(t2, tx1, ty1)
}

func (b *Board) swapNoAnim(r1, c1, r2, c2 int) {
	b.tiles[r1][c1], b.tiles[r2][c2] = b.tiles[r2][c2], b.tiles[r1][c1]
	for _, rc := range [][2]int{{r1, c1}, {r2, c2}} {
		r, c := rc[0], rc[1]
		if t := b.tiles[r][c]; t != nil {
			tx, ty := tilePixelPos(r, c)
			t.X, t.Y = float64(tx), float64(ty)
			t.TargetX, t.TargetY = float64(tx), float64(ty)
			t.Swapping = false
		}
	}
}

// ── Match detection ───────────────────────────────────────────────

type matchGroup struct {
	kind           TileKind
	cells          [][2]int
	bonus          TileKind
	bonusR, bonusC int
}

func (b *Board) findMatchGroups() []matchGroup {
	var mark [Rows][Cols]bool

	for r := 0; r < Rows; r++ {
		c := 0
		for c < Cols {
			k := b.matchKindAt(r, c)
			if k == TileEmpty {
				c++
				continue
			}
			end := c + 1
			for end < Cols && b.matchKindAt(r, end) == k {
				end++
			}
			if end-c >= 3 {
				for i := c; i < end; i++ {
					mark[r][i] = true
				}
			}
			c = end
		}
	}
	for c := 0; c < Cols; c++ {
		r := 0
		for r < Rows {
			k := b.matchKindAt(r, c)
			if k == TileEmpty {
				r++
				continue
			}
			end := r + 1
			for end < Rows && b.matchKindAt(end, c) == k {
				end++
			}
			if end-r >= 3 {
				for i := r; i < end; i++ {
					mark[i][c] = true
				}
			}
			r = end
		}
	}

	hasAny := false
	for r := 0; r < Rows; r++ {
		for c := 0; c < Cols; c++ {
			if mark[r][c] {
				hasAny = true
			}
		}
	}
	if !hasAny {
		return nil
	}

	var vis [Rows][Cols]bool
	var groups []matchGroup
	for r := 0; r < Rows; r++ {
		for c := 0; c < Cols; c++ {
			if mark[r][c] && !vis[r][c] {
				groups = append(groups, b.floodGroup(r, c, mark, &vis))
			}
		}
	}
	return groups
}

func (b *Board) floodGroup(sr, sc int, mark [Rows][Cols]bool, vis *[Rows][Cols]bool) matchGroup {
	queue := [][2]int{{sr, sc}}
	vis[sr][sc] = true
	var cells [][2]int
	kmap := map[TileKind]int{}

	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		r, c := cur[0], cur[1]
		cells = append(cells, cur)
		kmap[b.matchKindAt(r, c)]++
		for _, d := range [][2]int{{-1, 0}, {1, 0}, {0, -1}, {0, 1}} {
			nr, nc := r+d[0], c+d[1]
			if nr >= 0 && nr < Rows && nc >= 0 && nc < Cols && mark[nr][nc] && !vis[nr][nc] {
				vis[nr][nc] = true
				queue = append(queue, [2]int{nr, nc})
			}
		}
	}

	dom := TileEmpty
	best := 0
	for k, cnt := range kmap {
		if cnt > best {
			best, dom = cnt, k
		}
	}

	n := len(cells)
	br, bc := cells[n/2][0], cells[n/2][1]

	rows, cols := map[int]int{}, map[int]int{}
	for _, p := range cells {
		rows[p[0]]++
		cols[p[1]]++
	}
	maxR, maxC := 0, 0
	for _, v := range rows {
		if v > maxR {
			maxR = v
		}
	}
	for _, v := range cols {
		if v > maxC {
			maxC = v
		}
	}

	bonus := TileEmpty
	switch {
	case n >= 5:
		bonus = TileRainbow
	case n == 4 && maxR == 4:
		bonus = TileRowBomb
	case n == 4 && maxC == 4:
		bonus = TileColBomb
	case n >= 4:
		bonus = TileBomb
	}

	return matchGroup{kind: dom, cells: cells, bonus: bonus, bonusR: br, bonusC: bc}
}

// ── Explosion ─────────────────────────────────────────────────────

func (b *Board) startExplode(groups []matchGroup) int {
	hit := map[[2]int]bool{}
	for _, g := range groups {
		for _, p := range g.cells {
			hit[[2]int{p[0], p[1]}] = true
		}
	}

	// Cascade bonuses inside the hit set
	b.expandBonuses(hit)

	// Damage adjacent obstacles
	adjacentObstacleDmg := map[[2]int]bool{}
	for p := range hit {
		for _, d := range [][2]int{{-1, 0}, {1, 0}, {0, -1}, {0, 1}} {
			nr, nc := p[0]+d[0], p[1]+d[1]
			if nr < 0 || nr >= Rows || nc < 0 || nc >= Cols {
				continue
			}
			t := b.tiles[nr][nc]
			if t == nil {
				continue
			}
			switch t.Kind {
			case TileIce:
				adjacentObstacleDmg[[2]int{nr, nc}] = true
			case TileChain:
				adjacentObstacleDmg[[2]int{nr, nc}] = true
			}
		}
	}

	// Apply adjacent damage
	for p := range adjacentObstacleDmg {
		if hit[p] { // already in blast, skip
			continue
		}
		r, c := p[0], p[1]
		t := b.tiles[r][c]
		if t == nil {
			continue
		}
		switch t.Kind {
		case TileIce:
			t.DmgFlash = 8
			t.Layers--
			if t.Layers <= 0 {
				hit[p] = true // explode it
			}
		case TileChain:
			t.DmgFlash = 8
			// Convert to normal gem
			t.Kind = t.BaseColor
			t.BaseColor = TileEmpty
			b.ClearedChain++
		}
	}

	// Mark all hit tiles for explosion
	for p := range hit {
		r, c := p[0], p[1]
		t := b.tiles[r][c]
		if t != nil && !t.Exploding {
			t.Exploding = true
			t.ExplodeTick = 0
			// Track stats
			switch t.Kind {
			case TileIce:
				b.ClearedIce++
			case TileStone:
				b.ClearedStone++
			case TileChain:
				b.ClearedChain++
			default:
				baseK := t.Kind
				if t.Kind.IsBonus() {
					baseK = t.BaseColor
				}
				if int(baseK) >= int(TileRed) && int(baseK) < int(GemCount) {
					b.ClearedByKind[baseK]++
				}
			}
		}
	}

	b.pendingSpawns = groups
	b.exploding = true

	return len(hit)*10 + (b.combo-1)*25
}

func (b *Board) expandBonuses(hit map[[2]int]bool) {
	changed := true
	for changed {
		changed = false
		for p := range hit {
			t := b.tiles[p[0]][p[1]]
			if t == nil || !t.Kind.IsBonus() {
				continue
			}
			r, c := p[0], p[1]
			switch t.Kind {
			case TileRowBomb:
				for col := 0; col < Cols; col++ {
					if !hit[[2]int{r, col}] {
						hit[[2]int{r, col}] = true
						changed = true
					}
				}
			case TileColBomb:
				for row := 0; row < Rows; row++ {
					if !hit[[2]int{row, c}] {
						hit[[2]int{row, c}] = true
						changed = true
					}
				}
			case TileBomb:
				for dr := -1; dr <= 1; dr++ {
					for dc := -1; dc <= 1; dc++ {
						nr, nc := r+dr, c+dc
						if nr >= 0 && nr < Rows && nc >= 0 && nc < Cols && !hit[[2]int{nr, nc}] {
							hit[[2]int{nr, nc}] = true
							changed = true
						}
					}
				}
			case TileRainbow:
				bc := t.BaseColor
				for row := 0; row < Rows; row++ {
					for col := 0; col < Cols; col++ {
						if b.matchKindAt(row, col) == bc && !hit[[2]int{row, col}] {
							hit[[2]int{row, col}] = true
							changed = true
						}
					}
				}
			}
			// Mark bonus as "processed" by converting it to TileEmpty temporarily
			// so we don't re-process it infinitely
			t.Kind = TileEmpty // neutralise (will be exploded anyway)
		}
	}
}

func (b *Board) removeExploded() {
	for r := 0; r < Rows; r++ {
		for c := 0; c < Cols; c++ {
			t := b.tiles[r][c]
			if t != nil && (t.Exploding || t.Kind == TileEmpty) {
				b.tiles[r][c] = nil
			}
		}
	}
	for _, g := range b.pendingSpawns {
		if g.bonus == TileEmpty {
			continue
		}
		r, c := g.bonusR, g.bonusC
		if b.tiles[r][c] == nil {
			b.tiles[r][c] = makeTile(g.bonus, g.kind, r, c, false)
		}
	}
	b.pendingSpawns = b.pendingSpawns[:0]
}

// ── Gravity + refill ──────────────────────────────────────────────

func (b *Board) applyGravity() {
	for c := 0; c < Cols; c++ {
		write := Rows - 1
		for read := Rows - 1; read >= 0; read-- {
			t := b.tiles[read][c]
			if t != nil && !t.Kind.IsObstacle() { // obstacles don't fall
				b.tiles[write][c] = t
				if write != read {
					b.tiles[read][c] = nil
				}
				write--
			} else if t != nil && t.Kind.IsObstacle() {
				// Obstacles stay, just update their target
				_, ty := tilePixelPos(read, c)
				t.TargetY = float64(ty)
				write-- // still take up the slot
			}
		}
		for r := write; r >= 0; r-- {
			b.tiles[r][c] = nil
		}
		for r := 0; r < Rows; r++ {
			if t := b.tiles[r][c]; t != nil {
				_, ty := tilePixelPos(r, c)
				t.TargetY = float64(ty)
			}
		}
	}
}

func (b *Board) spawnMissing() {
	for c := 0; c < Cols; c++ {
		slot := 0
		for r := 0; r < Rows; r++ {
			if b.tiles[r][c] == nil {
				k := b.safeGem(r, c)
				t := makeTile(k, TileEmpty, r, c, true)
				t.Y = float64(BoardOffsetY - TileOuter*(slot+1))
				b.tiles[r][c] = t
				slot++
			}
		}
	}
}

// ── Draw ──────────────────────────────────────────────────────────

func (b *Board) Draw(screen *ebiten.Image) {
	vector.DrawFilledRect(screen,
		float32(BoardOffsetX-TilePadding), float32(BoardOffsetY-TilePadding),
		float32(Cols*TileOuter+TilePadding), float32(Rows*TileOuter+TilePadding),
		color.RGBA{20, 12, 40, 255}, false)

	for r := 0; r < Rows; r++ {
		for c := 0; c < Cols; c++ {
			x := float32(BoardOffsetX + c*TileOuter)
			y := float32(BoardOffsetY + r*TileOuter)
			sc := color.RGBA{35, 22, 65, 220}
			if (r+c)%2 == 0 {
				sc = color.RGBA{42, 28, 75, 220}
			}
			vector.DrawFilledRect(screen, x, y, TileSize, TileSize, sc, false)
		}
	}

	for r := 0; r < Rows; r++ {
		for c := 0; c < Cols; c++ {
			if b.tiles[r][c] != nil {
				b.tiles[r][c].Draw(screen)
			}
		}
	}
}
