package game

import (
	"bladeascend/server/internal/config"
	"math"
	"testing"
)

// Tiên giới theo spec v6.4 (§3.4, §4.1c, §5.1) — Tiên khí **×20** so với v5.9.
//
// Con số trong các test này lấy thẳng từ checklist §17 của spec, không tự đặt — có thế thì
// test mới là bản dịch của spec chứ không phải bản chụp của code.
//
// Trần: 1 000 / 3 000 / 6 000 / 10 000 / 15 000. Đáy: 1 / 1 001 / 3 001 / 6 001 / 10 001.
// Thiêu Đốt vẫn **8 · 16 · 24 · 32 · 40** mỗi giây — `tien_frac_per_sec` đi ngược lại đúng
// bằng mức trần đổi (v6.4 ×20 → 0,4%; v6.6 ×10 → 0,8%). Đó là chỗ dễ sửa sai nhất của mỗi
// lần đổi thang: đổi trần mà quên đổi tỉ lệ thì một giây chạy nhanh tốn cả trăm viên.

// tienPlayer dựng một người đã vào Tiên giới với đúng ngần này Tiên khí.
func tienPlayer(t *testing.T, g *Game, khi float64) *Player {
	t.Helper()
	p := g.AddPlayer(SpawnOptions{Name: "T", Char: g.Cfg.Characters[0]})
	p.InvulnUntil = 0
	// Đầy thanh Đại Thừa tầng 10 mới nhặt được Tiên khí (spec v5.7).
	g.ForceState(p, g.Prog.TopMortalLevel(), 10, p.Pos)
	p.TotalExp = g.Prog.MaxExp()
	if !g.BanksTienKhi(p) {
		t.Fatal("đầy thanh Đại Thừa 10 mà vẫn chưa sang đường tu Tiên khí")
	}
	g.GainTienKhi(p, khi)
	return p
}

// Đại Thừa tầng 10 CHƯA đầy thanh thì vẫn tu Linh khí, không nhặt được Tiên khí.
func TestTienEntryNeedsFullBar(t *testing.T) {
	g := rulesGame(t)
	p := g.AddPlayer(SpawnOptions{Name: "A", Char: g.Cfg.Characters[0]})
	g.ForceState(p, g.Prog.TopMortalLevel(), 10, p.Pos)
	p.TotalExp = g.Prog.MaxExp() - 1

	if g.BanksTienKhi(p) {
		t.Error("chưa đầy thanh mà đã sang đường Tiên khí")
	}
	if !g.CanAbsorbExp(p) {
		t.Error("chưa đầy thanh thì phải còn nhặt được Linh khí")
	}
	if g.CanAbsorbTienKhi(p) {
		t.Error("chưa đầy thanh mà đã nhặt được Tiên khí")
	}

	g.GainExp(p, 1, false)
	if !g.BanksTienKhi(p) {
		t.Error("đầy thanh rồi mà vẫn chưa sang đường Tiên khí")
	}
}

// Cảnh giới Tiên suy ra thẳng từ tổng Tiên khí; viên đầu tiên là vào Chân Tiên ngay.
func TestTienRealmFromTotalKhi(t *testing.T) {
	g := rulesGame(t)
	for _, c := range []struct {
		khi  float64
		want int
		name string
	}{
		{0, -1, "chưa có Tiên khí"},
		{1, 0, "1 → Chân Tiên"},
		{1000, 0, "1 000 → vẫn Chân Tiên"},
		{1001, 1, "1 001 → Kim Tiên"},
		{3000, 1, "3 000 → vẫn Kim Tiên"},
		{3001, 2, "3 001 → Thái Ất"},
		{6000, 2, "6 000 → vẫn Thái Ất"},
		{6001, 3, "6 001 → Đại La"},
		{10000, 3, "10 000 → vẫn Đại La"},
		{10001, 4, "10 001 → Đạo Tổ"},
		{15000, 4, "15 000 → vẫn Đạo Tổ"},
	} {
		if got := g.Prog.TienRealmForKhi(c.khi); got != c.want {
			t.Errorf("%s: ra %d, muốn %d", c.name, got, c.want)
		}
	}
}

// Nạp Tiên khí thì cảnh giới (và `Level`) đi theo.
func TestTienPromotesByKhi(t *testing.T) {
	g := rulesGame(t)
	p := tienPlayer(t, g, 1)
	if got := g.TienRealmOf(p); got != 0 {
		t.Fatalf("1 Tiên khí phải là Chân Tiên, ra idx %d", got)
	}
	if want := g.Prog.TienLevel(0); p.Level != want {
		t.Errorf("level %d, muốn %d — `Level` phải đi theo Tiên khí", p.Level, want)
	}
	g.GainTienKhi(p, 1000) // 1 001 → Kim Tiên
	if got := g.TienRealmOf(p); got != 1 {
		t.Errorf("1 001 Tiên khí phải là Kim Tiên, ra idx %d", got)
	}
	if want := g.Prog.TienLevel(1); p.Level != want {
		t.Errorf("level %d, muốn %d", p.Level, want)
	}
}

