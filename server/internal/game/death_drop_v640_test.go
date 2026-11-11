package game

import (
	"math"
	"testing"
)

// GAME_SPEC §7 (v6.40) — hai dòng nghiệm thu mới của §17.
//
// "Kim Đan 60 kiếm bị mài mất 50 trong 8s rồi chết với 10 kiếm → xác rơi 60 kiếm (≤ 12 bó);
// kiếm bị Hút Mì cướp không tính vào số rơi."

// kimDanVictim dựng một nạn nhân Kim Đan đầy 60 kiếm, đứng sẵn ở chỗ trống.
func kimDanVictim(t *testing.T, g *Game) *Player {
	t.Helper()
	lv := 2 * g.Prog.Tiers() // Kim Đan tầng 1
	p := g.AddPlayer(SpawnOptions{Name: "B", Level: lv, Swords: 60})
	if got := g.Prog.MaxSwords(g.Prog.Realm(p.Level)); got != 60 {
		t.Fatalf("Kim Đan trần %d kiếm, bài này viết theo mốc 60", got)
	}
	return p
}

func TestDeathDropCountsSwordsBrokenInWindow(t *testing.T) {
	g := quietGame(t, nil)
	b := kimDanVictim(t, g)
	place(g, b, Vec{X: 5000, Y: 5000}, 0)

	// Mài mất 50 kiếm rải trong 8 giây — vẫn nằm trong cửa sổ 10 giây.
	for _, at := range []float64{0, 3, 8} {
		g.Time = at
		g.breakSwords(b, 50/3)
	}
	g.breakSwords(b, 50-3*(50/3))
	if b.Ring.Count != 10 {
		t.Fatalf("mài xong còn %d kiếm, muốn 10", b.Ring.Count)
	}

	g.Time = 8.5
	g.dropDeathSwords(b)
	bundles, swords := countItems(g, "bundle")
	if int(swords) != 60 {
		t.Fatalf("xác rơi %.0f kiếm, muốn 60 (50 bể trong 8s + 10 còn lại)", swords)
	}
	if bundles != g.Cfg.Player.DeathDropMaxBundles {
		t.Fatalf("xác rơi %d bó, muốn %d bó chia đều", bundles, g.Cfg.Player.DeathDropMaxBundles)
	}
	if b.Ring.Count != 0 {
		t.Fatalf("chết rồi còn %d kiếm trên người", b.Ring.Count)
	}
}

// Kiếm bể TRƯỚC cửa sổ thì không tính: sổ là cửa sổ TRƯỢT, không phải tổng cả ván.
func TestDeathDropIgnoresLossesOutsideWindow(t *testing.T) {
	g := quietGame(t, nil)
	b := kimDanVictim(t, g)
	place(g, b, Vec{X: 5000, Y: 5000}, 0)

	g.Time = 0
	g.breakSwords(b, 40) // 40 cây bể từ rất lâu
	g.Time = 0.5 + g.Cfg.Player.DeathDropLostWindowSecs
	g.breakSwords(b, 5) // 5 cây bể ngay trước khi chết

	g.Time += 0.1
	g.dropDeathSwords(b)
	_, swords := countItems(g, "bundle")
	if want := 5 + 15; int(swords) != want {
		t.Fatalf("xác rơi %.0f kiếm, muốn %d (5 bể trong cửa sổ + 15 còn lại; 40 cây cũ đã hết hạn)", swords, want)
	}
}

// Kiếm bị CƯỚP không vào sổ: Hút Mì (Sư Phụ Mì) và Trộm Kiếm (Ông Ba Võng) chuyển kiếm sang
// vòng kẻ cướp, tính lại khi nạn nhân chết là nhân đôi số kiếm trong phòng.
func TestDeathDropExcludesStolenSwords(t *testing.T) {
	g := quietGame(t, nil)
	b := kimDanVictim(t, g)
	place(g, b, Vec{X: 5000, Y: 5000}, 0)
	thief := g.AddPlayer(SpawnOptions{Name: "A", Char: "su_phu_mi", Level: 2 * g.Prog.Tiers(), Swords: 1})
	place(g, thief, Vec{X: 5400, Y: 5000}, 0)

	g.Time = 1
	before := b.Ring.Count
	g.stealSwords(thief, b) // Hút Mì
	stolen := before - b.Ring.Count
	if stolen <= 0 {
		t.Fatal("Hút Mì không cướp được cây nào, bài kiểm vô nghĩa")
	}

	g.Time = 2
	remain := b.Ring.Count
	g.dropDeathSwords(b)
	_, swords := countItems(g, "bundle")
	if int(swords) != remain {
		t.Fatalf("xác rơi %.0f kiếm, muốn %d (chỉ số còn trên người; %d cây bị cướp không tính)",
			swords, remain, stolen)
	}
}

