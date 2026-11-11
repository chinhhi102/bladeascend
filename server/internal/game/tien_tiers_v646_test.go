package game

import (
	"math"
	"testing"
)

// GAME_SPEC §3.4 / §11.2 (v6.46) — ba dòng nghiệm thu mới của §17.

// "Đại Thừa 10 chưa đầy: HUD thanh Linh khí, viên Linh khí sáng, Tiên khí mờ;
// đầy 694 960 → đổi sang thanh Tiên khí."
//
// Phía server cờ này là `BanksTienKhi`, và nó đi ra client trong `FlagCanTien`. Client KHÔNG
// được tự đoán theo cảnh giới — đó đúng là lỗi 07/11: đứng Đại Thừa tầng 10 mà thanh chưa
// đầy thì viên Linh khí đã mờ đi, trong khi người chơi vẫn đang cần nhặt chúng.
func TestCanAbsorbTienOnlyWhenBarFull(t *testing.T) {
	g := quietGame(t, nil)
	top := g.Prog.TopMortalLevel()
	p := g.AddPlayer(SpawnOptions{Name: "P", Level: top, Swords: 1})
	full := g.Prog.MaxExp()

	// Đứng Đại Thừa tầng 10 nhưng thanh chưa đầy.
	g.setTotalExp(p, full-1)
	if g.BanksTienKhi(p) {
		t.Fatalf("thanh %.0f/%.0f chưa đầy mà đã sang đường Tiên khí", p.TotalExp, full)
	}
	if !g.CanAbsorbExp(p) || g.CanAbsorbTienKhi(p) {
		t.Fatal("chưa đầy thanh thì phải hấp thụ Linh khí, không hấp thụ Tiên khí")
	}

	// Đầy đúng mốc.
	g.setTotalExp(p, full)
	if !g.BanksTienKhi(p) {
		t.Fatalf("thanh đầy %.0f mà chưa sang đường Tiên khí", full)
	}
	if g.CanAbsorbExp(p) || !g.CanAbsorbTienKhi(p) {
		t.Fatal("đầy thanh rồi thì chỉ hấp thụ Tiên khí")
	}

	// Thiêu Đốt kéo xuống dưới mốc → đổi ngược lại.
	g.setTotalExp(p, full-1)
	if g.BanksTienKhi(p) {
		t.Fatal("tụt xuống dưới mốc mà vẫn ở đường Tiên khí")
	}
}

// "Tiên khí 160 → hiện Chân Tiên tầng 3 (154–238); 1 001 → Kim Tiên tầng 1;
// scale Chân Tiên 1 = 5,02, Đạo Tổ 10 = 6,00."
func TestTienTiersMatchSpec(t *testing.T) {
	g := quietGame(t, nil)
	pr := g.Prog
	if pr.TienTiers() != 10 {
		t.Fatalf("Tiên giới %d tầng, spec nói 10", pr.TienTiers())
	}

	// Mốc tầng của Chân Tiên: tầng 1 = 1–74, tầng 2 = 75–153, tầng 3 = 154–238.
	for _, c := range []struct {
		tier int
		top  float64
	}{{0, 74}, {1, 153}, {2, 238}, {9, 1000}} {
		if got := pr.TienTierTop(0, c.tier); math.Abs(got-c.top) > 0.5 {
			t.Errorf("Chân Tiên tầng %d trần %.0f, spec nói %.0f", c.tier+1, got, c.top)
		}
	}
	if got := pr.TienTierForKhi(160); got != 2 {
		t.Errorf("160 Tiên khí ra tầng %d, spec nói tầng 3", got+1)
	}
	// Đạo Tổ tầng 1 = 10 001–10 368.
	last := pr.TienCount() - 1
	if got := pr.TienTierTop(last, 0); math.Abs(got-10368) > 0.5 {
		t.Errorf("Đạo Tổ tầng 1 trần %.0f, spec nói 10 368", got)
	}

	// 1 001 → Kim Tiên tầng 1.
	if idx, tier := pr.TienRealmForKhi(1001), pr.TienTierForKhi(1001); idx != 1 || tier != 0 {
		t.Errorf("1 001 Tiên khí ra cảnh giới %d tầng %d, spec nói Kim Tiên tầng 1", idx, tier+1)
	}

	// Scale theo tầng: Chân Tiên 1 = 5,02 … Đạo Tổ 10 = 6,00.
	if got := pr.Scale(pr.TienLevel(0, 0)); math.Abs(got-5.02) > 1e-9 {
		t.Errorf("Chân Tiên tầng 1 scale %.3f, spec nói 5.02", got)
	}
	if got := pr.Scale(pr.TienLevel(last, pr.TienTiers()-1)); math.Abs(got-6.0) > 1e-9 {
		t.Errorf("Đạo Tổ tầng 10 scale %.3f, spec nói 6.00", got)
	}
	// Số kiếm tối đa vẫn theo CẢNH GIỚI, không theo tầng.
	lo, hi := pr.TienLevel(last, 0), pr.TienLevel(last, pr.TienTiers()-1)
	if pr.MaxSwords(pr.Realm(lo)) != pr.MaxSwords(pr.Realm(hi)) {
		t.Error("kiếm tối đa đổi theo tầng Tiên — spec nói chỉ theo cảnh giới")
	}
}

