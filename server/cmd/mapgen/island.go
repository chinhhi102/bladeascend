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
// islandsStep lấp vật cản cho tới khi mật độ HÌNH VẼ chạm `target`.
//
// Đặt theo CỤM LỚN CÁCH XA NHAU, không rải đều. Đây là bài học đo được của lượt 1: rải đảo
// nhỏ khắp khu cho đúng mật độ 18–28% thì **khoảng thở = 0%** — không có một điểm nào trên
// cả năm khu cách mọi vật cản quá 400 đơn vị. Mật độ đúng mà bản đồ vẫn hỏng: chỗ nào cũng
// vướng một tí, không chỗ nào rượt đuổi hay đánh nhau được.
//
// Số học của cách làm mới: cụm rộng `L`, cách nhau `S`. Phần "thở" là phần nằm ngoài cụm đã
// nở thêm 400 mỗi bên, tức `1 − (L+800)²/S²`. Phủ `L²/S² = 20%`. Hai phương trình ấy cho
//
//	(L+800)/L ≤ √(0,70/0,20) = 1,87  →  L ≥ 920
//
// Nghĩa là **cụm phải rộng ít nhất ~920 đơn vị**. Rải đảo 500 thì dù xếp thế nào cũng không
// bao giờ có 30% khoảng thở — biên 400 quanh mỗi đảo nuốt hết chỗ trống. Cụm 1 200 cách
// nhau 2 700 cho ra thở 44%, phủ 20%.
func (g *gen) islandsStep(target float64) {
	lo, hi := g.b.playable()
	// Nhiều bước lưới nhỏ dần, dừng NGAY khi chạm mật độ mục tiêu. Bước thưa đặt trước nên
	// các cụm lớn chiếm chỗ đẹp và khoảng thở hình thành sớm; bước dày chỉ chạy nếu còn
	// thiếu mật độ, và mỗi lượt lại kiểm mục tiêu nên nó tự dừng ở chỗ vừa đủ.
	for _, pitch := range []float64{clusterPitch, clusterPitch * 0.78, clusterPitch * 0.62, clusterPitch * 0.5} {
		for _, off := range []float64{0, 0.5, 0.25} {
			if g.b.visualDensity() >= target || g.b.breathing() <= g.breathFloorOf() {
				return
			}
			g.clusterLattice(lo, hi, pitch, off, target)
		}
	}
}

// clusterPitch: khoảng cách tâm-tâm giữa hai cụm. Xem phần số học ở `islandsStep`.
const clusterPitch = 2700.0

// breathFloor: bộ sinh ngừng thêm cụm khi khoảng thở xuống tới mức này.
//
// Để dư 4 điểm so với ngưỡng 30% của bộ kiểm: nội dung rải rác của từng khu (nhà, rào, bãi
// đá) còn đặt SAU cụm và sẽ ăn thêm vài phần trăm nữa.
const breathFloor = 0.24

// breathFloorOf: sàn khoảng thở của khu đang dựng.
//
// Dư 12 điểm so với ngưỡng bộ kiểm, không phải 4: nội dung rải rác của khu (nhà, rào, bãi
// đá, mê cung, dải vách) đặt SAU cụm và đo được là ăn mất ~12 điểm khoảng thở. Để dư 4 thì
// cụm dừng ở 24%, nội dung rải rác kéo xuống 14%, và bộ kiểm đỏ.
func (g *gen) breathFloorOf() float64 { return breathFloorFor(g.zoneID) + frameCost(g.zoneID) }

// frameCost: BỘ KHUNG CỐ ĐỊNH của khu ăn mất bao nhiêu điểm khoảng thở sau khi cụm đã đặt
// xong. Đo thực tế, không phải phỏng đoán.
//
// Khu thường mất ~12 điểm vì nhà/rào/bãi đá rải sau. Thiên Đình mất nhiều hơn hẳn vì khung
// của nó là thứ đặt TRƯỚC và không bỏ được: vòng cột quanh Thiên Tiên Đài, bốn sân môn phái
// có tường bao, cộng tám cổng với vùng giữ trống. Để chung 12 điểm thì nó dừng ở 32% rồi
// tụt còn 17% — dưới ngưỡng.
func frameCost(zoneID string) float64 {
	if zoneID == "thien_dinh" {
		return 0.18
	}
	return 0.12
}

