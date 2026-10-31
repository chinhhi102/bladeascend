package main

import (
	"math"
	"sort"
	"strconv"
)

// Đảo vật cản — spec §8.1b mục 2, 3, 5.
//
// Bản đồ sinh ra trước đây có mật độ chân đế 1,2–6,6%: Thiên Đình chỉ có vòng cột quanh
// quảng trường, Băng Nguyệt Hồ gần như một mặt hồ trống. Hậu quả không nằm ở chỗ "nhìn
// nhạt" mà ở chỗ **cả nhóm kỹ năng dựa vào địa hình mất tác dụng**: Shipper xuyên vật cản
// không có gì để xuyên, đẩy lùi không ép được ai vào tường, dãy sạp của Cô Ba chặn một cánh
// đồng mà đi vòng mất hai giây.
//
// Cách lấp KHÔNG phải rải thêm vật cản lẻ khắp nơi — làm thế thì chỗ nào cũng vướng một tí
// và không chỗ nào dùng được. §8.1b mục 2 nói rõ: xếp thành **đảo** giữa các đường, mỗi đảo
// tối đa ~1 200 × 1 200, và mục 3: **đảo phân tầng** — viền ngoài bậc thấp (rào, sạp), lõi
// bậc cao (gạch, đá). Người nhỏ chui vào viền nấp; người lớn chém thẳng vào lõi.
//
// Vật cản thêm vào ưu tiên **bậc ≥ 5** (chủ dự án chốt 31/10/2026): bậc thấp thì tới Luyện
// Hư là chém sạch, và bản đồ của người Đại Thừa lại trống y như cũ.

// islandTiers là bảng chọn loại cho từng lớp của đảo.
//
// Lõi lấy từ bậc 5 trở lên để địa hình còn lại ở cảnh giới cao; viền lấy bậc thấp để người
// mới vẫn phá được và vẫn có chỗ nấp.
// Hai hũ lõi, và tỉ lệ giữa chúng là con số quan trọng nhất của cả module này.
//
// `islandCoreMid` (bậc 5–7) đòi hành lang 240–290 nên nhét được dày; `islandCoreHigh`
// (bậc 9–13) đòi 340 nên thưa hơn hẳn. Nếu chỉ dùng hũ giữa thì mật độ lên đẹp Ở CẢNH GIỚI
// THẤP mà tới Đại Thừa (r=8) chém sạch — đo được 0,4–0,7% chân đế còn lại, tức bản đồ của
// người mạnh nhất vẫn là cánh đồng trống, đúng cái lỗi đang đi sửa.
//
// Nên một phần ba số đảo lấy lõi bậc ≥ 9: Đại Thừa vẫn còn chỗ nấp, chỗ bị ép vào tường,
// chỗ cho Cô Ba chặn lối.
var islandCoreMid = []string{"stone_wall", "stone_wall", "bamboo_ancient", "sect_wall", "cliff"}

var islandCoreHigh = []string{"boulder_huge", "boulder_huge", "ice_eternal", "jade_pillar", "seal_stele"}

// highCoreShare: phần đảo có lõi bậc ≥ 9 (chém được từ Chân Tiên trở lên).
const highCoreShare = 0.34

// islandCoreLow (bậc 0–3) chỉ đòi hành lang 150–180, nên xếp được dày gấp đôi hũ giữa.
// Dùng cho lớp lấp cuối: những khe còn lại sau các lớp trên đều quá hẹp cho đảo bậc 5+.
//
// §8.1b mục 3 gọi đúng tên cách này: "đảo phân tầng — viền ngoài bậc thấp (rào, sạp), lõi
// bậc cao". Ở đây là cả một đảo nhỏ bậc thấp nằm giữa các đảo lớn, cùng tác dụng: người
// mới có chỗ nấp, người cao cảnh giới chém qua không cần nghĩ.
var islandCoreLow = []string{"brick", "bamboo_wall", "boulder", "cart", "stall"}

func (g *gen) pickCoreFor(low bool) string {
	if low {
		return islandCoreLow[g.b.rng.Intn(len(islandCoreLow))]
	}
	return g.pickCore()
}

func (g *gen) pickCore() string {
	if g.b.rng.Float64() < highCoreShare {
		return islandCoreHigh[g.b.rng.Intn(len(islandCoreHigh))]
	}
	return islandCoreMid[g.b.rng.Intn(len(islandCoreMid))]
}