// Đạo Tổ: nạp thanh bình thường tới 15 000; CHỈ khi đang đúng 15 000 thì dư mới đổi Uy Danh.
func TestDaoToFillsThenConverts(t *testing.T) {
	g := rulesGame(t)
	p := tienPlayer(t, g, 10001) // vừa vào Đạo Tổ
	if got := g.TienRealmOf(p); got != 4 {
		t.Fatalf("10 001 Tiên khí phải là Đạo Tổ, ra idx %d", got)
	}
	scoreBefore := p.Score

	g.GainTienKhi(p, 100) // 10 101, chưa đầy
	if p.TienKhi != 10101 {
		t.Errorf("Tiên khí %.0f, muốn 10 101 — Đạo Tổ phải nạp được thanh", p.TienKhi)
	}
	if p.Score != scoreBefore {
		t.Error("chưa đầy 15 000 mà đã cộng Uy Danh")
	}

	g.GainTienKhi(p, 4899) // đúng 15 000
	if p.TienKhi != 15000 || p.Score != scoreBefore {
		t.Errorf("Tiên khí %.0f / Uy Danh +%.1f — đúng 15 000 thì chưa được cộng gì",
			p.TienKhi, p.Score-scoreBefore)
	}

	g.GainTienKhi(p, 500) // dư 500 → +50 Uy Danh (10:1)
	if p.TienKhi != 15000 {
		t.Errorf("Tiên khí %.0f — thanh không được vượt 15 000", p.TienKhi)
	}
	if got := p.Score - scoreBefore; math.Abs(got-50) > 1e-6 {
		t.Errorf("dư 500 Tiên khí được %.1f Uy Danh, muốn 50 (10:1)", got)
	}
}

// Đạo Tổ Thiêu Đốt được, và đốt xong phải nạp lại đầy mới đổi Uy Danh tiếp (Q26 đã xong).
func TestDaoToCanBurn(t *testing.T) {
	g := rulesGame(t)
	p := tienPlayer(t, g, 15000)
	p.Moving, p.Burning = true, true
	if !g.Burning(p) {
		t.Fatal("Đạo Tổ đầy thanh mà không Thiêu Đốt được — đây chính là Q26")
	}
	g.burnExp(p, 1.0)
	if want := 15000 - 40.0; math.Abs(p.TienKhi-want) > 1e-6 {
		t.Errorf("đốt 1 giây còn %.1f Tiên khí, muốn %.0f (0,8%% × 5 000 = 40/giây)", p.TienKhi, want)
	}
	if g.TienRealmOf(p) != 4 {
		t.Error("đốt một giây mà rớt khỏi Đạo Tổ")
	}
}

// Tốc độ đốt = 0,8% ĐỘ RỘNG KHOẢNG của cảnh giới: 8 · 16 · 24 · 32 · 40 (spec v6.6 §4.1c).
//
// Lượng TUYỆT ĐỐI giữ nguyên qua CẢ HAI lần đổi thang (v6.4 ×20, v6.6 ×10) — đó là cả ý
// của việc tỉ lệ đi ngược lại đúng bằng mức trần đổi.
func TestTienBurnRatePerRealm(t *testing.T) {
	g := rulesGame(t)
	for _, c := range []struct {
		khi  float64
		want float64
		name string
	}{
		{500, 8, "Chân Tiên"},
		{2000, 16, "Kim Tiên"},
		{4000, 24, "Thái Ất"},
		{8000, 32, "Đại La"},
		{12000, 40, "Đạo Tổ"},
	} {
		p := tienPlayer(t, g, c.khi)
		if got := g.tienBurnPerSec(p); math.Abs(got-c.want) > 1e-6 {
			t.Errorf("%s đốt %.1f/giây, muốn %.0f", c.name, got, c.want)
		}
		g.RemovePlayer(p.ID)
	}
}

// Đốt không bao giờ xuống dưới đáy cảnh giới đang đứng.
func TestTienBurnStopsAtRealmFloor(t *testing.T) {
	g := rulesGame(t)
	p := tienPlayer(t, g, 3500) // Thái Ất, đáy 3 001
	p.Moving, p.Burning = true, true
	// 499 điểm trên đáy, đốt 24/giây → ~21 giây là chạm đáy; 60 vòng là dư để thử vượt qua.
	for i := 0; i < 60; i++ {
		g.burnExp(p, 1.0)
	}
	if p.TienKhi < 3001-1e-9 {
		t.Errorf("Tiên khí tụt xuống %.1f, đáy Thái Ất là 3 001", p.TienKhi)
	}
	if g.TienRealmOf(p) != 2 {
		t.Error("Thiêu Đốt làm rớt cảnh giới — spec cấm")
	}
	if g.Burning(p) {
		t.Error("chạm đáy rồi mà vẫn còn đốt")
	}
}

// countTienOrbs đếm SỐ VIÊN (khác `totalTienOrbs` đếm tổng ĐIỂM).
func countTienOrbs(g *Game) int {
	n := 0
	for _, it := range g.Items {
		if it.Kind == ItemTienKhi {
			n++
		}
	}
	return n
}

