package game

import "testing"

// Sinh THẲNG ở một cảnh giới Tiên (bot chép cảnh giới người hạng N) phải có thanh Tiên khí
// ở mốc khởi điểm, không phải 0.
//
// Lỗi thật đã gặp: bot Chân Tiên bị chém mà **không rớt cảnh giới** — vẫn bị đẩy lùi, vẫn
// hiện "Hit", nhưng đứng nguyên ở Chân Tiên. Nguyên nhân là hai nguồn sự thật nói ngược
// nhau: `Level` ở Tiên giới nhưng `TienKhi = 0`, nên `TienRealmOf` trả -1 và `robTienKhi`
// thoát ngay ở dòng đầu.
func TestSpawnAtTienRealmSeatsTienKhi(t *testing.T) {
	g := rulesGame(t)
	for idx := 0; idx < g.Prog.TienCount(); idx++ {
		lv := g.Prog.TienLevel(idx)
		p := g.AddPlayer(SpawnOptions{Name: "B", IsBot: true, Level: lv, Swords: 10})
		if got := g.TienRealmOf(p); got != idx {
			t.Errorf("sinh ở %s (cấp %d): TienRealmOf = %d, muốn %d — thanh Tiên khí %v",
				g.Prog.RealmName(g.Prog.Realm(lv)), lv, got, idx, p.TienKhi)
		}
		g.RemovePlayer(p.ID)
	}
}

// Và hệ quả của nó: bot Chân Tiên bị chém thân thì PHẢI rớt khỏi Tiên giới.
func TestChanTienBotDropsWhenHit(t *testing.T) {
	g := rulesGame(t)
	lv := g.Prog.TienLevel(0) // Chân Tiên
	victim := g.AddPlayer(SpawnOptions{Name: "B", IsBot: true, Level: lv, Swords: 10})
	attacker := g.AddPlayer(SpawnOptions{Name: "A", Level: lv, Swords: 10})
	if !g.Prog.IsTien(g.Prog.Realm(victim.Level)) {
		t.Fatalf("cấp %d không phải Tiên giới — bài test đặt sai", lv)
	}

	g.applyBodyHit(attacker, victim)

	if g.Prog.IsTien(g.Prog.Realm(victim.Level)) {
		t.Fatalf("Chân Tiên bị chém mà vẫn ở Tiên giới (cấp %d)", victim.Level)
	}
	if victim.TienKhi != 0 {
		t.Errorf("rớt khỏi Tiên giới mà thanh Tiên khí còn %v, phải về 0", victim.TienKhi)
	}
}
