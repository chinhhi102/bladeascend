package room

import "slices"

// heldBoard là thứ tự top 10 phòng ĐÃ QUA chống nhấp nháy — nguồn DUY NHẤT cho cả thứ tự
// bảng xếp hạng (§11.5) lẫn danh hiệu Top 1–10 (§11.3, giữ 3 s).
//
// Trước 1/11 hai thứ này giữ nhịp RIÊNG: bảng xếp theo Uy Danh tức thời, danh hiệu thì mỗi
// người tự đếm 3 s của mình. Kết quả chủ dự án chụp được: thẻ hạng 2 mang danh hiệu hạng 3.
// Đếm riêng từng người còn sinh ra hai người cùng đeo một danh hiệu khi các lần đổi hạng
// lệch nhịp nhau. Giữ CẢ THỨ TỰ một lần thì hai thứ không thể lệch — chúng đọc cùng một mảng.
//
// MỌI con số hiện trên bảng phải từ CÙNG bản chụp với thứ tự — Uy Danh và cả CẢNH GIỚI.
// Duyệt 1/11 bắt được kiểu lỗi này với điểm: dòng 8 = 75 đứng trên dòng 9 = 95, vì thứ tự
// giữ 3 s còn điểm thì cập nhật ngay. Từ v6.43 cảnh giới là khoá sắp xếp CHÍNH (§11.2), nên
// nó dính y hệt: ai vừa đột phá sẽ hiện cảnh giới mới ngay mà vẫn nằm ở hạng cũ, bảng đọc
// như sai suốt 3–6 giây.
//
// Bản chụp chỉ cập nhật lúc thứ tự thật TRÙNG thứ tự đang hiện, hoặc lúc nhận thứ tự mới —
// cả hai lúc đó dãy số đều đúng chiều với `RankLess`, nên bảng không bao giờ tự mâu thuẫn.
type heldBoard struct {
	held         []uint32
	scores       map[uint32]float64
	levels       map[uint32]int
	pending      []uint32
	pendingSince float64
	// divergedSince: lúc thứ tự thật bắt đầu khác thứ tự đang hiện; 0 = đang khớp.
	divergedSince float64
}

// step cập nhật thứ tự đang hiện từ thứ tự thật `live` (top N người còn sống).
//
// Thứ tự mới chỉ được nhận khi đứng yên đủ `hold` giây. Nhưng phòng đông thì top 10 có thể
// đổi liên tục chưa tới 3 s một lần — chờ "đứng yên" thì bảng đóng băng mãi. Nên lệch quá
// `2×hold` là nhận luôn thứ tự hiện tại: bảng trễ tối đa 6 s, không bao giờ kẹt.
//
// `present`: mọi người đang được xếp hạng. Ai rời phòng hoặc chết thì rút khỏi bảng NGAY
// (không chờ 3 s) — giữ một cái tên đã đi khỏi phòng là sai, không phải "chống nhấp nháy".
func (b *heldBoard) step(live []uint32, score map[uint32]float64, level map[uint32]int, present map[uint32]bool, now, hold float64) {
	b.held = slices.DeleteFunc(b.held, func(id uint32) bool { return !present[id] })
	b.fill(live, score, level)
	if slices.Equal(live, b.held) {
		b.pending, b.divergedSince = nil, 0
		b.snap(score, level)
		return
	}
	if b.divergedSince == 0 {
		b.divergedSince = now
	}
	if !slices.Equal(live, b.pending) {
		b.pending, b.pendingSince = slices.Clone(live), now
	}
	if len(b.held) == 0 || now-b.pendingSince >= hold || now-b.divergedSince >= 2*hold {
		b.held, b.pending, b.divergedSince = slices.Clone(live), nil, 0
		b.snap(score, level)
	}
}

// fill bù cho bảng đủ người từ thứ tự THẬT, giữ nguyên chỗ của những ai đang đứng trên bảng.
//
// Đây là nửa server của lỗi 07/11 "bảng tụt còn 3 dòng rồi bung lại 10" (v6.47 §11.5). Hai
// việc ở trên đúng riêng lẻ nhưng sai khi đi cùng nhau: người chết/rời bị rút NGAY, còn thứ
// tự mới thì phải chờ đứng yên 3 s. Khoảng giữa hai mốc đó `held` ngắn hơn 10 và gói đi ra
// đúng y số dòng ngắn ấy — đo được 5 dòng trong phòng còn 13 người sống.
//
// Bù vào CUỐI theo thứ tự thật: ai vừa lọt vào top hiện ở dòng dưới rồi mới trèo lên khi
// thứ tự mới được nhận. Đổi chỗ của những người cũ vẫn phải chờ đủ 3 s như trước.
func (b *heldBoard) fill(live []uint32, score map[uint32]float64, level map[uint32]int) {
	if b.scores == nil {
		b.scores, b.levels = map[uint32]float64{}, map[uint32]int{}
	}
	for _, id := range live {
		if len(b.held) >= len(live) {
			return
		}
		if !slices.Contains(b.held, id) {
			b.held = append(b.held, id)
			b.scores[id], b.levels[id] = score[id], level[id]
		}
	}
}

// snap chụp điểm VÀ cấp của những người đang hiện trên bảng.
func (b *heldBoard) snap(score map[uint32]float64, level map[uint32]int) {
	b.scores = make(map[uint32]float64, len(b.held))
	b.levels = make(map[uint32]int, len(b.held))
	for _, id := range b.held {
		b.scores[id] = score[id]
		b.levels[id] = level[id]
	}
}

// score: Uy Danh hiển thị trên bảng của `id` — từ bản chụp, không phải điểm tức thời.
// `ok` = false khi `id` không có trong bản chụp (người vừa được bù vào cho đủ 10 dòng).
func (b *heldBoard) score(id uint32) (float64, bool) {
	v, ok := b.scores[id]
	return v, ok
}

// level: cấp hiển thị trên bảng của `id` — từ bản chụp, cùng nhịp với thứ tự. `ok` = false
// khi `id` không có trong bảng, người gọi tự quyết lấy cấp tức thời hay bỏ qua.
func (b *heldBoard) level(id uint32) (int, bool) {
	l, ok := b.levels[id]
	return l, ok
}

// rank trả hạng đang hiện (1-based) của `id`; 0 nếu không có trong bảng.
func (b *heldBoard) rank(id uint32) int {
	return slices.Index(b.held, id) + 1
}
