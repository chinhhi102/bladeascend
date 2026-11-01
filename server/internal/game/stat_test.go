package game

import (
	"encoding/json"
	"math"
	"os"
	"testing"
)

type statVector struct {
	Name  string    `json:"name"`
	Base  float64   `json:"base"`
	Floor float64   `json:"floor"`
	Cap   float64   `json:"cap"`
	Terms []float64 `json:"terms"`
	Want  float64   `json:"want"`
}

// Bộ số mẫu dùng chung với client (`npm run check:stat`) — spec v6.32 §4.0.
func TestStatVectors(t *testing.T) {
	b, err := os.ReadFile("testdata/stat_vectors.json")
	if err != nil {
		t.Fatal(err)
	}
	var f struct {
		Vectors []statVector `json:"vectors"`
	}
	if err := json.Unmarshal(b, &f); err != nil {
		t.Fatal(err)
	}
	for _, v := range f.Vectors {
		got := Stat(v.Base, StatOpts{Floor: v.Floor, Cap: v.Cap}, v.Terms...)
		if math.Abs(got-v.Want) > 1e-6 {
			t.Errorf("%s: %.6f, muốn %.6f", v.Name, got, v.Want)
		}
	}
}

// Cộng, không nhân: Nam Châm + Gom Đơn + cảnh giới trên Shipper ra đúng ví dụ spec §4.0 khi
// Nam Châm +0.5 — và KHÁC hẳn cách nhân dồn cũ (≈ 332).
func TestStatAdditiveNotCompound(t *testing.T) {
	add := Stat(90, StatOpts{Cap: 4}, math.Sqrt(5)-1, 0.5, 0.1)
	compound := 90 * math.Sqrt(5) * 1.5 * 1.1
	if math.Abs(add-255.246) > 0.01 || math.Abs(add-compound) < 50 {
		t.Fatalf("cộng = %.1f (muốn ≈ 255), nhân dồn = %.1f", add, compound)
	}
}

// §6.4: nội tại vào đúng 4 chỉ số; Shipper chạy nhanh nhất, Cụ Rùa chậm nhất.
func TestInnateAppliesToStats(t *testing.T) {
	g := quietGame(t, nil)
	ship := g.AddPlayer(SpawnOptions{Name: "S", Char: "shipper", Swords: 1})
	rua := g.AddPlayer(SpawnOptions{Name: "R", Char: "cu_rua", Swords: 1})
	base := g.Cfg.Player.MoveBase
	if got := g.MoveSpeed(ship); math.Abs(got-base*1.15) > 1e-6 {
		t.Fatalf("Shipper chạy %.2f, muốn %.2f (+15%%)", got, base*1.15)
	}
	if got := g.MoveSpeed(rua); math.Abs(got-base*0.85) > 1e-6 {
		t.Fatalf("Cụ Rùa chạy %.2f, muốn %.2f (−15%%)", got, base*0.85)
	}
	w := 2 * math.Pi / g.Cfg.Sword.RevSecs
	if got := g.Omega(ship); math.Abs(got-w*0.95) > 1e-9 {
		t.Fatalf("Shipper quay %.3f, muốn %.3f (−5%%)", got, w*0.95)
	}
	if got := g.durBonus(rua); math.Abs(got-0.15) > 1e-9 {
		t.Fatalf("Cụ Rùa bền %+.2f, muốn +0.15", got)
	}
	mt := g.AddPlayer(SpawnOptions{Name: "M", Char: "map_tra_sua", Swords: 1})
	if got := g.PickupRadius(mt); math.Abs(got-g.Cfg.Player.PickupR*1.15) > 1e-6 {
		t.Fatalf("Thằng Mập nhặt %.1f, muốn %.1f (+15%%)", got, g.Cfg.Player.PickupR*1.15)
	}
}

// v6.31: Gom Đơn chỉ +0.1 base bán kính nhặt, không còn hút 3 U.
func TestGomDonPickupOnlyTenPercent(t *testing.T) {
	g := quietGame(t, nil)
	p := g.AddPlayer(SpawnOptions{Name: "S", Char: "shipper", Level: 8 * g.Prog.Tiers(), Swords: 1})
	before := g.PickupRadius(p)
	g.ApplyStatus(p, StGhost, 4)
	after := g.PickupRadius(p)
	if math.Abs(after-before-0.1*g.Cfg.Player.PickupR) > 1e-6 {
		t.Fatalf("Gom Đơn thêm %.1f, muốn đúng 0.1 base = %.1f", after-before, 0.1*g.Cfg.Player.PickupR)
	}
	if after >= 3*g.Cfg.Player.PickupR*p.Scale {
		t.Fatalf("Gom Đơn vẫn hút tới 3 U (%.0f)", after)
	}
}

// Tốc Kiếm cộng nội tại, kẹp trần ×2 (§4.0).
func TestSpinCap(t *testing.T) {
	g := quietGame(t, nil)
	p := g.AddPlayer(SpawnOptions{Name: "P", Char: "su_phu_mi", Swords: 1})
	g.ApplyBuff(p, "spin")
	w := 2 * math.Pi / g.Cfg.Sword.RevSecs
	want := math.Min(w*(1+0.10+(g.Cfg.Buffs["spin"].Mult-1)), w*g.Cfg.Stats.SpinCap)
	if got := g.Omega(p); math.Abs(got-want) > 1e-9 {
		t.Fatalf("Sư Phụ + Tốc Kiếm quay %.3f, muốn %.3f", got, want)
	}
}
