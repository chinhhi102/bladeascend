package game

import (
	"math"
	"testing"

	"kiemxoay/server/internal/config"
)

// Checklist §17: "Cô Ba Kim Đan bày sạp (bậc 3): người Kim Đan và thấp hơn không chém được,
// phải đi vòng; người Nguyên Anh chém được; Cô Ba đi xuyên sạp của mình; dẫm vỏ chuối →
// trượt + choáng 1s; sau 6s sạp biến mất; Shipper đang Gom Đơn đi xuyên qua."
//
// Một hàm cho mỗi câu, vì mỗi câu là một luật người chơi cảm nhận được riêng.

// coBaAt dựng một Cô Ba ở cấp `level`, bày sạp hướng +X, và trả về dãy sạp vừa bày.
func coBaAt(t *testing.T, g *Game, level int) (*Player, []tempStall) {
	t.Helper()
	p := put(g, "co_ba_ganh", Vec{X: 3000, Y: 3000}, level, 20)
	p.Facing = 0
	g.AddSkillCharge(p)
	if !g.CastSkill(p) {
		t.Fatal("Cô Ba không bày được hàng")
	}
	if len(g.stalls) == 0 {
		t.Fatal("bày xong mà không có cái sạp nào")
	}
	return p, g.stalls
}

// Bậc sạp = cảnh giới Cô Ba + 1 (spec v6.10), tối đa bằng bậc cao nhất của bảng §8.2.
func TestStallTierIsOwnerRealmPlusOne(t *testing.T) {
	g := skillGame(t)
	p, stalls := coBaAt(t, g, 20) // cấp 20 = Kim Đan (chỉ số 2)
	wantTier := g.Prog.Realm(p.Level) + 1
	if wantTier != 3 {
		t.Fatalf("cấp 20 ra cảnh giới %d, bài test giả định Kim Đan (2)", g.Prog.Realm(p.Level))
	}
	for _, st := range stalls {
		if st.Tier != wantTier {
			t.Errorf("sạp bậc %d, muốn %d (cảnh giới Cô Ba + 1)", st.Tier, wantTier)
		}
	}

	// Kim Đan và thấp hơn: không chém được, phải đi vòng.
	for _, lv := range []int{0, 20} {
		o := put(g, "cu_rua", Vec{X: 9000, Y: 9000}, lv, 10)
		if g.CanCutObstacle(o, wantTier) {
			t.Errorf("cấp %d (cảnh giới %d) chém được sạp bậc %d", lv, g.Prog.Realm(o.Level), wantTier)
		}
		g.RemovePlayer(o.ID)
	}
	// Nguyên Anh (chỉ số 3): chém được.
	o := put(g, "cu_rua", Vec{X: 9000, Y: 9000}, 30, 10)
	if g.Prog.Realm(o.Level) != 3 {
		t.Fatalf("cấp 30 ra cảnh giới %d, bài test giả định Nguyên Anh (3)", g.Prog.Realm(o.Level))
	}
	if !g.CanCutObstacle(o, wantTier) {
		t.Error("Nguyên Anh phải chém được sạp bậc 3")
	}
}

// Bậc sạp không bao giờ vượt bậc cao nhất của bảng §8.2 — Đạo Tổ bày sạp thì chính Đạo Tổ
// (và mọi người) vẫn phải chém được, nếu không nó thành biên khu di động.
func TestStallTierCappedAtTableMax(t *testing.T) {
	g := skillGame(t)
	p := put(g, "co_ba_ganh", Vec{X: 3000, Y: 3000}, 10, 20)
	p.Facing = 0
	g.ForceState(p, g.Prog.TienLevel(4), 20, p.Pos) // Đạo Tổ = 13
	g.AddSkillCharge(p)
	g.CastSkill(p)

	maxTier := g.Cfg.Obstacles.MaxTier()
	for _, st := range g.stalls {
		if st.Tier > maxTier {
			t.Errorf("sạp bậc %d vượt bậc cao nhất của bảng (%d)", st.Tier, maxTier)
		}
		if st.Tier >= config.TierBorder {
			t.Error("sạp có bậc ∞ — nó thành biên khu di động")
		}
	}
}

