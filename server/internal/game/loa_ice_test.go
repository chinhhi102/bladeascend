package game

import "testing"

// Ba chiêu "phải chờ" (Loa Phường 0,8s · Ru Ngủ 0,5s · Hút Mì 0,6s) KHÔNG được mất trắng
// chỉ vì người dùng đang đứng trên MẶT BĂNG.
//
// Spec §6.3 cho từng chiêu chỉ một cách mất: "**bị chém chết** trong lúc lấy hơi / húp".
// Không dòng nào nói đứng trên băng thì mất chiêu.
//
// Vì sao băng lại làm mất chiêu: `movePlayer` chia hai nhánh —
//
//	if g.OnIce(p.Pos) || p.HasStatus(StSlip, ...) { g.slideOnIce(p, dt) }
//	else if p.Moving { ... MoveSpeedAt(...) ... }
//
// `frozen()` (trói, choáng, ngủ, **đang hát/lấy hơi**) nằm TRONG `MoveSpeed`, nên nó chỉ
// khoá được nhánh thứ hai. Trên băng, quán tính `p.Vel` vẫn đẩy người đi suốt 0,8 giây lấy
// hơi; đi quá `BodyR` là `updateChannel` huỷ chiêu và lượt đã trừ không được hoàn.
//
// Bản đồ `bang_nguyet_ho` có vùng băng thật, nên đây là chuyện xảy ra trong trận.
//
// Dùng lại `iceGame` của `ice_test.go`: hồ băng 6 000×6 000 ngay giữa bản đồ.
func TestChannelSkillsSurviveIce(t *testing.T) {
	cases := []struct {
		char    string
		channel string
		// done: dấu hiệu chiêu đã PHÁT (không phải chỉ bắt đầu).
		done func(g *Game, caster, foe *Player) bool
	}{
		// Dấu hiệu phải là thứ CÒN LẠI sau khi chiêu xong, không phải thứ hết hạn.
		// `KnockUntil` chỉ sống `skills.push_secs` (~0,25s) nên đo nó sau khi chờ thêm 0,3s
		// là luôn thấy "không bị đẩy" — một lần nữa đo sai thứ, lần này là trong chính bài
		// đo của tôi.
		{"ba_tam_loa", "loa_channel", func(g *Game, p, foe *Player) bool {
			return foe.Pos.Dist(p.Pos) > 400 // bị hét bay ra xa (đặt cách 200 lúc đầu)
		}},
		{"ong_ba_vong", "ru_ngu_channel", func(g *Game, _, foe *Player) bool {
			return foe.HasStatus(StSleep, g.Time) // đã ngủ gục
		}},
		{"su_phu_mi", "hut_mi_channel", func(g *Game, _, foe *Player) bool {
			return foe.Ring.Count < 20 // đã bị cướp kiếm
		}},
	}
	for _, c := range cases {
		t.Run(c.char, func(t *testing.T) {
			g := iceGame(t)
			ctr := g.Cfg.Room.MapSize / 2
			p := put(g, c.char, Vec{X: ctr, Y: ctr}, 10, 20)
			foe := put(g, "cu_rua", Vec{X: ctr + 200, Y: ctr}, 10, 20)
			foe.InvulnUntil, p.InvulnUntil = 1e9, 1e9
			if !g.OnIce(p.Pos) {
				t.Fatal("chỗ thử không nằm trên băng — bài đo sai, không phải game sai")
			}
			// Đang đi rồi mới bấm: đúng kiểu dùng kỹ năng thật, và trên băng thì quán tính
			// còn nguyên khi bấm.
			p.Moving, p.InputAngle = true, 0
			const dt = 1.0 / 30
			for i := 0; i < 15; i++ {
				g.Tick(dt) // lấy đà trên băng
			}
			start := p.Pos
			giveCharge(p)
			if !g.CastSkill(p) {
				t.Fatal("không dùng được kỹ năng")
			}
			p.Moving = false // nhả phím ngay khi bấm chiêu — spec nói chiêu bắt người đứng yên
			for i := 0; i < int((g.skillNum(c.channel)+0.3)/dt); i++ {
				g.Tick(dt)
			}
			if !c.done(g, p, foe) {
				t.Errorf("%s: chiêu MẤT TRẮNG khi đứng trên băng — trượt %.1f đơn vị khỏi chỗ bấm (BodyR %.1f), "+
					"`updateChannel` huỷ vì DỊCH CHUYỂN, trong khi §6.3 chỉ cho mất chiêu khi bị chém chết",
					c.char, p.Pos.Dist(start), g.BodyR(p))
			}
		})
	}
}
