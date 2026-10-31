package game

import (
	"math"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// Spec v6.5 §6.3: mọi phạm vi kỹ năng là `hệ số × U`, `U = pickup_r × scale người dùng`.

// Hệ số lấy THẲNG từ bảng spec — đây là bản dịch của spec, không phải bản chụp của config.
var specRangesU = map[string]float64{
	"loa_cone":       6,
	"loa_push":       6,
	"ru_ngu_radius":  5,
	"dan_em_follow":  3,
	"dan_em_seek":    8,
	"tran_chau_cone": 8,
	"dash":           7,
	"slick_width":    1,
	"luoi_mi_range":  4.5,
	"luoi_mi_radius": 3,
	"shell_wave":     2,
	"shell_push":     3,
}

func TestSkillRangesMatchSpec(t *testing.T) {
	g := skillGame(t)
	sc := g.Cfg.Skills
	if sc.RangeUnit != "pickup_r" {
		t.Errorf("range_unit = %q, spec nói U = pickup_r × scale", sc.RangeUnit)
	}
	if sc.PushSecs != 0.3 {
		t.Errorf("push_secs = %g, spec nói 0.3", sc.PushSecs)
	}
	for key, want := range specRangesU {
		if got := sc.RangesU[key]; got != want {
			t.Errorf("ranges_u[%q] = %g, spec nói %g", key, got, want)
		}
	}
	for key := range sc.RangesU {
		if _, ok := specRangesU[key]; !ok {
			t.Errorf("ranges_u có khoá lạ %q — không có trong bảng spec §6.3", key)
		}
	}
}

// U phải đúng bằng bán kính vòng hút của chính người đó, và bảng tham khảo của spec phải ra.
func TestSkillUMatchesPickupRadius(t *testing.T) {
	g := skillGame(t)
	// Level 0 = Luyện Khí tầng 1, scale 1.0 — đúng dòng đầu bảng tham khảo của spec.
	p := put(g, "ba_tam_loa", Vec{X: 3000, Y: 3000}, 0, 1)
	if got, want := g.SkillU(p), g.Cfg.Player.PickupR*p.Scale; got != want {
		t.Fatalf("U = %g, muốn pickup_r × scale = %g", got, want)
	}
	// Spec §6.3 bảng tham khảo: Luyện Khí 1 → U = 90.
	if got := g.SkillU(p); math.Abs(got-90) > 1e-9 {
		t.Errorf("Luyện Khí 1: U = %g, bảng spec ghi 90", got)
	}
	// Và Loa Phường ở cảnh giới đó với tới 540 (6 U).
	if got := g.skillRange(p, "loa_cone"); math.Abs(got-540) > 1e-9 {
		t.Errorf("Loa Phường Luyện Khí 1 với %g, bảng spec ghi 540", got)
	}
}

// Phạm vi phải LỚN LÊN theo cảnh giới người dùng — đó là cả điểm của việc đổi sang U.
func TestSkillRangeGrowsWithRealm(t *testing.T) {
	g := skillGame(t)
	low := put(g, "ba_tam_loa", Vec{X: 2000, Y: 2000}, 0, 1)
	high := put(g, "ba_tam_loa", Vec{X: 8000, Y: 8000}, 0, 1)
	g.ForceState(high, g.Prog.TopMortalLevel(), 10, high.Pos)

	lo, hi := g.skillRange(low, "loa_cone"), g.skillRange(high, "loa_cone")
	if hi <= lo {
		t.Fatalf("Đại Thừa với %g, Luyện Khí với %g — phạm vi phải tăng theo cảnh giới", hi, lo)
	}
	if ratio := hi / lo; math.Abs(ratio-high.Scale/low.Scale) > 1e-9 {
		t.Errorf("tỉ lệ phạm vi %g, phải đúng bằng tỉ lệ scale %g", ratio, high.Scale/low.Scale)
	}
}

// Khoảng đẩy tính theo U của NGƯỜI DÙNG, không theo người bị đẩy (spec v6.5 mục 2).
//
// Đây là luật dễ cài sai nhất của v6.5: lấy `o.Scale` thay vì `p.Scale` thì vẫn chạy, vẫn
// đẩy, chỉ là người to bị đẩy xa hơn — nhìn trong game không ai phân biệt được.
func TestPushDistanceUsesCasterU(t *testing.T) {
	g := skillGame(t)
	caster := put(g, "ba_tam_loa", Vec{X: 3000, Y: 3000}, 10, 1)
	caster.Facing = 0
	want := g.skillRange(caster, "loa_push")

	for _, c := range []struct {
		name  string
		level int
	}{{"nạn nhân nhỏ", 0}, {"nạn nhân TO hơn người đẩy", -1}} {
		victim := put(g, "cu_rua", Vec{X: 3000 + 200, Y: 3000}, 10, 1)
		if c.level < 0 {
			g.ForceState(victim, g.Prog.TopMortalLevel(), 10, victim.Pos)
		}
		caster.SkillCharges = 1
		caster.SkillReadyAt = 0
		g.CastSkill(caster)

		if victim.KnockUntil <= g.Time {
			t.Fatalf("%s: không bị đẩy", c.name)
		}
		// Vận tốc đẩy × thời gian đẩy = quãng đường.
		if got := victim.KnockVel.Len() * g.Cfg.Skills.PushSecs; math.Abs(got-want) > 1e-6 {
			t.Errorf("%s: bị đẩy %g, phải %g (U của NGƯỜI ĐẨY)", c.name, got, want)
		}
		g.RemovePlayer(victim.ID)
	}
}

// Bí Kíp: 2 lượt ở phàm giới, 3 ở Tiên giới; rớt về Đại Thừa đang cầm 3 thì còn 2.
func TestSkillChargesCapByRealm(t *testing.T) {
	g := skillGame(t)
	p := put(g, "cu_rua", Vec{X: 3000, Y: 3000}, 10, 1)

	if got := g.MaxSkillCharges(p); got != 2 {
		t.Errorf("người phàm giữ được %d lượt, spec nói 2", got)
	}
	for i := 0; i < 5; i++ {
		g.AddSkillCharge(p)
	}
	if p.SkillCharges != 2 {
		t.Errorf("nhặt 5 Bí Kíp ở phàm giới còn %d lượt, trần là 2", p.SkillCharges)
	}

	// Lên Tiên giới: trần thành 3, nhặt thêm được đúng 1.
	g.ForceState(p, g.Prog.TopMortalLevel(), 10, p.Pos)
	p.TotalExp = g.Prog.MaxExp()
	g.GainTienKhi(p, 1)
	if got := g.MaxSkillCharges(p); got != 3 {
		t.Fatalf("Tiên giới giữ được %d lượt, spec nói 3", got)
	}
	if !g.AddSkillCharge(p) || p.SkillCharges != 3 {
		t.Errorf("Tiên giới nhặt thêm: còn %d lượt, muốn 3", p.SkillCharges)
	}
	if g.AddSkillCharge(p) {
		t.Error("đã đủ 3 mà vẫn nhận thêm lượt")
	}

	// Rớt khỏi Tiên giới → cắt còn 2.
	att := put(g, "be_ti", Vec{X: 3100, Y: 3000}, 10, 1)
	g.robTienKhi(att, p)
	if g.TienRealmOf(p) >= 0 {
		t.Fatal("chém Chân Tiên mà vẫn còn ở Tiên giới")
	}
	if p.SkillCharges != 2 {
		t.Errorf("rớt về Đại Thừa còn %d lượt, spec nói cắt còn 2", p.SkillCharges)
	}
}

// Trong mã kỹ năng KHÔNG được còn một hằng số phạm vi nào (CHANGES_v6_5 mục 5).
//
// Kiểm bằng cách đọc chính tệp nguồn: phạm vi phải đi qua `skillRange` / `SkillU`. Một con
// số 500 sót lại trong `castLoaPhuong` sẽ không làm test nào khác đỏ — nó chỉ làm kỹ năng
// của Đạo Tổ ngắn bằng của Luyện Khí, và chuyện đó phải chơi hàng giờ mới nhận ra.
func TestNoHardcodedSkillRanges(t *testing.T) {
	for _, name := range []string{"skills.go", "skillstate.go", "skillrange.go"} {
		src, err := os.ReadFile(filepath.Join(".", name))
		if err != nil {
			t.Fatalf("đọc %s: %v", name, err)
		}
		for i, line := range strings.Split(string(src), "\n") {
			code := line
			if k := strings.Index(code, "//"); k >= 0 {
				code = code[:k] // bỏ phần chú thích: ở đó nói "6 U", "dài 500" là chuyện bình thường
			}
			if strings.TrimSpace(code) == "" {
				continue
			}
			// Đổi độ ↔ radian (`*180/math.Pi`) không phải phạm vi — bỏ qua dòng có `math.Pi`.
			if strings.Contains(code, "math.Pi") {
				continue
			}
			// Số có từ 3 chữ số trở lên trong mã kỹ năng gần như chắc chắn là một phạm vi.
			// Ngưỡng 3 chữ số để không bắt nhầm `0.3`, `1e-6`, chỉ số mảng hay `2` của `×2`.
			if m := bigNumber.FindString(code); m != "" {
				t.Errorf("%s:%d còn hằng số %q — phạm vi phải lấy qua skillRange()\n  %s",
					name, i+1, m, strings.TrimSpace(line))
			}
		}
	}
}

// bigNumber bắt số nguyên ≥ 3 chữ số không nằm trong tên định danh (`Vec2`, `StBrace1`).
var bigNumber = regexp.MustCompile(`(^|[^\w.])\d{3,}(\b|[^\w])`)
