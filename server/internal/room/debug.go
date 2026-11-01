package room

import (
	"log"
	"sort"

	"bladeascend/server/internal/game"
	"bladeascend/server/internal/netproto"
)

// Debug applies a test shortcut (M1 requirement); only reachable when the server runs with -debug.
func (r *Room) Debug(c *Client, d netproto.Debug) {
	r.Do(func() {
		z := r.zoneOf(c.PlayerID)
		if z == nil {
			log.Printf("dbg %q bỏ qua: người chơi %d không còn trong phòng", d.Cmd, c.PlayerID)
			return
		}
		p, ok := z.G.Players[c.PlayerID]
		if !ok || p.Dead {
			// Im lặng bỏ qua ở đây từng làm bài test 10 tab ra bảng số sai suốt hai lượt:
			// lệnh lên Đạo Tổ gửi đi lúc nhân vật đang nằm chờ hồi sinh thì rơi vào hư vô.
			log.Printf("dbg %q bỏ qua: người chơi %d %s", d.Cmd, c.PlayerID,
				map[bool]string{true: "đã chết", false: "không còn trong phòng"}[ok])
			return
		}
		switch d.Cmd {
		case "exp":
			z.G.GainExpFrom(p, float64(d.N), false, game.ExpDebug)
		case "realm":
			// Jump to the start of the next realm. Từ Đại Thừa tầng 10 trở lên không hấp thụ
			// Linh khí nữa (§3.4) nên phải cấp thẳng Tiên khí, không thì phím test tắc ở đó
			// và không bao giờ thử được Tiên giới.
			if z.G.BanksTienKhi(p) {
				// Cấp đủ Tiên khí để vượt TRẦN cảnh giới hiện tại — vượt trần mới lên cảnh
				// giới mới (spec v5.7). Chưa có Tiên khí nào thì một điểm là vào Chân Tiên.
				idx := z.G.TienRealmOf(p)
				if ceil := z.G.Prog.TienCeil(idx); idx < 0 {
					z.G.GainTienKhi(p, 1)
				} else if idx+1 < z.G.Prog.TienCount() {
					z.G.GainTienKhi(p, ceil+1-p.TienKhi)
				}
				return
			}
			next := (z.G.Prog.Realm(p.Level) + 1) * z.G.Prog.Tiers()
			z.G.GainExp(p, z.G.Prog.Threshold(next)-p.TotalExp, false)
		case "max":
			// Lên thẳng Đạo Tổ với đủ kiếm. Chỉ dùng cho công cụ đo (tools/browsertest) —
			// bấm phím "realm" 14 lần rồi "sword" 105 lần thì vừa chậm vừa hay lệch.
			top := z.G.Prog.TienLevel(len(r.cfg.Cultivation.TienRealms) - 1)
			z.G.ForceState(p, top, z.G.Prog.MaxSwords(z.G.Prog.Realm(top)), p.Pos)
			if d.N == 0 {
				p.InvulnUntil = 0 // mặc định: đánh nhau thật
			}
			// n = 1 → giữ bất tử. Công cụ đo cần giữ nguyên cảnh NẶNG NHẤT suốt pha đo;
			// để chúng chém chết nhau thì sau vài giây cả phòng về Luyện Khí 1 kiếm và số
			// đo thành vô nghĩa (đã dính đúng lỗi này một lần).
		case "tien":
			// Nạp Tiên khí để thử luật Tiên giới (đổi dư sang Uy Danh, cướp đoạt §3.4).
			z.G.GainTienKhi(p, float64(max(1, d.N)))
		case "sword":
			z.G.AddSwords(p, max(1, d.N))
		case "skill":
			// Hồi đầy lượt ngay, khỏi chờ hồi chiêu khi thử kỹ năng (v6.25: không còn Bí Kíp).
			z.G.FillSkillCharges(p)
		case "die":
			z.G.DebugKill(p)
		case "buff":
			if d.N >= 0 && d.N < game.BuffCount {
				z.G.ApplyBuff(p, game.BuffNames[d.N])
			}
		case "summon":
			r.debugSummon(z, p, float64(d.N))
		case "bot":
			// Thả một bot ngay cạnh mình để thử đánh nhau; d.N = cảnh giới (0 = như mình).
			realm := d.N
			if realm <= 0 {
				realm = z.G.Prog.Realm(p.Level)
			}
			realm = min(realm, len(r.cfg.Cultivation.Realms)-1)
			b := r.spawnBotIn(z, realm*z.G.Prog.Tiers(), z.G.Prog.MaxSwords(realm)/4+1)
			b.Pos = p.Pos.Add(z.G.Vec(z.G.Geom(p).Reach+120, 0))
			b.InvulnUntil = 0
		default:
		}
	})
}

// debugSummon kéo mọi NGƯỜI THẬT khác trong phòng (kể cả ở khu khác) về cạnh `p` — công cụ
// chụp ảnh kỹ năng (`tools/browsertest/skillshots.mjs`, spec v6.36 §6.5). Người mới vào
// phòng được thả vào khu ÍT người thật nhất, nên ba trình duyệt thử luôn nằm ba khu khác nhau.
//
// Xếp theo TÊN (id không so được giữa các khu): người đầu đứng bên phải cách 3,5 U — trong
// tầm Ru Ngủ (4 U), Hút Mì (5 U), Loa (6 U), Trân Châu (8 U), mà hai vòng kiếm Luyện Khí
// không chạm nhau (đứng 1,8 U thì chém chết nhau ngay trong ảnh). Người sau đứng phía trên.
// dist > 0 = khoảng cách tự chọn. Chạy trong `r.Do` nên các khu đang không tick.
func (r *Room) debugSummon(z *Zone, p *game.Player, dist float64) {
	if dist <= 0 {
		dist = 3.5 * z.G.SkillU(p)
	}
	type other struct {
		o  *game.Player
		at *Zone
	}
	var list []other
	for _, oz := range r.zones {
		for _, o := range oz.G.PlayerList() {
			if o.ID != p.ID && !o.IsBot && !o.Dead {
				list = append(list, other{o, oz})
			}
		}
	}
	sort.Slice(list, func(i, j int) bool { return list[i].o.Name < list[j].o.Name })
	for k, it := range list {
		off := z.G.Vec(dist, 0)
		if k > 0 {
			off = z.G.Vec(dist*0.5, -dist*0.8*float64(k))
		}
		to := p.Pos.Add(off)
		it.o.Vel = game.Vec{}
		if it.at == z {
			it.o.Pos = to
			continue
		}
		r.queueTransfer(transfer{playerID: it.o.ID, from: it.at, to: z, at: to})
	}
}
