package game

import "math"

// Kinh tế Linh khí — spec v6.24 §6.0. Mọi hệ số trong `config.exp_economy`.

// OrbValueMult: hệ số giá trị viên EXP môi trường / EXP quái của người này = orb_value_mult^r.
// Đây là nguồn DUY NHẤT được nhân theo cảnh giới (§6.0 luật 7).
func (g *Game) OrbValueMult(p *Player) float64 {
	m := g.Cfg.ExpEconomy.OrbValueMult
	if m <= 0 {
		return 1
	}
	return math.Pow(m, float64(g.Prog.Realm(p.Level)))
}

// stealShare: phần EXP người chém nhận từ nạn nhân `b` — 50 % người thật, 25 % bot (§6.0
// luật 5), để không "nuôi" cấp bằng cách săn bot cấp cao.
func (g *Game) stealShare(b *Player) float64 {
	if b.IsBot && g.Cfg.ExpEconomy.BotStealShare > 0 {
		return g.Cfg.ExpEconomy.BotStealShare
	}
	return g.Cfg.Player.AttackerShare
}
