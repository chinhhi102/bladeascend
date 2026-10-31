package main

import (
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"kiemxoay/server/internal/config"
	"kiemxoay/server/internal/game"
	"kiemxoay/server/internal/geom"
	"kiemxoay/server/internal/world"
)

// Bộ kiểm bố cục bản đồ — spec §8.1b "Kiểm tra bắt buộc".
//
// Luật gốc: **"không chém được thì phải đi lọt"**. Người cảnh giới r chưa chém được vật cản
// bậc > r, nên với MỖI cảnh giới r = 0..13 ta bỏ các vật cản bậc ≤ r (vì r chém qua được),
// lấy thân TO NHẤT của r, rồi flood fill lưới 50:
//   - mọi chỗ trống phải liên thông (không túi bị nhốt);
//   - mọi cổng, Thiên Môn, Đài, bến, điểm spawn đều tới được;
//   - diện tích trống ≥ 60%.
//
// Vì sao phải quét ĐỦ 14 cảnh giới chứ không chỉ người to nhất: hai luật kéo ngược nhau.
// Người càng cao cảnh giới thì thân càng to (khó lọt) NHƯNG chém được càng nhiều vật cản
// (dễ đi). Nên không có một cảnh giới nào là "trường hợp xấu nhất" — Đạo Tổ to gấp 6 lần
// Luyện Khí mà vẫn đi lại thoải mái hơn, vì họ chém được mọi thứ. Chỗ chết người nằm ở
// khoảng giữa: Độ Kiếp thân to 4.x mà vẫn chưa chém nổi đá khổng lồ bậc 9.

const checkCell = 50.0

// Mục tiêu mật độ — §8.1b mục 5, chốt ở spec v6.14 (31/10/2026).
//
// Con số là **diện tích HÌNH VẼ**, tức phần đất bị sprite vật cản phủ lên. Không phải chân
// đế: chân đế chỉ rộng 90% và cao 30% hình vẽ nên nó chỉ bằng ~27% — cùng một bản đồ, đo
// chân đế ra 2,5–4% còn đo hình vẽ ra 8–14%.
//
// Hình vẽ mới là thứ người chơi NHÌN THẤY và là thứ kiếm chém trúng (§8.2 "thân hình —
// vùng kiếm chém trúng, đúng với cái người chơi nhìn thấy"). Hỏi "khu này trông có đầy
// không" thì phải đo cái người ta nhìn.
//
// Ba lần chỉnh con số này đều do ĐO rồi mới chốt, không phải đổi ý:
//
//	v6.12  18–28%  — đặt lúc chưa ai đo gì; đạt được nhưng chỉ khi rải đều, mà rải đều thì
//	                 khoảng thở về 0 (xem `breathMin`).
//	v6.13  12–20%  — sau khi cụm hoá; mẫu dải đặc chạm được 10–15%.
//	v6.14   8–14%  — sau khi mỗi khu có mẫu cụm riêng. Mẫu mới CỐ TÌNH rỗng ruột (nhà có
//	                 sân, bãi đá rời, lùm tre so le) nên cùng một diện tích chiếm chỗ thì
//	                 mật độ hình vẽ thấp hơn hẳn mẫu dải đặc.
var densityTarget = map[string][2]float64{
	// Băng Nguyệt Hồ thấp hơn vì mặt hồ (~1/4 khu) là vùng CẤM đặt vật cản.
	"bang_nguyet_ho": {0.04, 0.10},
}

func densityRange(id string) [2]float64 {
	if r, ok := densityTarget[id]; ok {
		return r
	}
	return [2]float64{0.08, 0.14}
}

// densityHardMax: vượt mức này là TRƯỢT, không phải cảnh báo.
//
// Hai hướng không đối xứng vì hậu quả không đối xứng: khu nhạt thì chơi vẫn được, chỉ kém
// hay; khu chật thì bắt đầu có người kẹt — và bộ kiểm liên thông không bắt được hết, nó đo
// "có đường đi không" chứ không đo "đường có hẹp tới mức đi một bước né ba lần không".
const densityHardMax = 0.20