// "Cô Ba tự đi xuyên sạp của mình" — đây là toàn bộ giá trị tấn công của kỹ năng: bỏ chạy
// chắc chắn trước người ngang cấp.
func TestStallOwnerPassesThrough(t *testing.T) {
	g := skillGame(t)
	p, stalls := coBaAt(t, g, 20)
	inside := stalls[len(stalls)/2].Shape.Center()

	if got := g.resolveBlockingFor(p, inside, g.BodyR(p)); got.Dist(inside) > 1e-9 {
		t.Errorf("Cô Ba bị chính sạp của mình chặn (dịch %.1f)", got.Dist(inside))
	}
	// Người ngang cấp thì KHÔNG: đó là nửa còn lại của cùng một luật.
	o := put(g, "cu_rua", Vec{X: 9000, Y: 9000}, 20, 10)
	if got := g.resolveBlockingFor(o, inside, g.BodyR(o)); got.Dist(inside) < 1e-9 {
		t.Error("người Kim Đan đi xuyên được sạp bậc 3")
	}
}

// "Shipper đang Gom Đơn đi xuyên qua" (§6.3 nguyên tắc 4: hai người khắc nhau).
func TestStallGhostPassesThrough(t *testing.T) {
	g := skillGame(t)
	_, stalls := coBaAt(t, g, 20)
	inside := stalls[len(stalls)/2].Shape.Center()

	sh := put(g, "shipper", Vec{X: 9000, Y: 9000}, 20, 10)
	if got := g.resolveBlockingFor(sh, inside, g.BodyR(sh)); got.Dist(inside) < 1e-9 {
		t.Fatal("Shipper chưa dùng kỹ năng mà đã xuyên được sạp")
	}
	g.AddSkillCharge(sh)
	g.CastSkill(sh)
	if got := g.resolveBlockingFor(sh, inside, g.BodyR(sh)); got.Dist(inside) > 1e-9 {
		t.Error("Shipper đang Gom Đơn vẫn bị dãy sạp chặn")
	}
}

// "Sau 6s sạp biến mất."
func TestStallExpires(t *testing.T) {
	g := skillGame(t)
	coBaAt(t, g, 20)
	n := len(g.stalls)

	g.Time += g.skillNum("bay_sap_secs") - 0.01
	g.updateStalls()
	if len(g.stalls) != n {
		t.Fatalf("còn %d/%d sạp trước hạn — biến mất quá sớm", len(g.stalls), n)
	}
	g.Time += 0.02
	g.updateStalls()
	if len(g.stalls) != 0 {
		t.Errorf("qua 6 giây vẫn còn %d sạp", len(g.stalls))
	}
}

// Dãy sạp phải là một VÒNG CUNG cách Cô Ba `bay_sap_dist` U, dài `bay_sap_len` U.
//
// Cung chứ không đường thẳng: đường thẳng vuông góc hướng nhắm thì hai đầu xa hơn khúc
// giữa, nên con mồi chạy chéo lách được đúng chỗ dãy mỏng nhất.
func TestStallArcGeometry(t *testing.T) {
	g := skillGame(t)
	p, stalls := coBaAt(t, g, 20)
	dist := g.skillRange(p, "bay_sap_dist")
	arc := g.skillRange(p, "bay_sap_len")

	if len(stalls) != g.skillInt("bay_sap_count") {
		t.Errorf("bày %d sạp, config nói %d", len(stalls), g.skillInt("bay_sap_count"))
	}
	for _, st := range stalls {
		if d := st.Shape.Center().Dist(p.Pos); math.Abs(d-dist) > 1 {
			t.Errorf("sạp cách %.0f, muốn %.0f (3 U)", d, dist)
		}
	}
	// Hai đầu dãy cách nhau xấp xỉ độ dài cung (sai số do cung không phải dây cung).
	ends := stalls[0].Shape.Center().Dist(stalls[len(stalls)-1].Shape.Center())
	if ends < arc*0.6 || ends > arc*1.05 {
		t.Errorf("hai đầu dãy cách %.0f, dãy phải dài ~%.0f (6 U)", ends, arc)
	}
	// Dãy nằm PHÍA TRƯỚC (Facing = 0 → +X), không bao quanh.
	for _, st := range stalls {
		if st.Shape.Center().X <= p.Pos.X {
			t.Error("có sạp bày ra phía sau lưng")
		}
	}
}