// Lên tầng trong cùng cảnh giới KHÔNG phải Thiên Kiếp; lên cảnh giới thì có.
func TestTienTierUpIsNotThienKiep(t *testing.T) {
	g := quietGame(t, nil)
	p := g.AddPlayer(SpawnOptions{Name: "P", Level: g.Prog.TopMortalLevel(), Swords: 1})
	g.setTotalExp(p, g.Prog.MaxExp())
	g.GainTienKhi(p, 1) // vào Chân Tiên tầng 1
	g.DrainEvents()

	g.GainTienKhi(p, 200) // 201 → vẫn Chân Tiên, nhưng lên tầng
	if got := g.Prog.Realm(p.Level); !g.Prog.IsTien(got) || g.Prog.TienIndex(got) != 0 {
		t.Fatalf("201 Tiên khí mà đã rời Chân Tiên (cảnh giới %d)", got)
	}
	if tier := g.Prog.Tier(p.Level); tier == 0 {
		t.Fatal("201 Tiên khí mà vẫn tầng 1 — `Level` không theo tầng")
	}
	for _, e := range g.DrainEvents() {
		if e.Kind == "Breakthrough" {
			t.Fatal("lên TẦNG mà báo Thiên Kiếp — chỉ lên cảnh giới mới được báo")
		}
	}

	g.GainTienKhi(p, 1000) // 1 201 → Kim Tiên: lần này phải có Thiên Kiếp
	saw := false
	for _, e := range g.DrainEvents() {
		if e.Kind == "Breakthrough" {
			saw = true
		}
	}
	if !saw {
		t.Fatal("lên cảnh giới Tiên mà không báo Thiên Kiếp")
	}
}

// "Phòng 40 bot Luyện Khí cùng tầng, Uy Danh 0: bảng không đổi thứ tự trong 60 s dù họ nhặt
// EXP liên tục." — v6.46 bỏ phá hoà bằng tổng EXP.
func TestRankingStableWhenTierAndScoreTie(t *testing.T) {
	g := quietGame(t, nil)
	const n = 40
	for i := 0; i < n; i++ {
		g.AddPlayer(SpawnOptions{Name: "Bot", Level: 0, Swords: 1, IsBot: true})
	}
	first := g.Ranking()
	want := make([]uint32, len(first))
	for i, p := range first {
		want[i] = p.ID
	}

	// Họ nhặt EXP liên tục — nhưng chưa ai đủ lên tầng, và Uy Danh vẫn 0.
	for round := 0; round < 60; round++ {
		for i, p := range g.PlayerList() {
			p.TotalExp += float64(i%7) + 1
		}
		got := g.Ranking()
		for i, p := range got {
			if p.ID != want[i] {
				t.Fatalf("vòng %d: bảng đổi thứ tự dù cùng tầng và cùng Uy Danh (dòng %d)", round, i)
			}
		}
	}

	// Vẫn phải đổi khi có người thật sự vượt lên.
	lead := g.Players[want[len(want)-1]]
	g.creditScore(lead, ScoreHit, 100)
	if top := g.Ranking()[0]; top.ID != lead.ID {
		t.Fatal("có người hơn Uy Danh mà bảng không đổi")
	}
}