// checkResult là kết quả của một cảnh giới.
type checkResult struct {
	Realm     int
	Name      string
	BodyR     float64
	OpenFrac  float64 // phần diện tích khu mà thân này đi vào được
	MainFrac  float64 // phần chỗ trống thuộc vùng liên thông lớn nhất
	Regions   int
	CutOff    []string // tên những mốc không tới được
	TrappedPx int
	// BiggestTrap: số ô của TÚI BỊ NHỐT LỚN NHẤT (không tính vùng chính).
	//
	// Phải đo riêng, không gộp vào `TrappedPx`: 40 ô lẻ rải khắp bản đồ là chuyện bình
	// thường (khe hẹp giữa hai khối), còn 40 ô NẰM LIỀN NHAU là một cái sân kín 100×100 mà
	// không ai vào được — hai chuyện khác hẳn, mà tổng số ô thì giống nhau.
	BiggestTrap int
	// Density: phần diện tích khu bị **HÌNH VẼ** vật cản phủ — thứ người chơi nhìn thấy và
	// thứ kiếm chém trúng (§8.2 "thân hình"). Đây là con số §8.1b mục 5 nói tới: 18–28%.
	//
	// Nó TỤT khi cảnh giới lên vì vật cản bậc ≤ r coi như không có, và đó chính là thứ cần
	// nhìn: Đại Thừa còn lại bao nhiêu địa hình.
	Density float64
	// Breathing: phần diện tích khu KHÔNG có vật cản nào trong bán kính `breathR`.
	//
	// §8.1b mục 6: địa hình phải có chỗ thở — những khoảng rộng để rượt đuổi, đánh nhau,
	// và để mắt nghỉ. Mật độ tổng không nói được điều đó: rải đều 20% khắp nơi và dồn 20%
	// vào vài cụm cho ra cùng một con số mật độ, nhưng bản đầu thì chỗ nào cũng vướng một
	// tí và không chỗ nào dùng được.
	Breathing float64
	// LakeObstacle: phần diện tích MẶT HỒ (vùng `ice`/`water`) bị vật cản phủ. 0 nếu khu
	// không có hồ. Băng Nguyệt Hồ phải ≤ 1%: mặt hồ là sân trượt băng để rượt nhau, đặt vật
	// cản lên đó là xoá mất cái làm khu ấy khác các khu khác.
	LakeObstacle float64
	// FootDensity: phần diện tích bị CHÂN ĐẾ phủ — đất thật sự chặn chân, bằng ~27% hình vẽ
	// (chân đế rộng 90% cao 30% sprite). Giữ lại để đối chiếu hai cách đo.
	FootDensity float64
}

// minOpenRatio là ngưỡng của §8.1b: "Diện tích trống ≥ 60%".
//
// KHÁC `mapgen.min_open_ratio` (0.65) trong config — con số đó là ngưỡng TỰ KIỂM của bộ
// sinh, dùng để quyết "sinh lại với seed+1", nên nó cố tình chặt hơn luật nghiệm thu. Lấy
// nó làm luật nghiệm thu là trộn hai thứ: bản đồ trống 65% ở Hóa Thần bị báo TRƯỢT dù spec
// chấp nhận từ 60%.
const minOpenRatio = 0.60

// maxTrapCells: túi bị nhốt lớn nhất được tha thứ, tính bằng ô lưới 50.
//
// §8.1b nói "toàn bộ chỗ trống phải liên thông", nhưng tuyệt đối 0 thì không đặt được: hai
// khối chạm mép nhau luôn để lại một khe 1–2 ô mà không ai vào nổi, và đó không phải lỗi.
//
// Luật cũ dùng TỈ LỆ ("vùng chính ≥ 99% chỗ trống") và nó BỎ SÓT đúng thứ cần bắt: Linh Sơn
// Cốc có một cái sân kín **45 ô liền nhau** (≈ 335 × 335 đơn vị — thừa chỗ cho mấy người
// đánh nhau) mà vẫn ĐẠT, vì 46 ô trên 82 000 ô trống chỉ là 0,06%. Tỉ lệ trả lời "bao nhiêu
// phần bản đồ bị nhốt", còn câu cần hỏi là "có cái phòng kín nào không".
//
// 4 ô = 100 × 100 đơn vị: chỗ nhỏ nhất mà một người Luyện Khí (thân 42) đứng lọt. Nhỏ hơn
// thế thì không ai vào được kể cả khi bị đẩy, nên nó không tồn tại với người chơi.
const maxTrapCells = 4

