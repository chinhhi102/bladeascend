package game

import (
	"math"

	"kiemxoay/server/internal/config"
	"kiemxoay/server/internal/geom"
)

// Ring is a player's orbiting sword ring (spec §4.1a).
type Ring struct {
	Count int
	Angle float64 // phase of slot 0 at the current tick
	Dir   float64 // +1 | -1
	// Tilt is the current blade lean in radians. At rest it is -Dir*tilt_deg so blades trail
	// behind the spin ("ngả về phía sau so với chiều quay"); it tweens over tilt_tween_secs on a flip.
	Tilt  float64
	Crack float64 // accumulated crack from clashes with lower realms (§5.2)
	// SweepLo/Sweep: this tick the ring turned from phase SweepLo through Sweep radians
	// (always counter-clockwise from SweepLo), used for swept collisions (§4.1b).
	SweepLo, Sweep float64
}

// RingGeom is the derived ring geometry for one tick (spec §4.1a).
type RingGeom struct {
	Base     float64 // SWORD_LEN * scale * buff_long
	Inner    float64 // hilt radius of ring 0
	Width    float64
	HiltStep float64
	Growth   float64
	Compound bool
	PerRing  int
	Rings    int     // R: number of rings currently drawn
	Reach    float64 // farthest blade tip at rest tilt + half width (camera, AI, display)
	outer    float64 // farthest any blade point can be at any tilt (broad phase)
}

// ComputeRingGeom applies the §4.1a formulas; longMult is buff_long_sword (1.3 | 1.0).
func ComputeRingGeom(c config.Sword, bodyR, scale float64, n int, longMult float64) RingGeom {
	per := max(c.SwordsPerRing, 1)
	g := RingGeom{
		Base: c.Length * scale * longMult, Inner: (bodyR + c.HiltGap) * scale, Width: c.Width * scale,
		HiltStep: c.HiltStep, Growth: c.RingGrowth, Compound: c.RingGrowthMode == "compound",
		PerRing: per, Rings: max(1, (n+per-1)/per),
	}
	tilt := c.TiltDeg * math.Pi / 180
	for k := 0; k < g.Rings; k++ {
		g.Reach = math.Max(g.Reach, g.TipDist(k, tilt))
		g.outer = math.Max(g.outer, g.HiltR(k)+g.Len(k))
	}
	g.Reach += g.Width / 2
	g.outer += g.Width / 2
	return g
}

// HiltR is hilt_r(k) = inner_r + HILT_STEP * k * base.
func (g RingGeom) HiltR(k int) float64 { return g.Inner + g.HiltStep*float64(k)*g.Base }

// Len is len(k): each ring's tip reaches ring_growth * base farther than the previous one.
func (g RingGeom) Len(k int) float64 {
	if g.Compound {
		return g.Base*math.Pow(1+g.Growth, float64(k)) - g.HiltStep*float64(k)*g.Base
	}
	return g.Base * (1 + (g.Growth-g.HiltStep)*float64(k))
}

// TipDist is the distance from the centre to ring k's blade tip when blades lean by tilt.
func (g RingGeom) TipDist(k int, tilt float64) float64 {
	h, l := g.HiltR(k), g.Len(k)
	return math.Sqrt(h*h + l*l + 2*h*l*math.Cos(tilt))
}

// Outer bounds every blade point at any tilt (used for broad-phase tests).
func (g RingGeom) Outer() float64 { return g.outer }

// RingOf returns the ring index of sword i.
func (g RingGeom) RingOf(i int) int { return i / g.PerRing }

// SlotAngle is the angle of sword i relative to the ring phase: n ≤ per → evenly spread;
// otherwise 20 slots per ring, a partial ring spreads over the slots, and ring k is staggered
// by k * slot / R so the rings share out the gap between two slots (spec §4.1a).
func SlotAngle(per, n, i int) float64 {
	if n <= per {
		return 2 * math.Pi * float64(i) / float64(n)
	}
	k, j := i/per, i%per
	m := min(per, n-k*per)
	slot := math.Round(float64(j*per) / float64(m))
	step := 2 * math.Pi / float64(per)
	rings := (n + per - 1) / per
	return slot*step + float64(k)*step/float64(rings)
}

// SwordSegment returns the capsule centre-line (hilt → tip) of sword i around center.
func SwordSegment(center Vec, ring Ring, g RingGeom, i int) (Vec, Vec) {
	return swordAt(center, ring.Angle, ring, g, i)
}