// islandJitter: biên xê dịch mỗi đảo khỏi ô lưới, để không lộ ra bàn cờ (§8.6 cấm "khối
// hình học lộ ra").
//
// Phải cộng vào bước lưới (xem `islandLattice`), nên xê dịch càng lớn thì lưới càng thưa và
// mật độ càng thấp. 90 là chỗ cân: đủ để nhìn không ra hàng lối, mà bước chỉ dài thêm 180.
const islandJitter = 60.0

var islandRim = []string{"fence", "stall", "cart", "bamboo_thin", "brick", "bamboo_wall", "boulder"}

// islandPlan mô tả hình dạng một đảo sẽ đặt.
type islandPlan struct {
	cx, cy float64
	r      float64 // bán kính đảo; đường kính ≤ 1200 theo §8.1b mục 2
	core   string
	rim    string
}

// islandsStep lấp thêm đảo cho tới khi mật độ chân đế chạm `target`.
//
// Vòng lặp theo mật độ ĐO ĐƯỢC chứ không theo "đặt N cái đảo": mỗi khu đã có sẵn một lượng
// vật cản riêng (Trúc Lâm 275 rặng tre, Băng Nguyệt Hồ 25 tảng đá), nên cùng một con số N
// cho ra mật độ khác nhau hẳn. Đo rồi mới quyết là cách duy nhất để năm khu cùng về đích.
// islandsStep lấp đảo cho tới khi mật độ chân đế chạm `target`.
//
// Đặt theo **LƯỚI CÓ XÊ DỊCH**, không rải ngẫu nhiên. Lý do là số học chứ không phải thẩm
// mỹ: rải ngẫu nhiên (random sequential adsorption) chỉ đạt khoảng một nửa mật độ xếp tối
// ưu, vì mỗi đảo đặt xuống lại cắt nát chỗ trống còn lại thành những mảnh không đủ cho đảo
// sau. Đo được: rải ngẫu nhiên kẹt ở 2,6–4,4% chân đế dù đã thử 30 000 lượt.
//
// Lưới cũng chính là thứ §8.1b mục 1 và mục 6 mô tả: đường chính và đường phụ chia khu
// thành ô, vật cản nằm trong ô — và "đối xứng để mọi hướng vào công bằng". Xê dịch mỗi đảo
// ±25% bước lưới để không lộ ra bàn cờ (§8.6 cấm khối hình học lộ liễu).
func (g *gen) islandsStep(target float64) {
	lo, hi := g.b.playable()
	// Quét nhiều vòng, mỗi vòng một cỡ đảo nhỏ dần: vòng đầu đặt đảo to vào chỗ rộng, các
	// vòng sau nhét đảo nhỏ vào khe còn lại. Dừng ngay khi chạm mục tiêu.
	// Cỡ đảo bắt đầu từ 250 chứ không 380: khe thật giữa các vật cản sẵn có chỉ ~645, mà
	// một đảo bán kính 380 cần 760 + hành lang. Đảo to chỉ lọt được vài chỗ rồi thôi.
	//
	// Mỗi cỡ quét HAI lượt, lượt sau lệch nửa bước lưới. Lưới thẳng để lại một khoảng trống
	// hình thoi ở giữa bốn đảo — đúng chỗ một đảo cỡ đó vừa vặn nằm vào. Không có lượt lệch
	// thì mật độ dừng ở ~5,7% dù còn chỗ nhìn thấy rõ trên ảnh.
	// Lớp đảo bậc ≥ 5: xương sống của địa hình, còn lại tới Đại Thừa.
	for _, r := range []float64{250, 200, 160, 125, 95} {
		for _, off := range []float64{0, 0.5} {
			if g.b.footprintDensity() >= target {
				return
			}
			g.islandLattice(lo, hi, r, target, off, false)
		}
	}
	// Lớp lấp cuối, lõi bậc thấp: hành lang chỉ 150–180 nên lọt được vào khe mà đảo bậc 5+
	// không vào nổi. Không có lớp này thì mật độ dừng ở ~6–9% dù ảnh còn thấy rõ chỗ trống.
	for _, r := range []float64{110, 85, 65} {
		for _, off := range []float64{0, 0.5, 0.25} {
			if g.b.footprintDensity() >= target {
				return
			}
			g.islandLattice(lo, hi, r, target, off, true)
		}
	}
}