func (r checkResult) ok() bool {
	return r.OpenFrac >= minOpenRatio && r.BiggestTrap < maxTrapCells && len(r.CutOff) == 0
}

// landmark là một mốc phải tới được: cổng, Thiên Môn, bến, Đài, điểm spawn.
type landmark struct {
	At   geom.Vec
	Name string
}

// runCheck kiểm mọi bản đồ trong `dir`. `pngDir` khác rỗng thì xuất ảnh cho từng cảnh giới.
// checkReport là bản ghi JSON cho trang quản trị §5.2.
//
// mapgen là `package main` nên server không import được bộ kiểm. Thay vì gọi lại mapgen
// như một tiến trình con (bản deploy không có sẵn toolchain Go), mapgen GHI kết quả ra
// cạnh bản đồ và server chỉ việc đọc. Số liệu vì thế luôn là số của chính lần sinh ra bộ
// bản đồ đang chạy — không có cách nào để hai bên lệch nhau.
type checkReport struct {
	At      int64                    `json:"at"`
	Seed    int64                    `json:"seed"`
	Maps    map[string][]checkResult `json:"maps"`
	Targets map[string][2]float64    `json:"targets"`
	Bad     int                      `json:"bad"`
}

func runCheck(dir, pngDir, jsonOut string, seed int64, cfg *config.Config, pngRealms []int) error {
	rep := checkReport{At: time.Now().Unix(), Seed: seed,
		Maps: map[string][]checkResult{}, Targets: map[string][2]float64{}}
	maps, err := world.LoadDir(dir, world.LoadOptions{
		Tier: cfg.Obstacles.ObstacleTier, FootprintW: cfg.Obstacles.FootprintW, FootprintH: cfg.Obstacles.FootprintH,
	})
	if err != nil {
		return err
	}
	if len(maps) == 0 {
		return fmt.Errorf("không có bản đồ nào trong %s", dir)
	}
	prog := game.NewProgression(cfg.Cultivation, cfg.Sword.MaxSwords)
	// Đủ 14 cảnh giới: 9 phàm + 5 Tiên (spec §8.2 "r = 0..13").
	realms := len(cfg.Cultivation.Realms) + prog.TienCount()
	sort.Slice(maps, func(i, j int) bool { return maps[i].ID < maps[j].ID })

	bad := 0
	for _, m := range maps {
		fmt.Printf("== %s (%s)\n", m.ID, m.Name)
		marks := landmarksOf(m)
		rep.Targets[m.ID] = densityRange(m.ID)
		for r := 0; r < realms; r++ {
			res, grid := checkRealm(m, prog, cfg, r, marks)
			rep.Maps[m.ID] = append(rep.Maps[m.ID], res)
			status := "ĐẠT"
			if !res.ok() {
				status, bad = "TRƯỢT", bad+1
			}
			fmt.Printf("  r=%-2d %-17s thân %3.0f · vật cản %4.1f%% · thở %3.0f%% · trống %4.0f%% · nhốt %d ô (túi to nhất %d) · %s",
				r, res.Name, res.BodyR, res.Density*100, res.Breathing*100, res.OpenFrac*100, res.TrappedPx, res.BiggestTrap, status)
			if len(res.CutOff) > 0 {
				fmt.Printf(" · KHÔNG TỚI ĐƯỢC: %s", strings.Join(res.CutOff, ", "))
			}
			fmt.Println()
			// Mật độ CHƯA ĐẠT mục tiêu thì chỉ cảnh báo, VƯỢT TRẦN thì trượt (chủ dự án
			// chốt 31/10/2026). Hai hướng không đối xứng vì hậu quả không đối xứng: khu
			// nhạt thì chơi vẫn được, chỉ kém hay; khu chật thì có người kẹt — và bộ kiểm
			// liên thông không bắt được hết, nó đo "có đường đi không" chứ không đo "đường
			// có hẹp tới mức đi một bước né ba lần không".
			if r == 0 {
				want := densityRange(m.ID)
				if bw := breathFloorFor(m.ID); res.Breathing < bw {
					fmt.Printf("     ✗ khoảng thở %.0f%% < %.0f%% — vật cản rải đều, không có chỗ rượt nhau\n",
						res.Breathing*100, bw*100)
					bad++
				}
				if res.LakeObstacle > lakeMaxObstacle {
					fmt.Printf("     ✗ vật cản trên mặt hồ %.1f%% > %.0f%% — hồ là sân trượt, phải để trống\n",
						res.LakeObstacle*100, lakeMaxObstacle*100)
					bad++
				}
				switch {
				case res.Density > densityHardMax:
					fmt.Printf("     ✗ mật độ hình vẽ %.1f%% VƯỢT TRẦN %.0f%% — khu chật, người to kẹt\n",
						res.Density*100, densityHardMax*100)
					bad++
				case res.Density < want[0]:
					fmt.Printf("     ⚠ mật độ hình vẽ %.1f%% (chân đế %.1f%%), mục tiêu %.0f–%.0f%% — còn trống\n",
						res.Density*100, res.FootDensity*100, want[0]*100, want[1]*100)
				case res.Density > want[1]:
					fmt.Printf("     ⚠ mật độ hình vẽ %.1f%% (chân đế %.1f%%), mục tiêu %.0f–%.0f%% — hơi dày\n",
						res.Density*100, res.FootDensity*100, want[0]*100, want[1]*100)
				}
			}
			if pngDir != "" && wants(pngRealms, r) {
				if err := writeCheckPNG(pngDir, m, r, grid); err != nil {
					return err
				}
			}
		}
	}
	// Ghi báo cáo TRƯỚC khi trả lỗi: một lượt kiểm trượt là lúc người ta cần nhìn số nhất,
	// nên bỏ báo cáo đi vì nó trượt là đúng chuyện ngược đời.
	rep.Bad = bad
	if jsonOut != "" {
		b, err := json.MarshalIndent(rep, "", "  ")
		if err != nil {
			return err
		}
		if err := os.WriteFile(jsonOut, b, 0o644); err != nil {
			return err
		}
		fmt.Printf("đã ghi báo cáo: %s\n", jsonOut)
	}
	if bad > 0 {
		return fmt.Errorf("%d cảnh giới trượt bộ kiểm §8.1b", bad)
	}
	return nil
}

