package game

import (
	"math"
	"testing"
)

// Thiêu Đốt Tinh Huyết trên đường tu Tiên khí (spec v5.6 §4.1c).
//
// Trước v5.6 chỉ có nhánh EXP, mà người Tiên giới đứng đúng ở đáy cảnh giới nên lên Tiên
// giới là mất hẳn khả năng chạy nhanh. Đó là Q24, và v5.6 trả lời: đốt Tiên khí.

// Người Tiên giới có Tiên khí dư thì đốt được, và đốt đúng vào Tiên khí chứ không vào EXP.
func TestTienGioiBurnsTienKhi(t *testing.T) {
	g := rulesGame(t)
	p := g.AddPlayer(SpawnOptions{Name: "A", Char: g.Cfg.Characters[0]})

	// Lên Chân Tiên rồi nhặt thêm Tiên khí để có phần dư trên đáy cảnh giới.
	g.ForceState(p, g.Prog.TienLevel(0), 10, p.Pos)
	idx := g.TienRealmOf(p)
	if idx != 0 {
		t.Fatalf("muốn đứng ở Chân Tiên (idx 0), đang ở idx %d", idx)
	}
	entry := g.Prog.TienEntry(0)
	p.TienKhi = entry + 50
	expBefore := p.TotalExp

	p.Moving, p.Burning = true, true
	if !g.Burning(p) {
		t.Fatal("Tiên giới còn Tiên khí dư mà không đốt được — đây chính là lỗi Q24")
	}

	dt := 1.0
	g.burnExp(p, dt)
	wantCost := g.Cfg.Burn.TienFracPerSec * g.Cfg.Cultivation.TienKhiPerStep[0] * dt
	if got := entry + 50 - p.TienKhi; math.Abs(got-wantCost) > 1e-9 {
		t.Errorf("đốt %.2f Tiên khí/giây, muốn %.2f", got, wantCost)
	}
	if p.TotalExp != expBefore {
		t.Errorf("EXP đổi từ %.1f sang %.1f — người Tiên giới phải đốt Tiên khí, không đụng EXP",
			expBefore, p.TotalExp)
	}
	if p.BurnedExp <= 0 {
		t.Error("màn Tổng kết sẽ hiện 0 tu vi đã thiêu đốt")
	}
}

// Đốt không bao giờ kéo xuống dưới mốc khởi điểm của cảnh giới đang đứng — tức không rớt
// cảnh giới vì chạy nhanh (spec: rớt cảnh giới chỉ đến từ việc bị chém, §5.1).
func TestTienBurnStopsAtRealmEntry(t *testing.T) {
	g := rulesGame(t)
	p := g.AddPlayer(SpawnOptions{Name: "A", Char: g.Cfg.Characters[0]})
	g.ForceState(p, g.Prog.TienLevel(1), 10, p.Pos) // Thần Tiên
	entry := g.Prog.TienEntry(1)
	p.TienKhi = entry + 1
	levelBefore := p.Level

	p.Moving, p.Burning = true, true
	for i := 0; i < 200; i++ {
		g.burnExp(p, 1.0)
	}
	if p.TienKhi < entry-1e-9 {
		t.Errorf("Tiên khí tụt xuống %.2f, đáy cảnh giới là %.2f", p.TienKhi, entry)
	}
	if p.Level != levelBefore {
		t.Errorf("rớt cảnh giới vì chạy nhanh (level %d → %d) — spec cấm", levelBefore, p.Level)
	}
	if g.Burning(p) {
		t.Error("chạm đáy rồi mà vẫn còn đốt — phải tự tắt")
	}
}

// Đại Thừa tầng 10 đã sang đường Tiên khí nhưng chưa nhặt viên nào thì chưa đốt được;
// nhặt một viên là đốt được ngay (spec v5.6 nói rõ trường hợp này).
func TestTopMortalNeedsFirstTienKhi(t *testing.T) {
	g := rulesGame(t)
	p := g.AddPlayer(SpawnOptions{Name: "A", Char: g.Cfg.Characters[0]})
	g.ForceState(p, g.Prog.TopMortalLevel(), 10, p.Pos)
	if !g.BanksTienKhi(p) {
		t.Fatal("Đại Thừa tầng 10 phải đã sang đường tu Tiên khí")
	}
	p.TienKhi = 0
	p.Moving, p.Burning = true, true
	if g.Burning(p) {
		t.Error("chưa có Tiên khí nào mà đã đốt được")
	}
	g.GainTienKhi(p, 20)
	if !g.Burning(p) {
		t.Error("nhặt Tiên khí rồi mà vẫn không đốt được")
	}
}

// Người phàm vẫn đốt EXP như cũ — v5.6 chỉ thêm nhánh Tiên khí, không đổi nhánh cũ.
func TestMortalStillBurnsExp(t *testing.T) {
	g := rulesGame(t)
	p := g.AddPlayer(SpawnOptions{Name: "A", Char: g.Cfg.Characters[0]})
	g.GainExp(p, g.Prog.ExpNeeded(p.Level)*0.9, false)
	before := p.TotalExp

	p.Moving, p.Burning = true, true
	if !g.Burning(p) {
		t.Fatal("người phàm còn EXP tầng mà không đốt được")
	}
	g.burnExp(p, 1.0)
	if p.TotalExp >= before {
		t.Error("đốt mà EXP không giảm")
	}
}

// Đạo Tổ: Tiên khí bị `overflowToScore` ghim đúng ở mốc khởi điểm (1 500) nên theo luật
// "không xuống dưới mốc khởi điểm" thì không còn gì để đốt.
//
// Test này khoá lại HIỆN TRẠNG chứ không khẳng định đó là ý đồ — xem Q26. Nếu anh chốt cho
// Đạo Tổ đốt được thì sửa test cùng với luật.
func TestDaoToHasNoTienKhiHeadroom(t *testing.T) {
	g := rulesGame(t)
	p := g.AddPlayer(SpawnOptions{Name: "A", Char: g.Cfg.Characters[0]})
	top := g.Prog.TienLevel(g.Prog.TienCount() - 1)
	g.ForceState(p, top, 10, p.Pos)
	g.GainTienKhi(p, 5000) // dư bao nhiêu cũng đổi hết sang Uy Danh

	if p.TienKhi > g.tienFloor(p) {
		t.Fatalf("Đạo Tổ giữ được %.1f Tiên khí trên đáy %.1f — luật đã đổi, cập nhật Q26",
			p.TienKhi, g.tienFloor(p))
	}
	p.Moving, p.Burning = true, true
	if g.Burning(p) {
		t.Error("Đạo Tổ đốt được trong khi Tiên khí đang ghim ở đáy — mâu thuẫn")
	}
}
