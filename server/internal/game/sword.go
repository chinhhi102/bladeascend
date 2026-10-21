package game

import (
	"math"

	"kiemxoay/server/internal/config"
	"kiemxoay/server/internal/geom"
)

// Ring is a player's orbiting sword ring (spec §4).
type Ring struct {
	Count int
	Angle float64 // phase of sword 0 at the current tick
	Dir   float64 // +1 | -1
	Crack float64 // accumulated crack from clashes with lower realms (§5.2)
}

// RingGeom is the derived size of a ring for one tick.
type RingGeom struct {
	Radius, Length, Width float64
}

// ComputeRingGeom applies the §4 formulas; longMult is buff_long_sword (1.3 | 1.0).
func ComputeRingGeom(c config.Sword, scale float64, n int, longMult float64) RingGeom {
	base := math.Max(c.RingMinR, float64(n)*c.Spacing/(2*math.Pi))
	return RingGeom{
		Radius: scale * base * longMult,
		Length: c.Length * scale * longMult,
		Width:  c.Width * scale,
	}
}

// SwordSegment returns the capsule centre-line of sword i for a ring around center.
func SwordSegment(center Vec, ring Ring, g RingGeom, i int) (Vec, Vec) {
	a := ring.Angle + 2*math.Pi*float64(i)/float64(ring.Count)
	d := geom.FromAngle(a)
	return center.Add(d.Mul(g.Radius - g.Length/2)), center.Add(d.Mul(g.Radius + g.Length/2))
}

// ClashRatio returns k = base + perRealm * |rA - rB| (spec §16).
func ClashRatio(c config.Sword, rA, rB int) int {
	d := rA - rB
	if d < 0 {
		d = -d
	}
	return c.ClashRatioBase + c.ClashRatioPerRealm*d
}

// ClashSide is one participant's state going into a clash.
type ClashSide struct {
	Realm int
	Count int
	Crack float64
}

// ClashOutcome is the number of swords each side loses and their new crack values.
type ClashOutcome struct {
	LostA, LostB   int
	CrackA, CrackB float64
}

// ResolveClash applies the sword-break table of spec §5.2.
func ResolveClash(c config.Sword, a, b ClashSide) ClashOutcome {
	out := ClashOutcome{CrackA: a.Crack, CrackB: b.Crack}
	if a.Realm == b.Realm {
		out.LostA = loseUpTo(a.Count, 1, c.MinCount)
		out.LostB = loseUpTo(b.Count, 1, c.MinCount)
		return out
	}
	k := ClashRatio(c, a.Realm, b.Realm)
	if a.Realm < b.Realm {
		out.LostA = loseUpTo(a.Count, k, c.MinCount)
		out.LostB, out.CrackB = applyCrack(b, out.LostA, k, c.MinCount)
		return out
	}
	out.LostB = loseUpTo(b.Count, k, c.MinCount)
	out.LostA, out.CrackA = applyCrack(a, out.LostB, k, c.MinCount)
	return out
}

func loseUpTo(count, want, minCount int) int {
	return max(0, min(want, count-minCount))
}

// applyCrack adds lost/k crack to the higher side and converts whole cracks into lost swords.
func applyCrack(high ClashSide, lowLost, k, minCount int) (int, float64) {
	crack := high.Crack + float64(lowLost)/float64(k)
	// Guard against float error such as 0.25*4 = 0.9999999.
	whole := int(math.Floor(crack + 1e-9))
	if whole <= 0 {
		return 0, crack
	}
	crack = math.Max(0, crack-float64(whole))
	return loseUpTo(high.Count, whole, minCount), crack
}
