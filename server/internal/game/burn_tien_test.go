package game

import (
	"math"
	"testing"
)

// Tiên giới theo spec v6.4 (§3.4, §4.1c, §5.1) — Tiên khí **×20** so với v5.9.
//
// Con số trong các test này lấy thẳng từ checklist §17 của spec, không tự đặt — có thế thì
// test mới là bản dịch của spec chứ không phải bản chụp của code.
//
// Trần: 2 000 / 6 000 / 12 000 / 20 000 / 30 000. Đáy: 1 / 2 001 / 6 001 / 12 001 / 20 001.
// Thiêu Đốt vẫn **8 · 16 · 24 · 32 · 40** mỗi giây như trước ×20 — vì `tien_frac_per_sec`
// hạ từ 8% xuống 0,4% đúng theo §4.1c. Đó là chỗ dễ sửa sai nhất của lượt ×20: đổi trần mà
// quên đổi tỉ lệ thì một giây chạy nhanh tốn cả trăm viên Tiên khí.

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
		{2000, 0, "2 000 → vẫn Chân Tiên"},
		{2001, 1, "2 001 → Kim Tiên"},
		{6000, 1, "6 000 → vẫn Kim Tiên"},
		{6001, 2, "6 001 → Thái Ất"},
		{12000, 2, "12 000 → vẫn Thái Ất"},
		{12001, 3, "12 001 → Đại La"},
		{20000, 3, "20 000 → vẫn Đại La"},
		{20001, 4, "20 001 → Đạo Tổ"},
		{30000, 4, "30 000 → vẫn Đạo Tổ"},
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
	g.GainTienKhi(p, 2000) // 2 001 → Kim Tiên
	if got := g.TienRealmOf(p); got != 1 {
		t.Errorf("2 001 Tiên khí phải là Kim Tiên, ra idx %d", got)
	}
	if want := g.Prog.TienLevel(1); p.Level != want {
		t.Errorf("level %d, muốn %d", p.Level, want)
	}
}

// Đạo Tổ: nạp thanh bình thường tới 1 500; CHỈ khi đang đúng 1 500 thì dư mới đổi Uy Danh.
func TestDaoToFillsThenConverts(t *testing.T) {
	g := rulesGame(t)
	p := tienPlayer(t, g, 20001) // vừa vào Đạo Tổ
	if got := g.TienRealmOf(p); got != 4 {
		t.Fatalf("20 001 Tiên khí phải là Đạo Tổ, ra idx %d", got)
	}
	scoreBefore := p.Score

	g.GainTienKhi(p, 100) // 20 101, chưa đầy
	if p.TienKhi != 20101 {
		t.Errorf("Tiên khí %.0f, muốn 20 101 — Đạo Tổ phải nạp được thanh", p.TienKhi)
	}
	if p.Score != scoreBefore {
		t.Error("chưa đầy 30 000 mà đã cộng Uy Danh")
	}

	g.GainTienKhi(p, 9899) // đúng 30 000
	if p.TienKhi != 30000 || p.Score != scoreBefore {
		t.Errorf("Tiên khí %.0f / Uy Danh +%.1f — đúng 30 000 thì chưa được cộng gì",
			p.TienKhi, p.Score-scoreBefore)
	}

	g.GainTienKhi(p, 500) // dư 500 → +50 Uy Danh (10:1)
	if p.TienKhi != 30000 {
		t.Errorf("Tiên khí %.0f — thanh không được vượt 30 000", p.TienKhi)
	}
	if got := p.Score - scoreBefore; math.Abs(got-50) > 1e-6 {
		t.Errorf("dư 500 Tiên khí được %.1f Uy Danh, muốn 50 (10:1)", got)
	}
}

// Đạo Tổ Thiêu Đốt được, và đốt xong phải nạp lại đầy mới đổi Uy Danh tiếp (Q26 đã xong).
func TestDaoToCanBurn(t *testing.T) {
	g := rulesGame(t)
	p := tienPlayer(t, g, 30000)
	p.Moving, p.Burning = true, true
	if !g.Burning(p) {
		t.Fatal("Đạo Tổ đầy thanh mà không Thiêu Đốt được — đây chính là Q26")
	}
	g.burnExp(p, 1.0)
	if want := 30000 - 40.0; math.Abs(p.TienKhi-want) > 1e-6 {
		t.Errorf("đốt 1 giây còn %.1f Tiên khí, muốn %.0f (0,4%% × 10 000 = 40/giây)", p.TienKhi, want)
	}
	if g.TienRealmOf(p) != 4 {
		t.Error("đốt một giây mà rớt khỏi Đạo Tổ")
	}
}

