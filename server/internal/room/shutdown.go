package room

// Tắt êm — DEPLOY_K8S_SPEC §1.3.
//
// Phòng chơi nằm trong bộ nhớ của tiến trình (§0), nên thay pod là mọi trận đang chơi kết
// thúc. Điều đó không tránh được ở giai đoạn 1; thứ tránh được là **kết thúc tử tế**:
// người chơi thấy màn Tổng kết như hết ván bình thường, điểm vào bảng xếp hạng, và một câu
// nói rõ vì sao. Không làm gì cả thì WebSocket đứt giữa chừng, người chơi mất điểm của cả
// ván và chỉ thấy game đơ — đúng cảm giác "server hỏng".

import (
	"log"

	"bladeascend/server/internal/netproto"
)

// ShutdownDrain kết thúc mọi ván đang chơi trong phòng này một cách tử tế.
//
// Chạy trên goroutine của phòng (`Do`) nên nó nằm GIỮA hai tick, không phải song song với
// tick — bắt buộc, vì nó đụng `z.G.Players` và `z.clients`, đúng hai thứ mà tick đang ghi.
//
// Trả về số người thật đã tiễn.
func (r *Room) ShutdownDrain() int {
	n := 0
	r.Do(func() {
		for _, z := range r.zones {
			for _, p := range z.G.PlayerList() {
				if p.IsBot {
					continue
				}
				c, ok := z.clients[p.ID]
				if !ok {
					continue
				}
				// Ghi điểm TRƯỚC khi gửi: `record` là nơi điểm vào bảng xếp hạng, và nếu
				// tiến trình chết giữa chừng thì thà mất cái màn hình còn hơn mất điểm.
				r.record(z, p)
				c.send(netproto.Dead{T: "dead", KillerName: "", KillerChar: "",
					PeakLevel: p.PeakLevel, PeakExp: p.PeakExp, Kills: p.Kills,
					SurviveSec: z.G.Time - p.SpawnedAt, Rank: r.rankOf(p), Score: p.Score,
					Hits: p.Hits, ScoreBy: p.ScoreBy, BurnedExp: p.BurnedExp,
					DaiThuaSecs: p.DaiThuaSecs, Top1Secs: p.Top1Secs})
				// Rồi mới nói VÌ SAO. Thứ tự này có chủ ý: client hiện màn Tổng kết khi nhận
				// `dead`, và câu "máy chủ đang cập nhật" đè lên trên nó. Gửi ngược lại thì
				// câu thông báo hiện ra rồi bị chính màn Tổng kết đẩy đi.
				c.send(netproto.Kicked{T: netproto.TKicked, Reason: netproto.KickServerUpdate})
				r.detach(z, c)
				n++
			}
		}
	})
	if n > 0 {
		log.Printf("tắt êm: đã tiễn %d người chơi khỏi phòng %s", n, r.ID)
	}
	return n
}