// islandLattice rải một lớp đảo bán kính `r` theo lưới.
func (g *gen) islandLattice(lo, hi, r, target, off float64, low bool) {
	c := g.size / 2
	half := (hi - lo) / 2
	// Bước lưới = đường kính đảo + hành lang của bậc cao nhất có thể. Cộng hành lang vào
	// bước chứ không trông chờ `free` từ chối: trông chờ thì mỗi ô thử rồi trượt, còn cộng
	// vào bước thì ô nào cũng dùng được.
	// Bước lưới = đường kính đảo + hành lang bậc cao nhất + chỗ cho XÊ DỊCH.
	//
	// Cộng xê dịch vào bước, không chỉ cộng hành lang: hai đảo cạnh nhau cùng lệch về phía
	// nhau thì khoảng cách thật ngắn đi gấp đôi biên xê dịch, `free` từ chối, và cả hàng ô
	// trượt sạch. Lần đầu tôi xê dịch ±25% bước mà quên cộng vào, mật độ tụt từ 2,6% xuống
	// 2,1% — thêm đảo mà ít đi.
	coreTier := 5
	if low {
		coreTier = 2
	}
	pitch := 2*r + g.corridorFor(coreTier) + 2*islandJitter
	n := int((hi - lo) / pitch)
	if n < 1 {
		return
	}
	// Duyệt ô theo thứ tự XA TÂM TRƯỚC: §8.1b mục 5 muốn "trống ở giữa, dày hơn ở rìa", và
	// cách chắc chắn nhất là lấp rìa trước rồi mới vào trong.
	type cell struct {
		x, y, d float64
	}
	cells := make([]cell, 0, n*n)
	for iy := 0; iy < n; iy++ {
		for ix := 0; ix < n; ix++ {
			x := lo + (float64(ix)+0.5+off)*pitch
			y := lo + (float64(iy)+0.5+off)*pitch
			cells = append(cells, cell{x, y, math.Hypot(x-c, y-c)})
		}
	}
	sort.Slice(cells, func(a, b int) bool { return cells[a].d > cells[b].d })
	for _, cl := range cells {
		if g.b.footprintDensity() >= target {
			return
		}
		// Vùng giữa khu để trống hơn (§8.1b mục 5): bỏ qua một phần ô gần tâm. 0.55 ở tâm
		// chứ không 0.3 — mục tiêu mật độ tính trên CẢ khu, bỏ quá tay ở tâm thì rìa phải
		// bù lại tới mức chật.
		if g.b.rng.Float64() > 0.8+0.2*math.Min(cl.d/math.Max(half, 1), 1) {
			continue
		}
		p := islandPlan{
			cx:   cl.x + g.b.rnd(-islandJitter, islandJitter),
			cy:   cl.y + g.b.rnd(-islandJitter, islandJitter),
			r:    r * g.b.rnd(0.8, 1.0),
			core: g.pickCoreFor(low),
			rim:  islandRim[g.b.rng.Intn(len(islandRim))],
		}
		g.placeIsland(p)
	}
}