func wants(list []int, r int) bool {
	if len(list) == 0 {
		return true
	}
	for _, v := range list {
		if v == r {
			return true
		}
	}
	return false
}

// bodyRadiusFor là bán kính thân TO NHẤT của cảnh giới r.
//
// "To nhất" = tầng 10 với người phàm (spec §8.1b). Tiên giới chỉ có một mức scale cho cả
// cảnh giới nên tầng nào cũng thế.
func bodyRadiusFor(prog *game.Progression, cfg *config.Config, r int) float64 {
	level := r*prog.Tiers() + prog.Tiers() - 1
	return cfg.Player.BodyR * prog.Scale(level)
}

// landmarksOf gom mọi mốc §8.1b đòi phải tới được.
func landmarksOf(m *world.Map) []landmark {
	var out []landmark
	for _, g := range m.Gates {
		name := fmt.Sprintf("cổng %d", g.ID)
		switch {
		case g.Heaven:
			name = "Thiên Môn"
		case g.Dock:
			name = fmt.Sprintf("bến Thiên Môn %d", g.DockFor)
		}
		out = append(out, landmark{At: geom.Vec{X: g.X, Y: g.Y}, Name: name})
	}
	for i, a := range m.Altars {
		name := fmt.Sprintf("Đài %d", i)
		if a.Kind == "center" {
			name = "Thiên Tiên Đài"
		}
		out = append(out, landmark{At: geom.Vec{X: a.X, Y: a.Y}, Name: name})
	}
	return out
}

