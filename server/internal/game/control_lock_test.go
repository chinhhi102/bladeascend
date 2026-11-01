package game

import "testing"

// Spec §6.3: ngủ gục (Ru Ngủ) và choáng → không kỹ năng, không Thiêu Đốt.
// Ngoại lệ: Shipper — Gom Đơn "miễn mọi khống chế" nên vẫn bấm được để thoát (§17).
func TestSleepAndStunLockSkillAndBurn(t *testing.T) {
	for _, st := range []int{StSleep, StStun} {
		g := quietGame(t, nil)
		p := g.AddPlayer(SpawnOptions{Name: "A", Char: "a_hon", Swords: 3})
		p.SkillCharges = 1
		g.ApplyStatus(p, st, 2)
		if g.CanCastSkill(p) {
			t.Errorf("%s: vẫn dùng được kỹ năng", StatusNames[st])
		}
		p.Burning, p.Moving = true, true
		p.TotalExp = 1e6
		if g.Burning(p) {
			t.Errorf("%s: vẫn Thiêu Đốt (bị trừ tu vi) dù không chạy được", StatusNames[st])
		}
	}
}

func TestShipperCastsWhileControlled(t *testing.T) {
	for _, st := range []int{StSleep, StStun} {
		g := quietGame(t, nil)
		p := g.AddPlayer(SpawnOptions{Name: "S", Char: "shipper", Swords: 3})
		p.SkillCharges = 1
		g.ApplyStatus(p, st, 2)
		if !g.CanCastSkill(p) {
			t.Errorf("Shipper bị %s mà không bấm được Gom Đơn để thoát", StatusNames[st])
		}
	}
}
