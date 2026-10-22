package game

import "math"

// ViewHalfExtents returns the half width/height of what a player can see (spec §4.1: the camera backs off to show the whole ring, + AOI margin).
func (g *Game) ViewHalfExtents(p *Player) (float64, float64) {
	pc := g.Cfg.Player
	view := math.Max(pc.CameraViewH*math.Pow(p.Scale, pc.CameraExp), pc.CameraReachMul*p.Geom.Reach)
	h := view/2 + g.Cfg.Net.AOIMargin
	return h * g.Cfg.Net.AOIAspect, h
}

// InView reports whether pos is inside viewer's area of interest.
func (g *Game) InView(viewer *Player, pos Vec) bool {
	hw, hh := g.ViewHalfExtents(viewer)
	return math.Abs(pos.X-viewer.Pos.X) <= hw && math.Abs(pos.Y-viewer.Pos.Y) <= hh
}

// CanSee applies bush hiding and gate transit (spec §8.3). Players hidden this way are never
// sent to viewers who cannot see them, so a modified client cannot reveal them.
//
// Ẩn Thân no longer hides anyone from the packet (spec §6.2 from v3.3): the target is still
// sent, flagged, and drawn at visible_alpha. Use StealthFaded for that.
func (g *Game) CanSee(viewer, target *Player) bool {
	if target.Dead && g.Time-target.DiedAt > g.Cfg.Player.DeathAnimSecs {
		return false // corpses stay visible only while the death animation plays
	}
	if viewer == target {
		return true
	}
	if target.Hidden(g.Time) || !g.InView(viewer, target.Pos) {
		return false
	}
	d := viewer.Pos.Dist(target.Pos)
	if g.HiddenInBush(target) && viewer.Bush != target.Bush && d > g.Cfg.Zones.BushRevealR*viewer.Scale {
		return false
	}
	return true
}

// StealthFaded reports whether target should be drawn faded (and nameless) for this viewer:
// they have Ẩn Thân and the viewer is farther than reveal_r × the viewer's scale. Up close
// a stealthed player shows at full strength (spec §6.2).
func (g *Game) StealthFaded(viewer, target *Player) bool {
	if viewer == target || !target.HasBuff(BuffStealth, g.Time) {
		return false
	}
	return viewer.Pos.Dist(target.Pos) > g.Cfg.Buffs["stealth"].RevealR*viewer.Scale
}

// BotSightMult shrinks a bot's detection range against a stealthed target (spec §10).
func (g *Game) BotSightMult(target *Player) float64 {
	if !target.HasBuff(BuffStealth, g.Time) {
		return 1
	}
	if m := g.Cfg.Buffs["stealth"].BotDetectMult; m > 0 {
		return m
	}
	return 1
}

// HiddenInBush is true while the target is in a bush and not revealed by a recent hit.
func (g *Game) HiddenInBush(p *Player) bool {
	return p.Bush >= 0 && p.RevealedUntil <= g.Time
}