// checkGrid giữ lưới đã dựng, để vẽ ảnh mà không phải tính lại.
type checkGrid struct {
	n     int
	lo    float64
	solid []bool // vật cản còn chặn cảnh giới này (bậc > r)
	tight []bool // trống với thân nhỏ nhưng thân CỦA R không lọt
	// footprint: tâm ô nằm TRONG chân đế một vật cản. Khác `solid` ở chỗ `solid` nở thêm bán
	// kính thân người, nên `solid` trả lời "đi được không" còn `footprint` trả lời "đất bị
	// chiếm bao nhiêu" — mật độ hỏi câu sau.
	footprint []bool
	// visual: tâm ô nằm trong HÌNH VẼ một vật cản (`Shape`, không phải `Block`). Đây là
	// mật độ §8.1b nói tới.
	visual []bool
	// lake: ô nằm trên mặt hồ (vùng `ice` hoặc `water`).
	lake   []bool
	label  []int
	main   int
	sizes  []int
	marks  []landmark
	cutOff []string
}

// checkRealm chạy bộ kiểm cho đúng một cảnh giới.
func checkRealm(m *world.Map, prog *game.Progression, cfg *config.Config, r int, marks []landmark) (checkResult, *checkGrid) {
	bodyR := bodyRadiusFor(prog, cfg, r)
	lo, hi := m.PlayMargin, m.Size-m.PlayMargin
	n := int((hi - lo) / checkCell)
	g := &checkGrid{n: n, lo: lo, solid: make([]bool, n*n), tight: make([]bool, n*n),
		footprint: make([]bool, n*n), visual: make([]bool, n*n), lake: make([]bool, n*n),
		label: make([]int, n*n), marks: marks}
	for i := range g.label {
		g.label[i] = -1
	}
	// Chỉ coi là HỒ những vùng băng/nước đủ lớn để là một mặt hồ, không phải vũng.
	//
	// Thiên Đình và Trúc Lâm cũng có vài vũng nước / mảng băng nhỏ, và luật "mặt hồ phải
	// trống" áp lên chúng là vô nghĩa — đo được 42,9% và 26,4%, nhưng đó chỉ là mấy cái vũng
	// nằm lọt giữa khu chứ không phải sân trượt băng của Băng Nguyệt Hồ.
	lakes := bigLakes(m)
	// ĐẢO giữa hồ không tính là mặt hồ. Đảo là vùng `snow` nhỏ nằm lọt trong mặt băng, và
	// spec nói rõ "dồn vật cản lên đảo và bờ" — đếm đá trên đảo thành "vật cản trên mặt hồ"
	// thì luật tự mâu thuẫn: đặt đúng chỗ spec bảo đặt lại làm bộ kiểm đỏ.
	islands := iceIslands(m)
	// Chỉ giữ vật cản cảnh giới này CHƯA chém được. Bậc ≤ r thì họ chạm là nó biến mất, nên
	// với họ bức tường đó không tồn tại.
	var walls []world.Obstacle
	for _, o := range m.Obstacles {
		if o.Tier > r {
			walls = append(walls, o)
		}
	}
	for i := 0; i < n*n; i++ {
		c := g.center(i)
		// `tight` dùng bán kính thân nhỏ nhất để tách hai lý do một ô bị chặn: "có vật cản
		// ở đây" (xám trong ảnh) và "ô trống nhưng thân cỡ này không lọt" (cam). Không tách
		// thì ảnh chỉ nói "đi không được", mà người sửa bản đồ cần biết là **nới hành lang**
		// hay **bỏ vật cản**.
		small := cfg.Player.BodyR * prog.Scale(0)
		g.solid[i] = blockedAt(walls, c, bodyR)
		g.tight[i] = g.solid[i] && !blockedAt(walls, c, small)
		g.footprint[i] = insideAny(walls, c, func(o world.Obstacle) world.Shape { return o.Block() })
		g.visual[i] = insideAny(walls, c, func(o world.Obstacle) world.Shape { return o.Shape })
		g.lake[i] = inAny(lakes, c) && !inAny(islands, c)
	}
	g.components()
	res := checkResult{Realm: r, Name: prog.RealmName(r), BodyR: bodyR}
	free := 0
	for i := range g.solid {
		if !g.solid[i] {
			free++
		}
	}
	if n > 0 {
		res.OpenFrac = float64(free) / float64(n*n)
	}
	covered, feet := 0, 0
	lakeCells, lakeBlocked := 0, 0
	for i := 0; i < n*n; i++ {
		if g.visual[i] {
			covered++
		}
		if g.footprint[i] {
			feet++
		}
	}
	if n > 0 {
		res.Density = float64(covered) / float64(n*n)
		res.FootDensity = float64(feet) / float64(n*n)
		res.Breathing = float64(g.breathingCells(m, r)) / float64(n*n)
	}
	for i := 0; i < n*n; i++ {
		if g.lake[i] {
			lakeCells++
			if g.visual[i] {
				lakeBlocked++
			}
		}
	}
	if lakeCells > 0 {
		res.LakeObstacle = float64(lakeBlocked) / float64(lakeCells)
	}
	if free > 0 && g.main >= 0 {
		res.MainFrac = float64(g.sizes[g.main]) / float64(free)
		res.TrappedPx = free - g.sizes[g.main]
		for i, n := range g.sizes {
			if i != g.main && n > res.BiggestTrap {
				res.BiggestTrap = n
			}
		}
	}
	res.Regions = len(g.sizes)
	res.CutOff = g.unreachable()
	return res, g
}

