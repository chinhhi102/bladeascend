package game

import (
	"math"

	"bladeascend/server/internal/config"
)

// Luật chung spec v6.32 §4.0: mọi thưởng/phạt CỘNG trên base, không nhân dồn.
//
//	giá_trị = base × clamp(1 + Σ khoản, sàn, trần)
//
// MỌI chỉ số có nhiều nguồn ảnh hưởng (bán kính nhặt, tốc chạy, tốc quay kiếm, độ bền kiếm,
// EXP nhận) đi qua `Stat`. Bên client có bản sao y hệt (`client/src/net/stat.ts`), hai bên
// cùng chạy bộ số mẫu `testdata/stat_vectors.json` — lệch một bên là test đỏ.
//
// Hệ số kiểu "× k" trong config cũ (buffs.*.mult, durability_mult, brace_durability) được
// hiểu là khoản cộng `k − 1` (spec §4.0) — `Bonus` làm phép đổi đó ở một chỗ.

// StatOpts: sàn và trần của tổng (bội số của base). 0 = không kẹp phía đó.
type StatOpts struct {
	Floor float64
	Cap   float64
}

// Stat trả `base × clamp(1 + Σ terms)`. Đây là hàm DUY NHẤT được phép gộp các nguồn.
func Stat(base float64, opts StatOpts, terms ...float64) float64 {
	f := 1.0
	for _, t := range terms {
		f += t
	}
	if opts.Floor > 0 && f < opts.Floor {
		f = opts.Floor
	}
	if opts.Cap > 0 && f > opts.Cap {
		f = opts.Cap
	}
	if f < 0 {
		f = 0
	}
	return base * f
}

// Bonus đổi hệ số nhân cũ `× k` thành khoản cộng `k − 1`; k ≤ 0 (chưa cấu hình) là 0.
func Bonus(mult float64) float64 {
	if mult <= 0 {
		return 0
	}
	return mult - 1
}

// innate: nội tại của nhân vật p (spec v6.33 §6.4). Đàn em mang nội tại của đại ca? Không —
// đàn em là NPC, không có nhân vật riêng; nhân vật lạ cũng ra 0.
func (g *Game) innate(p *Player) config.Innate {
	if p.IsMinion {
		return config.Innate{}
	}
	return g.Cfg.Innate[p.Char]
}

// buffBonus: khoản cộng của buff `id` nếu p đang có, ngược lại 0.
func (g *Game) buffBonus(p *Player, kind int, id string) float64 {
	if !p.HasBuff(kind, g.Time) {
		return 0
	}
	return BuffTerm(g.Cfg.Buffs[id])
}

// BuffTerm: khoản cộng của một buff — `bonus` nếu config ghi (dạng §4.0), không thì `mult − 1`.
func BuffTerm(b config.Buff) float64 {
	if b.Bonus != 0 {
		return b.Bonus
	}
	return Bonus(b.Mult)
}

// scalePickupTerm: phần cảnh giới của bán kính nhặt `scale^e − 1` (e = 0.5 → √scale − 1),
// để Luyện Khí 1 (scale 1) đúng bằng base.
func (g *Game) scalePickupTerm(p *Player) float64 {
	return math.Pow(p.Scale, g.pickupScaleExp()) - 1
}

// expBonus: khoản cộng EXP nhận (Song Bội Đan +1.0).
func (g *Game) expBonus(p *Player) float64 { return g.buffBonus(p, BuffExp2, "exp2") }