// clusterLattice rải một lớp cụm theo lưới có xê dịch.
func (g *gen) clusterLattice(lo, hi, pitch, off, target float64) {
	n := int((hi - lo) / pitch)
	if n < 1 {
		return
	}
	type cell struct{ x, y, d float64 }
	c := g.size / 2
	cells := make([]cell, 0, n*n)
	for iy := 0; iy < n; iy++ {
		for ix := 0; ix < n; ix++ {
			x := lo + (float64(ix)+0.5+off)*pitch
			y := lo + (float64(iy)+0.5+off)*pitch
			cells = append(cells, cell{x, y, math.Hypot(x-c, y-c)})
		}
	}
	// Xa tâm trước: §8.1b mục 5 muốn trống ở giữa (chỗ đánh nhau), dày hơn ở rìa.
	sort.Slice(cells, func(a, b int) bool { return cells[a].d > cells[b].d })
	for _, cl := range cells {
		// Hai điều kiện dừng NGANG HÀNG nhau: đủ mật độ, hoặc hết chỗ thở. Khoảng thở là
		// thứ lượt bố cục 1 làm hỏng (0%) dù mật độ đúng, nên nó phải là ràng buộc chứ không
		// phải số đo hậu kiểm.
		if g.b.visualDensity() >= target || g.b.breathing() <= g.breathFloorOf() {
			return
		}
		g.placeCluster(
			cl.x+g.b.rnd(-islandJitter, islandJitter),
			cl.y+g.b.rnd(-islandJitter, islandJitter),
			g.b.rnd(clusterMinSpan, clusterMaxSpan)/2)
	}
}

const (
	// clusterMinSpan: dưới ~920 thì không bao giờ đạt 30% khoảng thở (xem `islandsStep`).
	clusterMinSpan = 1000.0
	clusterMaxSpan = 1500.0
)

// placeCluster dựng MỘT cụm: nhiều dải vật cản song song, dài ≤ 800, cách nhau đúng hành
// lang tối thiểu — đủ để đi xuyên qua cụm chứ không phải đi vòng.
//
// Ba thứ làm cụm trông có HÌNH DÁNG chứ không phải một đám ô vuông (§8.1b mục 6):
//   - dải chạy theo MỘT hướng chung cho cả cụm (ngang hoặc dọc), nên cụm có "thớ";
//   - độ dài mỗi dải cắt theo đường bao TRÒN của cụm, nên mép cụm cong chứ không vuông;
//   - lõi bậc cao nằm giữa, viền bậc thấp ở hai đầu dải (§8.1b mục 3 "đảo phân tầng").
func (g *gen) placeCluster(cx, cy, r float64) {
	core := g.pickCore()
	rim := islandRim[g.b.rng.Intn(len(islandRim))]
	if !g.b.freeFor(rect{cx - r, cy - r, cx + r, cy + r}, max(g.tierOf(core), g.tierOf(rim)), g.corridorFor) {
		return
	}
	vertical := g.b.rng.Intn(2) == 0
	// Dải XEN KẼ bậc cao / bậc thấp, và khe giữa hai dải kề nhau lấy theo `corridor_min` của
	// bậc THẤP HƠN (§8.1b: "khe giữa vật cản bậc A và bậc B ≥ corridor_min[min(A,B)]").
	//
	// Đây là chỗ quyết định cụm đặc hay rỗng. Nếu mọi dải cùng bậc 9 thì khe phải 340, và
	// với dải dày 200 thì cụm chỉ đặc 200/(200+340) = 37% — đo được mật độ cả khu 8–11%,
	// không bao giờ chạm 18%. Xen kẽ thì khe chỉ cần 160–180, cụm đặc ~70%, và mật độ lên
	// đúng tầm mà vẫn giữ nguyên luật đi lọt: ai không chém được dải bậc cao thì chém được
	// dải bậc thấp ngay cạnh, hoặc đi qua khe 160 vốn đủ rộng cho họ.
	lowGap := g.corridorFor(g.tierOf(rim))
	thick := g.b.rnd(300, 450)
	step := thick + lowGap
	i := 0
	for t := -r; t <= r; t += step {
		half := math.Sqrt(math.Max(r*r-t*t, 0))
		if half < 80 {
			i++
			continue
		}
		kind := core
		if i%2 == 1 {
			kind = rim
		}
		g.segmentedRun(cx, cy, t, half, thick, lowGap, vertical, kind)
		i++
	}
}

// maxRunLen: "Không tường liền dài quá 800" (§8.1b mục 2).
const maxRunLen = 800.0

// segmentedRun đặt một dải dài `2*half` thành từng khúc ≤ `maxRunLen`.
func (g *gen) segmentedRun(cx, cy, t, half, thick, gap float64, vertical bool, kind string) {
	total := 2 * half
	segs := int(math.Ceil(total / maxRunLen))
	if segs < 1 {
		return
	}
	segLen := (total - float64(segs-1)*gap) / float64(segs)
	if segLen < 80 {
		return
	}
	for i := 0; i < segs; i++ {
		a := -half + float64(i)*(segLen+gap)
		if vertical {
			g.b.obstacleForce(kind, cx+t-thick/2, cy+a, thick, segLen)
		} else {
			g.b.obstacleForce(kind, cx+a, cy+t-thick/2, segLen, thick)
		}
	}
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
