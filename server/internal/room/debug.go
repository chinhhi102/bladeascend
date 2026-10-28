package room

import (
	"kiemxoay/server/internal/game"
	"kiemxoay/server/internal/netproto"
)

// Debug applies a test shortcut (M1 requirement); only reachable when the server runs with -debug.
func (r *Room) Debug(c *Client, d netproto.Debug) {
	r.Do(func() {
		p, ok := r.G.Players[c.PlayerID]
		if !ok || p.Dead {
			return
		}
		switch d.Cmd {
		case "exp":
			r.G.GainExp(p, float64(d.N), false)
		case "realm":
			// Jump to the start of the next realm. Từ Đại Thừa tầng 10 trở lên không hấp thụ
			// Linh khí nữa (§3.4) nên phải cấp thẳng Tiên khí, không thì phím test tắc ở đó
			// và không bao giờ thử được Tiên giới.
			if r.G.BanksTienKhi(p) {
				need := r.G.Prog.TienNextAt(r.G.TienRealmOf(p))
				if need > 0 {
					r.G.GainTienKhi(p, need-p.TienKhi)
				}
				return
			}
			next := (r.G.Prog.Realm(p.Level) + 1) * r.G.Prog.Tiers()
			r.G.GainExp(p, r.G.Prog.Threshold(next)-p.TotalExp, false)
		case "max":
			// Lên thẳng Đạo Tổ với đủ kiếm. Chỉ dùng cho công cụ đo (tools/browsertest) —
			// bấm phím "realm" 14 lần rồi "sword" 105 lần thì vừa chậm vừa hay lệch.
			top := r.G.Prog.TienLevel(len(r.cfg.Cultivation.TienRealms) - 1)
			r.G.ForceState(p, top, r.G.Prog.MaxSwords(r.G.Prog.Realm(top)), p.Pos)
			if d.N == 0 {
				p.InvulnUntil = 0 // mặc định: đánh nhau thật
			}
			// n = 1 → giữ bất tử. Công cụ đo cần giữ nguyên cảnh NẶNG NHẤT suốt pha đo;
			// để chúng chém chết nhau thì sau vài giây cả phòng về Luyện Khí 1 kiếm và số
			// đo thành vô nghĩa (đã dính đúng lỗi này một lần).
		case "sword":
			r.G.AddSwords(p, max(1, d.N))
		case "die":
			r.G.DebugKill(p)
		case "buff":
			if d.N >= 0 && d.N < game.BuffCount {
				r.G.ApplyBuff(p, game.BuffNames[d.N])
			}
		case "bot":
			// Thả một bot ngay cạnh mình để thử đánh nhau; d.N = cảnh giới (0 = như mình).
			realm := d.N
			if realm <= 0 {
				realm = r.G.Prog.Realm(p.Level)
			}
			realm = min(realm, len(r.cfg.Cultivation.Realms)-1)
			b := r.spawnBotAt(realm*r.G.Prog.Tiers(), r.G.Prog.MaxSwords(realm)/4+1)
			b.Pos = p.Pos.Add(r.G.Vec(r.G.Geom(p).Reach+120, 0))
			b.InvulnUntil = 0
		default:
		}
	})
}
