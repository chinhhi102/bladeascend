package game

// Phạm vi kỹ năng theo đơn vị U của NGƯỜI DÙNG (GAME_SPEC §6.3, CHANGES_v6_5 mục 1–2).
//
//	U = player.pickup_r × scale_người_dùng
//
// Đây là đúng bán kính vòng hút Linh khí của họ: Luyện Khí 1 = 90 · Kim Đan ≈ 190 ·
// Đại Thừa 10 = 450 · Đạo Tổ = 540.
//
// Vì sao U chứ không phải tầm chém: vòng hút chỉ phụ thuộc cảnh giới (scale), **không đổi
// khi mất hay thêm kiếm**. Dùng tầm chém thì cùng một kỹ năng lúc vươn xa lúc hụt, tuỳ
// người chơi vừa bể mấy thanh — không ai đoán được, và không cân bằng được.
//
// Mọi hệ số nằm ở `config/game.json → skills.ranges_u`. Trong mã kỹ năng **không được có
// một con số phạm vi nào**; `TestNoHardcodedSkillRanges` canh đúng điều đó.

import "math"

// SkillU là đơn vị phạm vi của một người chơi.
func (g *Game) SkillU(p *Player) float64 {
	return g.Cfg.Player.PickupR * p.Scale
}

// skillRange đổi hệ số trong `ranges_u` thành khoảng cách thật cho người dùng `p`.
//
// Thiếu khoá thì trả 0 — tức kỹ năng không với tới ai. Chọn 0 thay vì một giá trị mặc định
// là có chủ ý: kỹ năng im lặng không ăn thì người ta đi tìm ngay, còn một giá trị mặc định
// "tàm tạm" thì sống trong game nhiều tuần mà không ai biết nó sai.
// `config.Validate` đã chặn từ lúc nạp nên đường này gần như không bao giờ chạy tới.
func (g *Game) skillRange(p *Player, key string) float64 {
	return g.Cfg.Skills.RangesU[key].Float() * g.SkillU(p)
}

// skillNum đọc một con số trong `ranges_u` **không** nhân với U: số giây, số lượng, tỉ lệ.
//
// Spec §16 gom cả ba kiểu vào một khối (`ru_ngu_r: 4` cạnh `ru_ngu_secs: 2.5`), nên phải có
// hai cửa đọc. Lẫn hai cửa này là lỗi nặng và im lặng: `ru_ngu_secs` nhân với U thành 225
// giây ngủ ở Luyện Khí.
func (g *Game) skillNum(key string) float64 {
	return g.Cfg.Skills.RangesU[key].Float()
}

// skillInt và skillBool: cùng khối, cho số nguyên (`bay_sap_count`) và cờ bật/tắt
// (`bay_sap_owner_passes`).
func (g *Game) skillInt(key string) int   { return g.Cfg.Skills.RangesU[key].Int() }
func (g *Game) skillBool(key string) bool { return g.Cfg.Skills.RangesU[key].Bool() }

// pushAway đẩy `o` ra xa `p` đúng `dist` trong `skills.push_secs` (spec §6.3).
//
// Khoảng đẩy tính theo U của NGƯỜI DÙNG, không theo kích thước người bị đẩy: người nhỏ hay
// to đều bay đúng một quãng. Bản trước lấy `KnockSecs` (0.25) của đòn bể kiếm — hai thứ
// khác nhau, và dùng chung thì chỉnh cái này là lệch cái kia.
func (g *Game) pushAway(p, o *Player, dist float64) {
	secs := math.Max(g.Cfg.Skills.PushSecs, 1e-6)
	away := o.Pos.Sub(p.Pos).Norm()
	o.KnockVel = away.Mul(dist / secs)
	o.KnockUntil = g.Time + secs
}

// SkillRangeOf: `skillRange` cho gói ngoài (bot quyết định dùng kỹ năng theo đúng tầm thật).
func (g *Game) SkillRangeOf(p *Player, key string) float64 { return g.skillRange(p, key) }
