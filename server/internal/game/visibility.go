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

// CanSee applies stealth and bush hiding (spec §6, §8.3). Hidden players are never sent
// to viewers who cannot see them, so a modified client cannot reveal them.
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
	if target.HasBuff(BuffStealth, g.Time) && d > g.Cfg.Buffs["stealth"].RevealR*viewer.Scale {
		return false
	}
	if g.HiddenInBush(target) && viewer.Bush != target.Bush && d > g.Cfg.Zones.BushRevealR*viewer.Scale {
		return false
	}
	return true
}

// HiddenInBush is true while the target is in a bush and not revealed by a recent hit.
func (g *Game) HiddenInBush(p *Player) bool {
	return p.Bush >= 0 && p.RevealedUntil <= g.Time
}
