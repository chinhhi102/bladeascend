package game

import (
	"math"
	"testing"

	"bladeascend/server/internal/config"
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
	giveCharge(p)
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
	g.ForceState(p, g.Prog.TienLevel(4, 0), 20, p.Pos) // Đạo Tổ = 13
	giveCharge(p)
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
	giveCharge(sh)
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

// v6.48: "Lúc sạp bật lên, địch đứng trên đường sạp bị ĐẨY VĂNG ra ngoài (theo hướng ra xa
// Cô Ba)" — thay cho luật cũ "bỏ trống đoạn có người". Dãy sạp hở đúng ngay chỗ con mồi
// đang đứng nghĩa là kỹ năng tự hỏng khi dùng đúng lúc nhất.
func TestStallPushesPeopleOffTheLine(t *testing.T) {
	g := skillGame(t)
	probe := put(g, "co_ba_ganh", Vec{X: 3000, Y: 3000}, 20, 20)
	probe.Facing = 0
	dist := g.skillRange(probe, "bay_sap_dist")
	g.RemovePlayer(probe.ID)

	// Đứng đúng giữa cung, nơi cái sạp giữa dãy sẽ rơi vào.
	blocker := put(g, "cu_rua", Vec{X: 3000 + dist, Y: 3000}, 20, 10)
	co, stalls := coBaAt(t, g, 20)

	if len(stalls) != g.skillInt("bay_sap_count") {
		t.Errorf("bày được %d sạp, muốn đủ %d — người đứng chắn không còn làm hở dãy",
			len(stalls), g.skillInt("bay_sap_count"))
	}
	if blocker.KnockUntil <= g.Time {
		t.Fatal("người đứng trên đường sạp không bị đẩy")
	}
	// Đẩy RA XA Cô Ba, không phải về phía cô.
	if blocker.KnockVel.Dot(blocker.Pos.Sub(co.Pos).Norm()) <= 0 {
		t.Error("bị đẩy về phía Cô Ba — phải đẩy ra xa")
	}
}

// "Địch trong 2 U quanh dãy sạp lúc bật bị buộc vào sạp gần nhất 3s: chỉ đi được trong
// vòng 2 U quanh sạp đó" (v6.48 §6.3).
func TestStallTethersNearbyFoes(t *testing.T) {
	g := skillGame(t)
	probe := put(g, "co_ba_ganh", Vec{X: 3000, Y: 3000}, 20, 20)
	probe.Facing = 0
	dist := g.skillRange(probe, "bay_sap_dist")
	leash := g.skillRange(probe, "bay_sap_tether_leash")
	g.RemovePlayer(probe.ID)

	near := put(g, "cu_rua", Vec{X: 3000 + dist + g.skillRange(probe, "bay_sap_tether_r")*0.5, Y: 3000}, 20, 10)
	far := put(g, "cu_rua", Vec{X: 3000 + dist*4, Y: 3000}, 20, 10)
	coBaAt(t, g, 20)

	if !near.HasStatus(StTether, g.Time) {
		t.Fatal("địch đứng sát dãy sạp mà không bị buộc")
	}
	if far.HasStatus(StTether, g.Time) {
		t.Error("địch ở xa cũng bị buộc")
	}
	// Chạy thẳng ra xa trong 2 giây: không bao giờ vượt quá dây.
	near.KnockUntil = 0
	for i := 0; i < 60; i++ {
		g.SetInput(near.ID, uint32(i+1), 0, true, false)
		g.movePlayer(near, 1.0/30)
		if d := near.Pos.Dist(near.tetherAt); d > leash+1 {
			t.Fatalf("thoát khỏi dây: cách sạp %.0f, dây dài %.0f", d, leash)
		}
	}
}

// v6.52 (Q51): Cô Ba KHÔNG còn rải vỏ chuối. Bài này canh đúng việc đó — bỏ một cơ chế mà
// không có bài kiểm thì lần sau rất dễ có người "sửa lại cho đủ" theo §6.5 bản cũ.
func TestNoBananaPeelsAnyMore(t *testing.T) {
	g := skillGame(t)
	co, _ := coBaAt(t, g, 20)
	// Đứng đúng chân dãy sạp — chỗ vỏ chuối từng nằm.
	at := co.Pos.Add(Vec{X: g.skillRange(co, "bay_sap_dist") * 0.9})
	o := put(g, "cu_rua", at, 20, 10)
	o.Status[StTether] = 0 // bỏ dây buộc để chỉ còn đo phần trượt/choáng
	g.updateSkillState(1.0 / 30)
	if o.HasStatus(StSlip, g.Time) || o.HasStatus(StStun, g.Time) {
		t.Fatal("vẫn còn vỏ chuối: đứng chân dãy sạp mà bị trượt/choáng")
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

// v6.48: sạp BẤT HOẠI — người hơn Cô Ba cả cảnh giới cũng không chém bay được, và thân
// vẫn bị chặn. Chỉ Cô Ba và Shipper đang Gom Đơn đi xuyên.
func TestStallsAreIndestructible(t *testing.T) {
	g := skillGame(t)
	co, stalls := coBaAt(t, g, 10)
	if len(stalls) == 0 {
		t.Fatal("không bày được sạp nào")
	}
	st := stalls[0]
	c := st.Shape.Center()

	// Người cảnh giới CAO HƠN hẳn, đứng sát sạp, kiếm quét qua.
	strong := put(g, "cu_rua", c.Add(Vec{X: g.BodyR(co) * 2}), g.Prog.TopMortalLevel(), 40)
	g.hitStalls(strong)
	if len(g.StallList()) != len(stalls) {
		t.Fatalf("còn %d sạp, muốn %d — sạp phải bất hoại", len(g.StallList()), len(stalls))
	}

	// …và thân họ vẫn bị chặn.
	inside := c
	if got := g.stallBlocking(strong, inside, g.BodyR(strong)); got.Dist(inside) < 1 {
		t.Error("người cảnh giới cao đi xuyên được sạp")
	}
	// Shipper đang Gom Đơn thì đi xuyên.
	ship := put(g, "shipper", c, 10, 10)
	g.ApplyStatus(ship, StGhost, 4)
	if got := g.stallBlocking(ship, inside, g.BodyR(ship)); got.Dist(inside) > 1 {
		t.Error("Shipper đang Gom Đơn phải đi xuyên được dãy sạp")
	}
}
