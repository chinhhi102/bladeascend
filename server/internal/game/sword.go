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
	Base     float64 // SWORD_LEN * scale
	Inner    float64 // hilt radius of ring 0
	Width    float64
	HiltStep float64
	Growth   float64
	Compound bool
	// Step: vòng thứ k (đếm từ 1) chứa Step × k kiếm — vòng 1 là 10, vòng 2 là 20… (§4.1a).
	Step  int
	Rings int     // R: number of rings currently drawn
	Reach    float64 // farthest blade tip at rest tilt + half width (camera, AI, display)
	outer    float64 // farthest any blade point can be at any tilt (broad phase)
}

// ComputeRingGeom applies the §4.1a formulas. Spec v4.3 bỏ Trường Kiếm nên không còn hệ số
// kéo dài: tầm chém chỉ tăng khi mở vòng mới hoặc lên cảnh giới.
func ComputeRingGeom(c config.Sword, bodyR, scale float64, n int) RingGeom {
	step := max(c.RingStep, 1)
	g := RingGeom{
		Base: c.Length * scale, Inner: (bodyR + c.HiltGap) * scale, Width: c.Width * scale,
		HiltStep: c.HiltStep, Growth: c.RingGrowth, Compound: c.RingGrowthMode == "compound",
		Step: step,
	}
	g.Rings = g.RingsFor(n)
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

// RingCap is how many swords ring k holds: Step × (k+1), so rings get roomier outwards
// (vòng 1 = 10 kiếm, vòng 2 = 20… spec §4.1a).
func (g RingGeom) RingCap(k int) int { return g.Step * (k + 1) }

// RingStart is the index of ring k's first sword = Step × k(k+1)/2.
func (g RingGeom) RingStart(k int) int { return g.Step * k * (k + 1) / 2 }

// RingsFor is how many rings n swords need (at least one, even with no swords).
func (g RingGeom) RingsFor(n int) int {
	r := 1
	for g.RingStart(r) < n {
		r++
	}
	return r
}

// RingOf returns the ring index of sword i, inverting RingStart.
func (g RingGeom) RingOf(i int) int {
	// k = floor((sqrt(1 + 8i/Step) - 1) / 2), then nudged in case of float rounding.
	k := int((math.Sqrt(1+8*float64(i)/float64(g.Step)) - 1) / 2)
	for k > 0 && g.RingStart(k) > i {
		k--
	}
	for g.RingStart(k+1) <= i {
		k++
	}
	return k
}

// SlotAngle is the angle of sword i relative to the ring phase. Each ring has RingCap(k)
// evenly spaced slots; a ring that is not full spreads its swords over those slots, and ring
// k is staggered by a fraction of one slot so the rings share out the gaps (spec §4.1a).
func (g RingGeom) SlotAngle(n, i int) float64 {
	k := g.RingOf(i)
	cap, start := g.RingCap(k), g.RingStart(k)
	j := i - start
	m := min(cap, n-start)
	if m <= 0 {
		return 0
	}
	slot := math.Round(float64(j*cap) / float64(m))
	step := 2 * math.Pi / float64(cap)
	return slot*step + float64(k)*step/float64(max(g.Rings, 1))
}

// SwordSegment returns the capsule centre-line (hilt → tip) of sword i around center.
func SwordSegment(center Vec, ring Ring, g RingGeom, i int) (Vec, Vec) {
	return swordAt(center, ring.Angle, ring, g, i)
}

// swordAt is SwordSegment for an arbitrary ring phase (sub-steps of a sweep).
func swordAt(center Vec, phase float64, ring Ring, g RingGeom, i int) (Vec, Vec) {
	k := g.RingOf(i)
	a := phase + g.SlotAngle(ring.Count, i)
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
		for i := g.RingStart(k); i < min(ring.Count, g.RingStart(k+1)); i++ {
			start := ring.SweepLo + g.SlotAngle(ring.Count, i) + off - margin
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
	m := min(ring.Count, g.RingCap(0))
	return float64(m)*ring.Sweep >= 2*math.Pi
}

// sweepSteps is how many sub-steps sample a sweep for shape tests (≈ every 8°).
func sweepSteps(ring Ring) int { return max(1, int(math.Ceil(ring.Sweep/(8*math.Pi/180)))) }

// sweepPhase is the ring phase at sub-step s of n.
func sweepPhase(ring Ring, s, n int) float64 {
	return ring.SweepLo + ring.Sweep*float64(s)/float64(n)
}

// ClashSide is one participant's state going into a clash.
type ClashSide struct {
	Realm int
	Count int
	Crack float64
	// DurMult: hệ số độ bền từ buff (Cương Kiếm = buffs.hard.durability_mult, spec v4.4
	// §5.2). 0 hoặc 1 đều nghĩa là không buff.
	DurMult float64
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
	// Spec v4.4 §5.2: so theo ĐỘ BỀN D chứ không so cảnh giới nữa — Cương Kiếm làm kiếm
	// cứng hơn nên Luyện Khí có buff vẫn có thể cứng hơn Trúc Cơ không buff.
	da, db := SideDurability(c, a), SideDurability(c, b)
	if da == db {
		out.LostA = loseUpTo(a.Count, u, c.MinCount)
		out.LostB = loseUpTo(b.Count, u, c.MinCount)
		return out
	}
	// Bên mềm hơn luôn mất đúng u kiếm; bên cứng hơn chỉ nứt `lost / k`.
	if da < db {
		k := db / da
		out.LostA = loseUpTo(a.Count, u, c.MinCount)
		out.LostB, out.CrackB = applyCrack(b, out.LostA, k, c.MinCount)
		return out
	}
	k := da / db
	out.LostB = loseUpTo(b.Count, u, c.MinCount)
	out.LostA, out.CrackA = applyCrack(a, out.LostB, k, c.MinCount)
	return out
}

// SideDurability is D = độ bền theo cảnh giới × hệ số Cương Kiếm (spec v4.4 §5.2).
func SideDurability(c config.Sword, s ClashSide) float64 {
	d := realmDurability(c, s.Realm)
	if s.DurMult > 1 {
		d *= s.DurMult
	}
	return d
}

// realmDurability là độ bền tuyệt đối của kiếm ở cảnh giới r.
func realmDurability(c config.Sword, r int) float64 {
	step := c.DurabilityPerRealm
	if step <= 0 {
		step = 1
	}
	if c.DurabilityMode == "linear" {
		return 1 + (step-1)*float64(r)
	}
	return math.Pow(step, float64(r))
}

// Durability is k = độ bền kiếm bên cao chia cho bên thấp (spec v4.1 §5.2).
// "compound": 1.2^d. "linear": 1 + 0.2 × d.
func Durability(c config.Sword, rA, rB int) float64 {
	d := rA - rB
	if d < 0 {
		d = -d
	}
	step := c.DurabilityPerRealm
	if step <= 0 {
		step = 1
	}
	if c.DurabilityMode == "linear" {
		return 1 + (step-1)*float64(d)
	}
	return math.Pow(step, float64(d))
}

func loseUpTo(count, want, minCount int) int {
	return max(0, min(want, count-minCount))
}

// applyCrack adds lost/k crack to the higher side and converts whole cracks into lost swords.
func applyCrack(high ClashSide, lowLost int, k float64, minCount int) (int, float64) {
	crack := high.Crack + float64(lowLost)/k
	// Guard against float error such as 0.25*4 = 0.9999999.
	whole := int(math.Floor(crack + 1e-9))
	if whole <= 0 {
		return 0, crack
	}
	crack = math.Max(0, crack-float64(whole))
	return loseUpTo(high.Count, whole, minCount), crack
}