// swordAt is SwordSegment for an arbitrary ring phase (sub-steps of a sweep).
func swordAt(center Vec, phase float64, ring Ring, g RingGeom, i int) (Vec, Vec) {
	k := g.RingOf(i)
	a := phase + SlotAngle(g.PerRing, ring.Count, i)
	hilt := center.Add(geom.FromAngle(a).Mul(g.HiltR(k)))
	return hilt, hilt.Add(geom.FromAngle(a + ring.Tilt).Mul(g.Len(k)))
}

// RestTilt is the blade lean for a spin direction.
func RestTilt(c config.Sword, dir float64) float64 { return -dir * c.TiltDeg * math.Pi / 180 }

// stepTilt moves the tilt toward its rest value for the current direction.
func stepTilt(c config.Sword, r *Ring, dt float64) {
	target := RestTilt(c, r.Dir)
	rate := 2 * c.TiltDeg * math.Pi / 180 / math.Max(c.TiltTweenSecs, 1e-6)
	switch d := target - r.Tilt; {
	case math.Abs(d) <= rate*dt:
		r.Tilt = target
	case d > 0:
		r.Tilt += rate * dt
	default:
		r.Tilt -= rate * dt
	}
}

// bladeOffset returns the angle, seen from the centre, between ring k's hilt direction and
// the blade point at distance d from the centre (d is clamped onto the blade).
func bladeOffset(g RingGeom, k int, tilt, d float64) float64 {
	h, l := g.HiltR(k), g.Len(k)
	tip := g.TipDist(k, tilt)
	dc := geom.Clamp(d, h, tip)
	sin, cos := math.Sincos(tilt)
	s := -h*cos + math.Sqrt(math.Max(0, dc*dc-h*h*sin*sin))
	s = geom.Clamp(s, 0, l)
	return math.Atan2(s*sin, h+s*cos)
}

// sweptHits calls hit(i) for every sword of the ring whose blade swept over circle (c, r)
// between the previous tick and this one (spec §4.1b). hit returns false to stop early.
func sweptHits(center Vec, ring Ring, g RingGeom, c Vec, r float64, hit func(i int) bool) {
	rel := c.Sub(center)
	d := rel.Len()
	rr := r + g.Width/2
	if d+rr < g.Inner || d-rr > g.Outer() {
		return
	}
	theta := rel.Angle()
	margin := math.Pi / 2
	if d > rr {
		margin = math.Asin(rr / d)
	}
	for k := 0; k < g.Rings; k++ {
		if d+rr < g.HiltR(k) || d-rr > g.TipDist(k, ring.Tilt)+g.Width/2 {
			continue
		}
		off := bladeOffset(g, k, ring.Tilt, d)
		for i := k * g.PerRing; i < min(ring.Count, (k+1)*g.PerRing); i++ {
			start := ring.SweepLo + SlotAngle(g.PerRing, ring.Count, i) + off - margin
			if normPos(theta-start) <= ring.Sweep+2*margin && !hit(i) {
				return
			}
		}
	}
}

// normPos wraps an angle into [0, 2π).
func normPos(a float64) float64 {
	a = math.Mod(a, 2*math.Pi)
	if a < 0 {
		a += 2 * math.Pi
	}
	return a
}

// denseRing is true when the inner ring's swords sweep its whole annulus every tick, so the
// annulus itself can be treated as the damage area (spec §4.1b).
func denseRing(ring Ring, g RingGeom) bool {
	m := min(ring.Count, g.PerRing)
	return float64(m)*ring.Sweep >= 2*math.Pi
}

// sweepSteps is how many sub-steps sample a sweep for shape tests (≈ every 8°).
func sweepSteps(ring Ring) int { return max(1, int(math.Ceil(ring.Sweep/(8*math.Pi/180)))) }

// sweepPhase is the ring phase at sub-step s of n.
func sweepPhase(ring Ring, s, n int) float64 {
	return ring.SweepLo + ring.Sweep*float64(s)/float64(n)
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

// ResolveClash applies the sword-break table of spec §5.2 with loss unit u = clash_unit.
func ResolveClash(c config.Sword, a, b ClashSide) ClashOutcome {
	out := ClashOutcome{CrackA: a.Crack, CrackB: b.Crack}
	u := max(c.ClashUnit, 1)
	if a.Realm == b.Realm {
		out.LostA = loseUpTo(a.Count, u, c.MinCount)
		out.LostB = loseUpTo(b.Count, u, c.MinCount)
		return out
	}
	k := ClashRatio(c, a.Realm, b.Realm)
	if a.Realm < b.Realm {
		out.LostA = loseUpTo(a.Count, u*k, c.MinCount)
		out.LostB, out.CrackB = applyCrack(b, out.LostA, k, c.MinCount)
		return out
	}
	out.LostB = loseUpTo(b.Count, u*k, c.MinCount)
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