// "Sạp không đặt đè lên người (đoạn nào trùng người thì bỏ trống đoạn đó)."
//
// Bỏ trống chứ không đẩy người ra: đẩy người là cho Cô Ba một đòn dịch chuyển cưỡng chế
// miễn phí, và dãy sạp thành vũ khí thay vì địa hình.
func TestStallSkipsSpotsWithPeopleOnThem(t *testing.T) {
	g := skillGame(t)
	probe := put(g, "co_ba_ganh", Vec{X: 3000, Y: 3000}, 20, 20)
	probe.Facing = 0
	dist := g.skillRange(probe, "bay_sap_dist")
	g.RemovePlayer(probe.ID)

	// Đứng đúng giữa cung, nơi cái sạp giữa dãy sẽ rơi vào.
	blocker := put(g, "cu_rua", Vec{X: 3000 + dist, Y: 3000}, 20, 10)
	_, stalls := coBaAt(t, g, 20)

	for _, st := range stalls {
		if _, hit := st.Shape.PushOut(blocker.Pos, g.BodyR(blocker)); hit {
			t.Fatalf("có sạp đè lên người đang đứng ở %+v", blocker.Pos)
		}
	}
	if len(stalls) >= g.skillInt("bay_sap_count") {
		t.Error("không bỏ trống đoạn nào — người đứng giữa cung mà vẫn bày đủ 5 sạp")
	}
}

// "Dẫm vỏ chuối → trượt + choáng 1s." Vỏ chuối rải dọc CHÂN dãy, tức phía Cô Ba: con mồi
// bị dãy sạp chặn, quay lại thì dẫm vỏ.
func TestBananaPeelsSlipAndStun(t *testing.T) {
	g := skillGame(t)
	p, _ := coBaAt(t, g, 20)
	if len(g.slicks) == 0 {
		t.Fatal("bày hàng mà không rải vỏ chuối nào")
	}
	peel := g.slicks[0]
	if d := peel.at.Dist(p.Pos); d >= g.skillRange(p, "bay_sap_dist") {
		t.Errorf("vỏ chuối cách %.0f, phải nằm trong cung sạp (%.0f)", d, g.skillRange(p, "bay_sap_dist"))
	}

	o := put(g, "cu_rua", Vec{X: peel.at.X, Y: peel.at.Y}, 20, 10)
	g.updateSkillState(1.0 / 30)
	if !o.HasStatus(StSlip, g.Time) {
		t.Error("dẫm vỏ chuối mà không trượt")
	}
	if !o.HasStatus(StStun, g.Time) {
		t.Error("dẫm vỏ chuối mà không choáng")
	}
}

// Chủ sạp không chém được sạp của mình: nó đi xuyên người cô, nên kiếm cô cũng đi xuyên nó.
// Nếu không, Cô Ba vừa bày là tự xoá ngay trong nhịp kiếm kế tiếp.
func TestOwnerCannotCutOwnStalls(t *testing.T) {
	g := skillGame(t)
	p, stalls := coBaAt(t, g, 20)
	n := len(stalls)
	// Dí Cô Ba vào sát cái sạp giữa dãy để kiếm chắc chắn quét trúng.
	p.Pos = stalls[n/2].Shape.Center()
	for i := 0; i < 10; i++ {
		g.hitStalls(p)
		g.Time += 1.0 / 30
	}
	if len(g.stalls) != n {
		t.Errorf("Cô Ba tự chém mất %d sạp của mình", n-len(g.stalls))
	}
}
