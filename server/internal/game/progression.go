package game

import (
	"math"
	"sort"

	"kiemxoay/server/internal/config"
)

// Progression holds the precomputed cultivation tables (spec §3).
type Progression struct {
	cfg        config.Cultivation
	swordsPer  int
	needed     []float64 // needed[L] = EXP to go from level L to L+1
	cumulative []float64 // cumulative[L] = total EXP required to reach level L
}

func NewProgression(c config.Cultivation) *Progression {
	levels := len(c.Realms) * c.TiersPerRealm
	p := &Progression{cfg: c, swordsPer: c.SwordsPerRealm}
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

// MaxSwords returns max_swords(r) = swords_per_realm * (r + 1).
func (p *Progression) MaxSwords(r int) int { return p.swordsPer * (r + 1) }

// Scale returns scale(L) = min + (max-min) * L / (levels-1).
func (p *Progression) Scale(L int) float64 {
	last := float64(len(p.needed) - 1)
	return p.cfg.ScaleMin + (p.cfg.ScaleMax-p.cfg.ScaleMin)*float64(L)/last
}

func (p *Progression) RealmName(r int) string {
	if r < 0 || r >= len(p.cfg.Realms) {
		return ""
	}
	return p.cfg.Realms[r]
}