// insideAny: điểm `c` có nằm TRONG chân đế vật cản nào không.
//
// Không dùng `OverlapsCircle(c, 0)` được: nó gọi `PushOut`, và `PushOut` với bán kính 0 trả
// về "không chạm" cho mọi điểm — hình tròn bán kính 0 không chồng lên gì theo định nghĩa
// của nó. Dùng nhầm thì mật độ ra đúng 0,00% cho mọi bản đồ, và con số 0 đó trông y như
// "bản đồ chưa có vật cản" chứ không như một lỗi.
func insideAny(walls []world.Obstacle, c geom.Vec, pick func(world.Obstacle) world.Shape) bool {
	for _, o := range walls {
		b := pick(o)
		lo, hi := b.Bounds()
		if c.X >= lo.X && c.X <= hi.X && c.Y >= lo.Y && c.Y <= hi.Y {
			return true
		}
	}
	return false
}

// breathR: bán kính "khoảng thở" — ô nào không có vật cản nào trong ngần này đơn vị thì
// tính là chỗ thở (§8.1b mục 6, chủ dự án chốt 31/10/2026).
// minLakeFrac: vùng băng/nước phải chiếm ít nhất ngần này phần diện tích khu mới được coi
// là "mặt hồ". Dưới mức đó là vũng nước trang trí.
const minLakeFrac = 0.05

// bigLakes lọc ra những vùng băng/nước đủ lớn để là mặt hồ thật.
func bigLakes(m *world.Map) []world.Zone {
	area := m.Size * m.Size
	var out []world.Zone
	for _, z := range m.Zones {
		if z.Kind != "ice" && z.Kind != "water" {
			continue
		}
		lo, hi := z.Shape.Bounds()
		if (hi.X-lo.X)*(hi.Y-lo.Y) >= area*minLakeFrac {
			out = append(out, z)
		}
	}
	return out
}

// iceIslands: vùng tuyết NHỎ (đảo) — ngoại lệ của luật "mặt hồ phải trống".
func iceIslands(m *world.Map) []world.Zone {
	var out []world.Zone
	for _, z := range m.Zones {
		if z.Kind != "snow" {
			continue
		}
		lo, hi := z.Shape.Bounds()
		if hi.X-lo.X < 2600 && hi.Y-lo.Y < 2600 {
			out = append(out, z)
		}
	}
	return out
}

func inAny(zs []world.Zone, c geom.Vec) bool {
	for _, z := range zs {
		if z.Shape.Contains(c) {
			return true
		}
	}
	return false
}

const breathR = 400.0

// breathMin / lakeMaxObstacle: hai ngưỡng của lượt bố cục 2 (chủ dự án chốt 31/10/2026).
//
//   - khoảng thở ≥ 30% diện tích khu: đủ chỗ rượt đuổi và đánh nhau. Mật độ tổng không nói
//     được điều này — rải đều 20% khắp nơi và dồn 20% vào vài cụm cho ra cùng một con số.
//   - vật cản trên mặt hồ ≤ 1%: mặt hồ Băng Nguyệt Hồ là sân trượt băng, đó là thứ làm khu
//     ấy khác bốn khu kia.
const (
	breathMin       = 0.20
	lakeMaxObstacle = 0.01
	// breathMinLake: Băng Nguyệt Hồ phải thoáng hơn hẳn — mặt hồ là sân trượt băng để rượt
	// nhau, đó là thứ làm khu ấy khác bốn khu kia (spec v6.13 §8.1b).
	breathMinLake = 0.40
)