// placeIsland đặt một đảo phân tầng: lõi bậc cao ở giữa, viền bậc thấp quanh ngoài.
//
// Hành lang tối thiểu (§8.1b, bảng `obstacles.corridor_min`) áp giữa HAI ĐẢO, không giữa
// từng khối trong cùng một đảo. Đây là điểm quyết định cả bài toán mật độ:
//
// Ép corridor giữa mọi khối thì một khối 400×400 bậc 9 thật ra chiếm 1 080 × 1 080 — gấp
// 7,3 lần diện tích của nó, và trần mật độ là ~3% chân đế. Cho các khối trong một đảo chạm
// nhau thì đảo thành một CỤM đặc, corridor chỉ phải giữ ở mép ngoài.
//
// Đó cũng là điều §8.1b mục 2 nói: "vật cản xếp thành đảo giữa các đường" — đảo là một
// khối, không phải một đám khối cách đều nhau.
func (g *gen) placeIsland(p islandPlan) {
	// Hành lang lấy theo bậc CAO NHẤT của đảo: người không chém được lõi phải đi vòng, và
	// khe họ đi là khe giữa đảo này với đảo bên cạnh.
	gap := g.corridorFor(max(g.tierOf(p.core), g.tierOf(p.rim)))
	// Kiểm chỗ cho CẢ ĐẢO một lần, với hành lang đầy đủ. Đặt từng khối rồi mới biết hụt thì
	// đảo ra nham nhở — nửa cái lõi không có viền.
	bbox := rect{p.cx - p.r, p.cy - p.r, p.cx + p.r, p.cy + p.r}
	if !g.b.freeFor(bbox, max(g.tierOf(p.core), g.tierOf(p.rim)), g.corridorFor) {
		return
	}
	// Lát hộp bao bằng lưới k×k **KHÔNG CHỒNG MÉP**: ô giữa là lõi bậc cao, ô ngoài là viền
	// bậc thấp (§8.1b mục 3 "đảo phân tầng").
	//
	// Không chồng mép là điều kiện bắt buộc, không phải thẩm mỹ. Phép đẩy ra khỏi vật cản
	// (`resolveBlocking`) lặp 3 lượt, mỗi lượt đẩy lần lượt ra khỏi từng khối. Với các khối
	// RỜI nhau thì nó hội tụ và client tính ra đúng kết quả server; với các khối CHỒNG nhau
	// thì kết quả phụ thuộc thứ tự và số lượt, mà server truy vấn theo lưới không gian quanh
	// vị trí ĐANG DI CHUYỂN còn client duyệt cả mảng — hai bên ra hai danh sách khác nhau.
	// Bản đầu cho chồng mép và bài đối chiếu 1 000 bước lệch ngay **18,4 đơn vị**.
	k := 2 + g.b.rng.Intn(2) // 2×2 hoặc 3×3
	cell := 2 * p.r / float64(k)
	for iy := 0; iy < k; iy++ {
		for ix := 0; ix < k; ix++ {
			// GÓC là chỗ duy nhất được bỏ trống và cũng là chỗ duy nhất dùng bậc thấp.
			//
			// Bản đầu coi cả vành ngoài là "viền": với lưới 3×3 thì 8/9 ô thành bậc thấp,
			// nên tới Đại Thừa cả đảo biến mất và mật độ ở r=8 tụt còn 0,6%. Góc thì chỉ
			// 4/9, lõi vẫn là phần lớn đảo — đúng ý §8.1b mục 3 "viền ngoài bậc thấp, LÕI
			// bậc cao".
			corner := (ix == 0 || ix == k-1) && (iy == 0 || iy == k-1)
			if corner && g.b.rng.Float64() < 0.3 {
				continue // khe để người nhỏ chui vào nấp, và để đảo không ra hình vuông
			}
			kind := p.core
			if k >= 3 && corner {
				kind = p.rim
			}
			x := p.cx - p.r + float64(ix)*cell
			y := p.cy - p.r + float64(iy)*cell
			g.b.obstacleForce(kind, x, y, cell, cell)
		}
	}
	_ = gap
}

// tierOf tra bậc của một loại vật cản từ config — không viết số vào mã.
func (g *gen) tierOf(kind string) int { return g.b.tierOf(kind) }

// corridorFor là hành lang tối thiểu của bậc `t` (bảng §8.2 `obstacles.corridor_min`).
//
// Bảng chỉ ghi vài bậc mốc (1, 2, 3, 5, 7, 9, 11, 13), nên bậc không có trong bảng thì lấy
// mốc GẦN NHẤT KHÔNG VƯỢT QUÁ nó — tra hụt rồi rơi về 0 là đặt hai khối bậc 9 sát nhau.
func (g *gen) corridorFor(t int) float64 {
	best, found := 0.0, false
	keys := make([]int, 0, len(g.obs.CorridorMin))
	for k := range g.obs.CorridorMin {
		n, err := strconv.Atoi(k)
		if err != nil {
			continue
		}
		keys = append(keys, n)
	}
	sort.Ints(keys)
	for _, k := range keys {
		if k <= t {
			best, found = g.obs.CorridorMin[strconv.Itoa(k)], true
		}
	}
	if !found {
		return 150 // bậc 0: ai cũng chém được, chỉ cần khe lọt người nhỏ nhất
	}
	return best
}