// Trộm Kiếm của Ông Ba: kiếm bể khi va với ông bay sang vòng ông — cũng không vào sổ.
func TestDeathDropExcludesTromKiemSteal(t *testing.T) {
	g := quietGame(t, nil)
	b := kimDanVictim(t, g)
	thief := g.AddPlayer(SpawnOptions{Name: "A", Char: "ong_ba_vong", Level: 2 * g.Prog.Tiers(), Swords: 20})
	g.Time = 1
	thief.StealUntil = g.Time + 5

	before := b.Ring.Count
	g.stealOnClash(thief, b, 0, 7)
	stolen := before - b.Ring.Count
	if stolen <= 0 {
		t.Fatal("Trộm Kiếm không lấy được cây nào, bài kiểm vô nghĩa")
	}
	if got := g.swordsLostRecently(b); got != 0 {
		t.Fatalf("sổ ghi %d cây bị cướp, phải ghi 0", got)
	}
}

// Trần tổng = kiếm tối đa của cảnh giới nạn nhân: bị mài đi mài lại nhiều đợt trong cửa sổ
// không đẻ ra nhiều kiếm hơn vòng của họ chứa nổi.
func TestDeathDropCappedAtRealmMax(t *testing.T) {
	g := quietGame(t, nil)
	b := kimDanVictim(t, g)
	place(g, b, Vec{X: 5000, Y: 5000}, 0)

	// Mài sạch rồi nhặt lại đầy, ba lượt — tổng số cây bể vượt xa trần 60.
	for i := 0; i < 3; i++ {
		g.Time = float64(i)
		g.breakSwords(b, 60)
		g.AddSwords(b, 60)
	}
	g.Time = 3
	g.dropDeathSwords(b)
	_, swords := countItems(g, "bundle")
	if int(swords) != 60 {
		t.Fatalf("xác rơi %.0f kiếm, muốn đúng trần Kim Đan = 60", swords)
	}
}

// "Bán kính nhặt: Luyện Khí 1 = 90 px, Đạo Tổ = 540 px; Shipper Đại Thừa 10 + Gom Đơn +
// Nam Châm = 720 px; Thằng Mập Đạo Tổ + Nam Châm = 891 px" (§17, v6.40).
func TestPickupRadiusSpecNumbers(t *testing.T) {
	g := quietGame(t, nil)
	tiers := g.Prog.Tiers()
	eq := func(name string, got, want float64) {
		t.Helper()
		if math.Abs(got-want) > 0.5 {
			t.Errorf("%s: %.1f px, spec nói %.1f px", name, got, want)
		}
	}
	// Nhân vật nội tại hút 0 để con số là của riêng luật cảnh giới.
	plain := ""
	for c, in := range g.Cfg.Innate {
		if in.Pickup == 0 {
			plain = c
			break
		}
	}
	if plain == "" {
		t.Fatal("không có nhân vật nào nội tại hút = 0")
	}
	low := g.AddPlayer(SpawnOptions{Name: "L", Char: plain, Level: 0, Swords: 1})
	eq("Luyện Khí 1", g.PickupRadius(low), 90)

	// v6.46: Tiên giới có 10 tầng, scale tăng theo TẦNG — 540 px của spec là Đạo Tổ tầng 10
	// (scale 6.00), không phải tầng 1 (5.82).
	daoTo := g.AddPlayer(SpawnOptions{Name: "D", Char: plain, Level: 13*tiers + tiers - 1, Swords: 1})
	if g.Prog.RealmName(g.Prog.Realm(daoTo.Level)) != "Đạo Tổ" {
		t.Fatalf("cấp %d không phải Đạo Tổ", daoTo.Level)
	}
	eq("Đạo Tổ", g.PickupRadius(daoTo), 540)

	ship := g.AddPlayer(SpawnOptions{Name: "S", Char: "shipper", Level: 9*tiers - 1, Swords: 1})
	g.ApplyStatus(ship, StGhost, 4)
	g.ApplyBuff(ship, "magnet")
	eq("Shipper Đại Thừa 10 + Gom Đơn + Nam Châm", g.PickupRadius(ship), 720)

	map_ := g.AddPlayer(SpawnOptions{Name: "M", Char: "map_tra_sua", Level: 13*tiers + tiers - 1, Swords: 1})
	g.ApplyBuff(map_, "magnet")
	eq("Thằng Mập Đạo Tổ + Nam Châm", g.PickupRadius(map_), 891)
}