// Tốc độ đốt = 0,4% ĐỘ RỘNG KHOẢNG của cảnh giới: 8 · 16 · 24 · 32 · 40 (spec v6.4 §4.1c).
//
// Lượng TUYỆT ĐỐI giữ nguyên sau khi Tiên khí ×20 — đó là cả ý của việc hạ 8% xuống 0,4%.
func TestTienBurnRatePerRealm(t *testing.T) {
	g := rulesGame(t)
	for _, c := range []struct {
		khi  float64
		want float64
		name string
	}{
		{1000, 8, "Chân Tiên"},
		{4000, 16, "Kim Tiên"},
		{8000, 24, "Thái Ất"},
		{16000, 32, "Đại La"},
		{25000, 40, "Đạo Tổ"},
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
	p := tienPlayer(t, g, 7000) // Thái Ất, đáy 6 001
	p.Moving, p.Burning = true, true
	// 999 điểm trên đáy, đốt 24/giây → ~42 giây là chạm đáy; 60 vòng là dư để thử vượt qua.
	for i := 0; i < 60; i++ {
		g.burnExp(p, 1.0)
	}
	if p.TienKhi < 6001-1e-9 {
		t.Errorf("Tiên khí tụt xuống %.1f, đáy Thái Ất là 6 001", p.TienKhi)
	}
	if g.TienRealmOf(p) != 2 {
		t.Error("Thiêu Đốt làm rớt cảnh giới — spec cấm")
	}
	if g.Burning(p) {
		t.Error("chạm đáy rồi mà vẫn còn đốt")
	}
}

// Checklist §17 (v6.4): Thái Ất đang **9 000** bị chém → Kim Tiên, Tiên khí **2 001**;
// người chém (Tiên giới) **+3 499**; rơi **2 099** Tiên khí thành ≤ 40 viên.
func TestCutTienGioiThaiAt9000(t *testing.T) {
	g := rulesGame(t)
	victim := tienPlayer(t, g, 9000)
	att := tienPlayer(t, g, 1)
	attBefore := att.TienKhi

	g.robTienKhi(att, victim)

	if victim.TienKhi != 2001 {
		t.Errorf("nạn nhân còn %.0f Tiên khí, muốn 2 001 (đáy Kim Tiên)", victim.TienKhi)
	}
	if got := g.TienRealmOf(victim); got != 1 {
		t.Errorf("nạn nhân ở idx %d, muốn 1 (Kim Tiên)", got)
	}
	// loss = 9 000 − 2 001 = 6 999; cướp 50% = 3 499.5
	if got := att.TienKhi - attBefore; math.Abs(got-3499.5) > 0.6 {
		t.Errorf("người chém nhận %.1f Tiên khí, muốn ≈3 499", got)
	}
	if got := totalTienOrbs(g); got != 2099 {
		t.Errorf("rơi %d điểm Tiên khí, muốn 2 099 (30%% của 6 999)", got)
	}
	// Trần 40 viên là phần MỚI của v6.4: với bộ mệnh giá cũ {10,2,1} thì 2 099 điểm là 210
	// viên rải quanh một cái xác.
	if n := countTienOrbs(g); n > 40 {
		t.Errorf("rơi %d viên, spec cho tối đa 40", n)
	}
}

// Checklist §17: Chân Tiên bị chém → sàn Đại Thừa (tầng 1, 0%), Tiên khí 0.
func TestCutChanTienFallsToDaiThuaFloor(t *testing.T) {
	g := rulesGame(t)
	victim := tienPlayer(t, g, 1600)
	att := g.AddPlayer(SpawnOptions{Name: "A", Char: g.Cfg.Characters[1]})

	g.robTienKhi(att, victim)

	if victim.TienKhi != 0 {
		t.Errorf("Tiên khí còn %.0f, muốn 0", victim.TienKhi)
	}
	wantLevel := (len(g.Cfg.Cultivation.Realms) - 1) * g.Prog.Tiers()
	if victim.Level != wantLevel {
		t.Errorf("level %d, muốn %d (Đại Thừa tầng 1)", victim.Level, wantLevel)
	}
	if want := g.Prog.Threshold(wantLevel); math.Abs(victim.TotalExp-want) > 1e-6 {
		t.Errorf("total_exp %.0f, muốn %.0f (sàn Đại Thừa)", victim.TotalExp, want)
	}
	// Checklist §17 (v6.4): Chân Tiên 1 600 Tiên khí bị chém → rơi 480.
	if got := totalTienOrbs(g); got != 480 {
		t.Errorf("rơi %d điểm Tiên khí, muốn 480 (30%% của 1 600)", got)
	}
}

// v5.9: người chém là PHÀM (không nhặt được Tiên khí) thì vẫn phải rơi đủ 30%.
func TestMortalAttackerStillDropsTienKhi(t *testing.T) {
	g := rulesGame(t)
	victim := tienPlayer(t, g, 1600)
	att := g.AddPlayer(SpawnOptions{Name: "A", Char: g.Cfg.Characters[1]}) // Luyện Khí

	g.robTienKhi(att, victim)

	if att.TienKhi != 0 {
		t.Errorf("người phàm nhận %.0f Tiên khí — spec nói không nhận được", att.TienKhi)
	}
	if got := totalTienOrbs(g); got != 480 {
		t.Errorf("rơi %d điểm Tiên khí, muốn 480 — phần rơi ra đất không phụ thuộc người chém", got)
	}
}

// Viên rơi ra phải văng ĐỦ XA khỏi xác: người Tiên giới to gấp 5–6 lần và vật phẩm vẽ dưới
// nhân vật, nên viên rơi sát chân là nằm khuất hẳn — đó là lỗi "chém mà không thấy rơi".
func TestTienOrbsScatterFarEnough(t *testing.T) {
	g := rulesGame(t)
	victim := tienPlayer(t, g, 9000)
	victim.Pos = Vec{X: 5000, Y: 5000}
	att := tienPlayer(t, g, 1)

	g.robTienKhi(att, victim)

	minD := g.Cfg.Cultivation.TienDropScatterMin
	n := 0
	for _, it := range g.Items {
		if it.Kind != ItemTienKhi {
			continue
		}
		n++
		if d := it.Pos.Dist(victim.Pos); d < minD-1e-6 {
			t.Errorf("viên rơi cách xác %.0f, tối thiểu phải %.0f", d, minD)
		}
	}
	if n == 0 {
		t.Fatal("không có viên nào rơi")
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
		orbs := splitTienOrbs(total)
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
	got := splitTienOrbs(2099)
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
	orbs := splitTienOrbs(huge)
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