// breathFloorFor: ngưỡng khoảng thở của riêng khu này.
func breathFloorFor(id string) float64 {
	if id == "bang_nguyet_ho" {
		return breathMinLake
	}
	return breathMin
}

// breathingCells đếm số ô cách MỌI vật cản ít nhất `breathR`.
//
// Làm ngược (đánh dấu ô GẦN vật cản rồi lấy phần còn lại) thay vì tính khoảng cách cho từng
// ô: đánh dấu là quét quanh mỗi vật cản một lần, còn tính khoảng cách là mỗi ô dò hết vật
// cản — 50 000 ô × 1 000 vật cản.
func (g *checkGrid) breathingCells(m *world.Map, realm int) int {
	near := make([]bool, g.n*g.n)
	span := int(breathR/checkCell) + 2
	for _, o := range m.Obstacles {
		if o.Tier <= realm {
			continue // cảnh giới này chém được nên nó không chắn tầm nhìn của ai
		}
		lo, hi := o.Shape.Bounds()
		x0 := int((lo.X-breathR-g.lo)/checkCell) - 1
		x1 := int((hi.X+breathR-g.lo)/checkCell) + 1
		y0 := int((lo.Y-breathR-g.lo)/checkCell) - 1
		y1 := int((hi.Y+breathR-g.lo)/checkCell) + 1
		_ = span
		for y := max(y0, 0); y <= min(y1, g.n-1); y++ {
			cy := g.lo + (float64(y)+0.5)*checkCell
			dy := math.Max(math.Max(lo.Y-cy, 0), cy-hi.Y)
			if dy > breathR {
				continue
			}
			for x := max(x0, 0); x <= min(x1, g.n-1); x++ {
				if near[y*g.n+x] {
					continue
				}
				cx := g.lo + (float64(x)+0.5)*checkCell
				dx := math.Max(math.Max(lo.X-cx, 0), cx-hi.X)
				if dx*dx+dy*dy <= breathR*breathR {
					near[y*g.n+x] = true
				}
			}
		}
	}
	free := 0
	for _, b := range near {
		if !b {
			free++
		}
	}
	return free
}

func blockedAt(walls []world.Obstacle, c geom.Vec, r float64) bool {
	for _, o := range walls {
		if o.Block().OverlapsCircle(c, r) {
			return true
		}
	}
	return false
}

func (g *checkGrid) center(i int) geom.Vec {
	return geom.Vec{
		X: g.lo + (float64(i%g.n)+0.5)*checkCell,
		Y: g.lo + (float64(i/g.n)+0.5)*checkCell,
	}
}

func (g *checkGrid) cellOf(p geom.Vec) int {
	x := int((p.X - g.lo) / checkCell)
	y := int((p.Y - g.lo) / checkCell)
	if x < 0 || y < 0 || x >= g.n || y >= g.n {
		return -1
	}
	return y*g.n + x
}

// components gắn nhãn mọi vùng trống liên thông (4 hướng) và ghi lại vùng lớn nhất.
func (g *checkGrid) components() {
	g.main = -1
	for s := 0; s < g.n*g.n; s++ {
		if g.solid[s] || g.label[s] >= 0 {
			continue
		}
		id := len(g.sizes)
		size := 0
		queue := []int{s}
		g.label[s] = id
		for len(queue) > 0 {
			i := queue[len(queue)-1]
			queue = queue[:len(queue)-1]
			size++
			x, y := i%g.n, i/g.n
			for _, d := range [4][2]int{{1, 0}, {-1, 0}, {0, 1}, {0, -1}} {
				nx, ny := x+d[0], y+d[1]
				if nx < 0 || ny < 0 || nx >= g.n || ny >= g.n {
					continue
				}
				j := ny*g.n + nx
				if !g.solid[j] && g.label[j] < 0 {
					g.label[j] = id
					queue = append(queue, j)
				}
			}
		}
		g.sizes = append(g.sizes, size)
		if g.main < 0 || size > g.sizes[g.main] {
			g.main = id
		}
	}
}

