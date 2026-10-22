package game

import (
	"math"
	"sort"

	"kiemxoay/server/internal/config"
)

// Progression holds the precomputed cultivation tables (spec §3).
type Progression struct {
	cfg        config.Cultivation
	maxSwords  []int
	needed     []float64 // needed[L] = EXP to go from level L to L+1
	cumulative []float64 // cumulative[L] = total EXP required to reach level L
}

func NewProgression(c config.Cultivation, maxSwords []int) *Progression {
	levels := len(c.Realms) * c.TiersPerRealm
	p := &Progression{cfg: c, maxSwords: maxSwords}
	p.needed = make([]float64, levels)
	p.cumulative = make([]float64, levels)
	sum := 0.0
	for L := 0; L < levels; L++ {
		r, s := L/c.TiersPerRealm, L%c.TiersPerRealm
		p.cumulative[L] = sum
		p.needed[L] = math.Round(c.ExpBase * math.Pow(c.RealmMult, float64(r)) * (1 + c.TierStep*float64(s)))
		sum += p.needed[L]
	}
	return p
}

func (p *Progression) Levels() int { return len(p.needed) }

func (p *Progression) Tiers() int { return p.cfg.TiersPerRealm }

// ExpNeeded returns exp_needed(r, s) for level index L.
func (p *Progression) ExpNeeded(L int) float64 {
	if L < 0 || L >= len(p.needed) {
		return 0
	}
	return p.needed[L]
}

// Threshold returns the cumulative EXP required to reach level L.
func (p *Progression) Threshold(L int) float64 {
	if L <= 0 {
		return 0
	}
	if L >= len(p.cumulative) {
		return p.cumulative[len(p.cumulative)-1] + p.needed[len(p.needed)-1]
	}
	return p.cumulative[L]
}

// LevelFromExp returns the level index reached with totalExp (capped at the last level).
func (p *Progression) LevelFromExp(totalExp float64) int {
	i := sort.Search(len(p.cumulative), func(i int) bool { return p.cumulative[i] > totalExp })
	if i == 0 {
		return 0
	}
	return i - 1
}

// MaxExp is the EXP at which progression stops counting (top of the final tier).
func (p *Progression) MaxExp() float64 { return p.Threshold(len(p.needed)) }

func (p *Progression) Realm(L int) int { return L / p.cfg.TiersPerRealm }

func (p *Progression) Tier(L int) int { return L % p.cfg.TiersPerRealm }

// MaxSwords returns max_swords[r]; realms past the end of the array add 40 each (spec §3.1).
func (p *Progression) MaxSwords(r int) int {
	if len(p.maxSwords) == 0 {
		return 1
	}
	if r < len(p.maxSwords) {
		return p.maxSwords[max(r, 0)]
	}
	return p.maxSwords[len(p.maxSwords)-1] + 40*(r-len(p.maxSwords)+1)
}

// DropRealmExp is total_exp after falling one whole realm while keeping the tier and the
// progress through it (spec §5.1): Trúc Cơ 2 at 40% → Luyện Khí 2 at 40%.
func (p *Progression) DropRealmExp(exp float64) float64 {
	L := p.LevelFromExp(exp)
	to := L - p.cfg.TiersPerRealm
	if to < 0 {
		return 0
	}
	progress := 0.0
	if n := p.ExpNeeded(L); n > 0 {
		progress = math.Min(1, (exp-p.Threshold(L))/n)
	}
	return p.Threshold(to) + progress*p.ExpNeeded(to)
}

// Scale returns scale(L) = min + (max-min) * L / (levels-1); Tiên giới keeps growing by
// tien_scale_step per realm (spec §3.4: Chân Tiên 5.2 … Đạo Tổ 6.0).
func (p *Progression) Scale(L int) float64 {
	if r := p.Realm(L); p.IsTien(r) {
		return p.cfg.ScaleMax + p.cfg.TienScaleStep*float64(p.TienIndex(r)+1)
	}
	last := float64(len(p.needed) - 1)
	return p.cfg.ScaleMin + (p.cfg.ScaleMax-p.cfg.ScaleMin)*float64(L)/last
}

func (p *Progression) RealmName(r int) string {
	if r < 0 {
		return ""
	}
	if r < len(p.cfg.Realms) {
		return p.cfg.Realms[r]
	}
	if i := p.TienIndex(r); i < len(p.cfg.TienRealms) {
		return p.cfg.TienRealms[i]
	}
	return ""
}

// ---- Tiên giới (spec §3.4) -------------------------------------------------
//
// Tiên realms sit above Đại Thừa and have no tiers, so they occupy only the first level of
// their 10-level block: Chân Tiên = level 90, Kim Tiên = 100 … That keeps Realm(L) = L/10
// working everywhere; EXP never reaches them because GainExp caps at MaxExp.

// IsTien reports whether realm index r is a Tiên realm.
func (p *Progression) IsTien(r int) bool { return r >= len(p.cfg.Realms) }

// TienIndex is the 0-based position inside tien_realms.
func (p *Progression) TienIndex(r int) int { return r - len(p.cfg.Realms) }

// TienCount is how many Tiên realms exist.
func (p *Progression) TienCount() int { return len(p.cfg.TienRealms) }

// TopMortalLevel is Đại Thừa tầng 10 — where sát mạng starts counting.
func (p *Progression) TopMortalLevel() int { return len(p.needed) - 1 }

// TienLevel is the level index of Tiên realm idx (0-based).
func (p *Progression) TienLevel(idx int) int {
	return (len(p.cfg.Realms) + idx) * p.cfg.TiersPerRealm
}

// TienEntry is the cumulative sát mạng needed to STAND at Tiên realm idx: 10, 30, 60, 100, 150.
func (p *Progression) TienEntry(idx int) float64 {
	sum := 0.0
	for i := 0; i <= idx && i < len(p.cfg.TienKillsPerStep); i++ {
		sum += p.cfg.TienKillsPerStep[i]
	}
	return sum
}

// TienNextAt is the cumulative sát mạng that promotes someone standing at Tiên realm idx
// (idx = -1 means Đại Thừa tầng 10, still mortal). Returns 0 when already at the top.
func (p *Progression) TienNextAt(idx int) float64 {
	if idx+1 >= len(p.cfg.TienKillsPerStep) {
		return 0
	}
	return p.TienEntry(idx + 1)
}