func totalTienOrbs(g *Game) int {
	total := 0.0
	for _, it := range g.Items {
		if it.Kind == ItemTienKhi {
			total += it.Value
		}
	}
	return int(total)
}

// Người phàm vẫn đốt EXP như cũ.
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

// splitTienOrbs: mệnh giá 500 / 100 / 10 / 2 / 1, lớn trước, tối đa 40 viên, và **tổng điểm
// phải đúng bằng số vào** (spec v6.4 §3.4).
//
// Tổng là thứ quan trọng nhất ở đây: trần 40 viên được thực hiện bằng cách GỘP phần nhỏ,
// nên một cách cài sai rất dễ xảy ra là cắt bớt viên — người chơi mất điểm mà không ai thấy,
// vì trên màn hình chỉ là "ít viên hơn một chút".
func TestSplitTienOrbs(t *testing.T) {
	for _, total := range []float64{0, 1, 2, 3, 9, 10, 99, 100, 480, 2099, 5399, 9000, 123456} {
		orbs := splitTienOrbs(total, testOrbUnits, tienOrbMax)
		sum := 0.0
		for _, v := range orbs {
			sum += v
			if v != 1 && v != 2 && v != 10 && v != 100 && v != 500 {
				// Viên gộp là ngoại lệ DUY NHẤT được phép lệch mệnh giá, và chỉ khi đã chạm trần.
				if len(orbs) != tienOrbMax || v != orbs[len(orbs)-1] {
					t.Errorf("total %.0f: có viên mệnh giá lạ %.0f", total, v)
				}
			}
		}
		if sum != math.Floor(total) {
			t.Errorf("total %.0f: tổng viên %.0f, phải đúng bằng %.0f", total, sum, math.Floor(total))
		}
		if len(orbs) > tienOrbMax {
			t.Errorf("total %.0f: %d viên, trần là %d", total, len(orbs), tienOrbMax)
		}
	}
}

// Ví dụ của checklist §17 chia ra đúng những viên nào.
func TestSplitTienOrbsChecklistExample(t *testing.T) {
	// 2 099 = 500×4 + 10×9 + 2×4 + 1×1 → 18 viên.
	got := splitTienOrbs(2099, testOrbUnits, tienOrbMax)
	want := map[float64]int{500: 4, 10: 9, 2: 4, 1: 1}
	have := map[float64]int{}
	for _, v := range got {
		have[v]++
	}
	if len(got) != 18 {
		t.Errorf("2 099 ra %d viên, muốn 18", len(got))
	}
	for unit, n := range want {
		if have[unit] != n {
			t.Errorf("2 099: có %d viên mệnh giá %.0f, muốn %d", have[unit], unit, n)
		}
	}
}

// Trần 40 viên phải THẬT SỰ chặn được, kể cả khi mệnh giá lớn nhất cũng không đủ gộp.
//
// Với bộ mệnh giá hiện tại thì phép chia tham lam không bao giờ vượt 40 (tối đa 27 viên ở
// mức Tiên khí cao nhất game cho phép). Test này gọi thẳng hàm với một con số vượt tầm để
// khoá luật lại cho lần sau ai đó nâng trần Tiên khí lên nữa.
func TestSplitTienOrbsCapsAt40(t *testing.T) {
	const huge = 500 * 60 // 60 viên 500 nếu không có trần
	orbs := splitTienOrbs(huge, testOrbUnits, tienOrbMax)
	if len(orbs) != tienOrbMax {
		t.Fatalf("%d ra %d viên, muốn đúng %d", huge, len(orbs), tienOrbMax)
	}
	sum := 0.0
	for _, v := range orbs {
		sum += v
	}
	if sum != huge {
		t.Errorf("gộp xong tổng %.0f, phải giữ nguyên %d", sum, huge)
	}
	if last := orbs[len(orbs)-1]; last <= 500 {
		t.Errorf("viên gộp cuối %.0f — phải to hơn mệnh giá thường vì nó ôm phần dư", last)
	}
}

// Bộ mệnh giá spec v6.4 §3.4 — bài test bảng ở trên khoá hàm thuần với đúng bộ này;
// TestTienDropUnitsInConfig khoá rằng game.json đang dùng đúng bộ đó.
var testOrbUnits = []float64{500, 100, 10, 2, 1}

const tienOrbMax = 40

func TestTienDropUnitsInConfig(t *testing.T) {
	c := config.MustLoadDefault().Cultivation
	if c.TienDropMaxOrbs != tienOrbMax || len(c.TienDropUnits) != len(testOrbUnits) {
		t.Fatalf("game.json: tien_drop_units=%v tien_drop_max_orbs=%d, spec v6.4 §3.4 là %v / %d",
			c.TienDropUnits, c.TienDropMaxOrbs, testOrbUnits, tienOrbMax)
	}
	for i, u := range testOrbUnits {
		if c.TienDropUnits[i] != u {
			t.Fatalf("tien_drop_units[%d] = %v, muốn %v", i, c.TienDropUnits[i], u)
		}
	}
}
