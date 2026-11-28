package game

import "testing"

// Bị ĐẨY thì mất chiêu "phải chờ" — khoá HÀNH VI ĐANG DÙNG của Q58.
//
// Đây là một luật KHÔNG có trong spec: §6.3 chỉ cho mất chiêu khi "bị chém chết" (Bà Tám,
// Sư Phụ) hoặc "bị chém trúng mà có khiên Kim Cang đỡ" (Ông Ba). Luật "dịch chuyển thì mất"
// do một lượt làm trước đặt ra trong `updateChannel`, và sau bản sửa 28/11 (mặt băng thôi
// huỷ chiêu) nó là cách chống lại duy nhất còn lại của ba chiêu này.
//
// Bài test này KHÔNG phán luật nào đúng — chủ dự án đang cân ở Q58. Việc của nó là: nếu một
// lượt làm sau gỡ luật đó đi thì phải gỡ CÓ Ý THỨC, vì bài này sẽ đỏ và chỉ thẳng vào Q58.
// Một luật cân bằng chỉ sống trong một câu ghi chú là một luật sẽ biến mất.
//
// Dùng `pushAway` trực tiếp vì MỌI cú đẩy trong game đều đi qua nó (Loa Phường, Bung Cơ của
// Bé Tí, sóng ra mai Cụ Rùa, sạp Cô Ba) — khoá một chỗ là khoá cả họ.
//
// ── Hai lần tôi làm bài này sai, ghi lại vì cả hai đều là bẫy chung ──────────────────────
//
//  1. Dùng `step` (chỉ chạy `updateSkillState`): cú đẩy chỉ đặt `KnockVel`, người chỉ thật
//     sự dịch trong `movePlayer`. Bài "đạt" vì cú đẩy không bao giờ xảy ra.
//  2. Đo "chiêu có phát không" qua NẠN NHÂN (số kiếm địch tụt, địch bị đẩy). Va kiếm thường
//     cũng mài mất kiếm địch, nên bài đỏ oan ba lượt liền dù chiêu đã bị huỷ đúng.
//
// Nên bản này KHÔNG có nạn nhân nào: dấu hiệu "chiêu đã phát" đọc thẳng từ SỰ KIỆN kết thúc
// mà server phát ra. Không nạn nhân thì không va kiếm, không gì nhiễu được.
func TestPushCancelsChannel(t *testing.T) {
	cases := []struct {
		char    string
		channel string
		// huy: sự kiện báo huỷ. done: những sự kiện CHỈ xuất hiện khi chiêu phát xong.
		huy  string
		done []string
	}{
		{"ba_tam_loa", "loa_channel", "loa_huy", []string{"loa_het", "loa"}},
		{"ong_ba_vong", "ru_ngu_channel", "ru_ngu_huy", []string{"ru_ngu", "trom_kiem"}},
		{"su_phu_mi", "hut_mi_channel", "hut_mi_huy", []string{"hut_mi", "hut_mi_miss"}},
	}
	for _, c := range cases {
		t.Run(c.char, func(t *testing.T) {
			g := skillGame(t)
			p := put(g, c.char, Vec{X: 3000, Y: 3000}, 10, 20)
			pusher := put(g, "be_ti", Vec{X: 1000, Y: 1000}, 10, 20)

			giveCharge(p)
			before := p.SkillCharges
			if !g.CastSkill(p) {
				t.Fatal("không dùng được kỹ năng")
			}
			if p.SkillCharges != before-1 {
				t.Fatalf("bấm chiêu mà không trừ lượt: %d → %d", before, p.SkillCharges)
			}
			g.pushAway(pusher, p, g.BodyR(p)*4)
			g.Events = nil

			const dt = 1.0 / 30
			seen := map[string]bool{}
			for i := 0; i < int((g.skillNum(c.channel)+0.4)/dt); i++ {
				g.Tick(dt)
				for _, e := range g.Events {
					if e.Kind == "SkillEffect" {
						seen[e.S] = true
					}
				}
				g.Events = nil
			}

			if p.channel.kind != "" {
				t.Errorf("%s: bị đẩy mà chiêu vẫn đang chờ", c.char)
			}
			if !seen[c.huy] {
				t.Errorf("%s: huỷ chiêu mà không phát %q — người chơi mất lượt mà không biết vì sao "+
					"(đúng nửa thứ hai của lỗi 28/11: client chỉ bắt `ru_ngu_huy`)", c.char, c.huy)
			}
			for _, d := range c.done {
				if seen[d] {
					t.Errorf("%s: bị đẩy giữa nhịp chờ mà chiêu VẪN PHÁT (thấy %q) — nếu đây là thay đổi "+
						"CÓ Ý thì sửa bài test này và ghi vào Q58", c.char, d)
				}
			}
			if p.SkillCharges != before-1 {
				t.Errorf("%s: huỷ chiêu mà HOÀN lượt (%d) — `cancelChannel` nói rõ lượt không được hoàn",
					c.char, p.SkillCharges)
			}
		})
	}
}
