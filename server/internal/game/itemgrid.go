package game

import "math"

// itemGrid buckets items by position so pickups and AOI only look at nearby cells; the map
// holds thousands of ambient EXP orbs (spec §6.1).
//
// Mỗi ô là một SLICE, không phải map. Trước đây là `map[uint32]*Item` và pprof cho thấy
// riêng việc duyệt map (`maps.Iter.Next` + `memhash32` + `matchH2`) ăn ~35% CPU server:
// bốn mươi người × 20 lần/giây × hàng nghìn viên là vài triệu lượt tra bảng băm mỗi giây,
// trong khi việc cần làm chỉ là "đọc hết mấy cái trong ô này". Xoá khỏi slice vẫn O(1) nhờ
// mỗi Item nhớ sẵn mình đang nằm ở ô nào, vị trí nào (hoán đổi với phần tử cuối).
type itemGrid struct {
	cell  float64
	n     int
	cells [][]*Item
}

func newItemGrid(size, cell float64) *itemGrid {
	n := max(1, int(math.Ceil(size/cell)))
	return &itemGrid{cell: cell, n: n, cells: make([][]*Item, n*n)}
}

func (ig *itemGrid) index(p Vec) int {
	x := min(max(int(p.X/ig.cell), 0), ig.n-1)
	y := min(max(int(p.Y/ig.cell), 0), ig.n-1)
	return y*ig.n + x
}

// Slot: chỉ số DÀY ĐẶC của vật phẩm, 0.. tăng dần và được dùng lại khi viên cũ biến mất.
// Client dùng nó để nhớ "đã gửi viên này chưa" bằng một slice thay vì một map băm — pprof
// cho thấy riêng việc tra `knownItems[id]` trong vòng lọc tầm nhìn ăn ~25% CPU server.
func (it *Item) Slot() int { return it.slotIdx }

// ItemBySlot trả về viên đang giữ slot này (nil nếu slot trống).
func (g *Game) ItemBySlot(i int) *Item {
	if i < 0 || i >= len(g.itemSlots) {
		return nil
	}
	return g.itemSlots[i]
}

// SlotCount là số slot đã cấp — client cấp mảng nhớ theo con số này.
func (g *Game) SlotCount() int { return len(g.itemSlots) }

// takeSlot cấp cho viên mới một slot dày đặc, ưu tiên dùng lại slot trống.
func (g *Game) takeSlot(it *Item) {
	if n := len(g.freeSlots); n > 0 {
		it.slotIdx = g.freeSlots[n-1]
		g.freeSlots = g.freeSlots[:n-1]
		g.itemSlots[it.slotIdx] = it
		return
	}
	it.slotIdx = len(g.itemSlots)
	g.itemSlots = append(g.itemSlots, it)
}

func (g *Game) freeSlot(it *Item) {
	if it.slotIdx >= 0 && it.slotIdx < len(g.itemSlots) && g.itemSlots[it.slotIdx] == it {
		g.itemSlots[it.slotIdx] = nil
		g.freeSlots = append(g.freeSlots, it.slotIdx)
	}
	it.slotIdx = -1
}

func (ig *itemGrid) add(it *Item) {
	i := ig.index(it.Pos)
	it.cell, it.slot = i, len(ig.cells[i])
	ig.cells[i] = append(ig.cells[i], it)
}

func (ig *itemGrid) remove(it *Item) {
	c := ig.cells[it.cell]
	last := len(c) - 1
	if it.slot < 0 || it.slot > last || c[it.slot] != it {
		return // đã bị gỡ rồi
	}
	c[it.slot] = c[last]
	c[it.slot].slot = it.slot
	// XOÁ Ô CUỐI trước khi cắt. Thiếu dòng này là **rò rỉ bộ nhớ thật**, tìm ra ngày
	// 12/11 bằng `cmd/leakcheck` (§13.9): heap server tăng đều 7 MB/phút, 420 MB sau một
	// giờ — đủ để giết tiến trình trên VPS 1 GB.
	//
	// Vì sao: `c[:last]` chỉ hạ `len`, mảng nền vẫn giữ con trỏ ở ô `last`. Bộ dọn rác
	// nhìn cả mảng nền, không nhìn `len`, nên viên vừa bị nhặt KHÔNG BAO GIỜ được thu
	// hồi. Nặng nhất đúng ở trường hợp thường gặp nhất: ô chỉ có một viên (`it.slot ==
	// last`), lúc đó dòng hoán đổi ở trên là gán chính nó, và viên vừa xoá nằm lại nguyên
	// trong mảng nền. Mà bản đồ có hàng nghìn viên EXP môi trường, mỗi viên nhặt xong
	// mọc lại sau 5 giây — tức là gần như mọi lần xoá đều rơi vào trường hợp đó.
	c[last] = nil
	ig.cells[it.cell] = c[:last]
	it.slot = -1
}

// each calls fn for every item in cells overlapping the rectangle [lo, hi].
func (ig *itemGrid) each(lo, hi Vec, fn func(*Item)) {
	a, b := ig.index(lo), ig.index(hi)
	x0, y0, x1, y1 := a%ig.n, a/ig.n, b%ig.n, b/ig.n
	for y := y0; y <= y1; y++ {
		row := y * ig.n
		for x := x0; x <= x1; x++ {
			for _, it := range ig.cells[row+x] {
				fn(it)
			}
		}
	}
}

// ItemsNear calls fn for items in cells overlapping the square of half-size r around c.
func (g *Game) ItemsNear(c Vec, rx, ry float64, fn func(*Item)) {
	g.itemGrid.each(Vec{X: c.X - rx, Y: c.Y - ry}, Vec{X: c.X + rx, Y: c.Y + ry}, fn)
}

// GridAudit là số đo trong ruột lưới vật phẩm — bài kiểm rò rỉ §13.9 đọc.
//
// `Cap` quan trọng hơn `Len`: mảng nền của một ô không bao giờ co lại, nên `Cap - Len` là
// số ô nhớ đang giữ chỗ. Chúng phải LUÔN bằng nil (xem `remove`); con số này chỉ để thấy
// lưới có phình không.
type GridAudit struct {
	Cells, Len, Cap, Stale int
}

// GridAudit duyệt cả lưới. Chỉ dùng trong test và `cmd/leakcheck` — O(số ô).
func (g *Game) GridAudit() GridAudit {
	var a GridAudit
	a.Cells = len(g.itemGrid.cells)
	for _, c := range g.itemGrid.cells {
		a.Len += len(c)
		a.Cap += cap(c)
		// Con trỏ còn sót ngoài `len`: đúng cái rò rỉ tìm ra ngày 12/11.
		full := c[:cap(c)]
		for i := len(c); i < len(full); i++ {
			if full[i] != nil {
				a.Stale++
			}
		}
	}
	return a
}

// SlotAudit: số slot đã cấp, số đang trống, và số ô còn trỏ vào vật phẩm.
func (g *Game) SlotAudit() (total, free, used int) {
	for _, it := range g.itemSlots {
		if it != nil {
			used++
		}
	}
	return len(g.itemSlots), len(g.freeSlots), used
}

// QueueAudit: độ dài các hàng chờ mọc lại. Chúng phải đứng yên quanh một mức.
func (g *Game) QueueAudit() (orbs, items int) {
	return len(g.orbRespawns), len(g.itemRespawns)
}