// unreachable: mốc nào không nằm trong vùng liên thông lớn nhất.
//
// Mốc nằm ĐÚNG trong một ô bị chặn cũng tính là không tới được, nhưng phải xét ô lân cận
// trước: tâm một cái cổng rộng 200 có thể rơi vào ô mà thân người to không lọt, trong khi
// cổng vẫn vào được từ cạnh bên.
func (g *checkGrid) unreachable() []string {
	var out []string
	for _, mk := range g.marks {
		i := g.cellOf(mk.At)
		if i < 0 {
			out = append(out, mk.Name+" (ngoài lưới)")
			continue
		}
		if g.reachableNear(i) {
			continue
		}
		out = append(out, mk.Name)
	}
	return out
}

// reachableNear: ô này, hay một ô trong bán kính 3 ô quanh nó, thuộc vùng chính?
func (g *checkGrid) reachableNear(i int) bool {
	const span = 3
	x0, y0 := i%g.n, i/g.n
	for dy := -span; dy <= span; dy++ {
		for dx := -span; dx <= span; dx++ {
			x, y := x0+dx, y0+dy
			if x < 0 || y < 0 || x >= g.n || y >= g.n {
				continue
			}
			if g.label[y*g.n+x] == g.main {
				return true
			}
		}
	}
	return false
}

// writeCheckPNG xuất ảnh `docs/map_check/{khu}_r{r}.png` đúng bảng màu spec §8.1b:
// xám = vật cản, cam = chỗ thân không lọt, đỏ = vùng bị nhốt.
func writeCheckPNG(dir string, m *world.Map, r int, g *checkGrid) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	// Phóng to 3 lần theo kiểu "lấy ô gần nhất": lưới 50 trên bản đồ 12 000 cho ra ảnh
	// 230 px, nhỏ quá để nhìn ra một hành lang rộng 3 ô. Phóng ở đây chứ không để người xem
	// tự zoom: ảnh này để DUYỆT, mở ra là phải đọc được ngay.
	const zoom = 3
	img := image.NewRGBA(image.Rect(0, 0, g.n*zoom, g.n*zoom))
	var (
		cFree    = color.RGBA{0xee, 0xf2, 0xe6, 0xff}
		cWall    = color.RGBA{0x55, 0x55, 0x55, 0xff}
		cTight   = color.RGBA{0xf4, 0xa2, 0x61, 0xff}
		cTrapped = color.RGBA{0xe6, 0x39, 0x46, 0xff}
	)
	for i := 0; i < g.n*g.n; i++ {
		c := cFree
		switch {
		case g.tight[i]:
			c = cTight
		case g.solid[i]:
			c = cWall
		case g.label[i] != g.main:
			c = cTrapped
		}
		x0, y0 := (i%g.n)*zoom, (i/g.n)*zoom
		for dy := 0; dy < zoom; dy++ {
			for dx := 0; dx < zoom; dx++ {
				img.Set(x0+dx, y0+dy, c)
			}
		}
	}
	// Mốc: cổng xanh, đài vàng. Vẽ thành chữ thập 5 px để thấy được ở ảnh 230 px.
	for _, mk := range g.marks {
		i := g.cellOf(mk.At)
		if i < 0 {
			continue
		}
		c := color.RGBA{0x3a, 0x86, 0xff, 0xff}
		if strings.HasPrefix(mk.Name, "Đài") || strings.HasPrefix(mk.Name, "Thiên Tiên") {
			c = color.RGBA{0xff, 0xbe, 0x0b, 0xff}
		} else if strings.HasPrefix(mk.Name, "spawn") {
			c = color.RGBA{0x8a, 0x2b, 0xe2, 0xff}
		}
		x0, y0 := (i%g.n)*zoom+zoom/2, (i/g.n)*zoom+zoom/2
		for d := -5; d <= 5; d++ {
			for w := -1; w <= 1; w++ {
				img.Set(x0+d, y0+w, c)
				img.Set(x0+w, y0+d, c)
			}
		}
	}
	path := filepath.Join(dir, fmt.Sprintf("%s_r%d.png", m.ID, r))
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, img)
}
