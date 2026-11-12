// Command leakcheck — bài kiểm rò rỉ phía SERVER (GAME_SPEC §13.9).
//
//	cd server && go run ./cmd/leakcheck                 # 60 phút game (mặc định của spec)
//	cd server && go run ./cmd/leakcheck -minutes 5      # bản ngắn cho cổng trước deploy
//	cd server && go run ./cmd/leakcheck -minutes 60 -v  # in mốc từng phút
//
// Spec §13.9: "loadtest 1 giờ có người vào/ra liên tục → goroutine và heap về lại mức nền
// (±10%) sau khi mọi người rời".
//
// ── Vì sao không chạy qua mạng thật ──────────────────────────────────────────────────────
//
// Bài này đo CHÍNH tiến trình server, nên nó chạy phòng trong cùng tiến trình với nó: đo
// heap của một tiến trình khác qua mạng là đo được tổng, không đo được "phần nào của tổng
// là do người chơi vào ra". Tick chạy bằng đồng hồ GIẢ (`r.Step(dt)` trong vòng lặp), nên
// một giờ game chạy xong trong vài chục giây và kết quả không phụ thuộc tải máy — khác hẳn
// các bài đo FPS, vốn phải đợi máy rảnh (CLAUDE.md luật 16).
//
// ── Đọc kết quả ──────────────────────────────────────────────────────────────────────────
//
// Mốc NỀN lấy sau khi phòng đã chạy ấm (bot đã lấp đủ sàn, bản đồ đã sinh xong). So sánh
// là nền ↔ sau khi mọi người rời VÀ đã chạy thêm vài lượt quét mồ côi 60 s. Ngưỡng 10%
// theo spec; thoát mã 1 khi vượt, để cổng deploy chặn được.
package main

import (
	"flag"
	"fmt"
	"math/rand"
	"os"
	"runtime"
	"runtime/pprof"

	"bladeascend/server/internal/config"
	"bladeascend/server/internal/netproto"
	"bladeascend/server/internal/room"
)

// nullConn nuốt mọi gói: bài này đo bộ nhớ của server, không đo client.
type nullConn struct{}

func (nullConn) Send([]byte) bool { return true }

// sample là một mốc đo.
type sample struct {
	label      string
	heapMB     float64
	goroutines int
	objects    uint64
	// rooms: số phòng đang sống lúc lấy mẫu.
	//
	// Heap phải chia cho con số này mới so được. Mỗi phòng là một thế giới ≈ 8 MB (5 khu,
	// 28 000 viên EXP, 5 lưới A*), nên một phòng mở thêm làm heap nhảy 8 MB — và đó là
	// game đang chạy đúng, không phải rò rỉ. Thiếu phép chia này thì bài đo đỏ mỗi lần
	// người chơi đông lên.
	rooms int
}

func (s sample) String() string {
	return fmt.Sprintf("%-22s heap %7.2f MB · goroutine %4d · object %d", s.label, s.heapMB, s.goroutines, s.objects)
}

// measure ép GC hai lần rồi đọc.
//
// Hai lần chứ không một: lần đầu chạy finalizer và thả những thứ chỉ còn tham chiếu yếu,
// lần hai mới dọn được chính những thứ lần đầu vừa thả. Đo sau một lần GC thì con số "sau"
// luôn cao hơn thật và bài test báo rò rỉ giả.
func measure(label string) sample {
	runtime.GC()
	runtime.GC()
	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)
	return sample{
		label:      label,
		heapMB:     float64(ms.HeapAlloc) / (1024 * 1024),
		goroutines: runtime.NumGoroutine(),
		objects:    ms.HeapObjects,
	}
}

// perRoom là heap trên MỘT phòng đang sống — con số duy nhất so được giữa hai thời điểm.
func (s sample) perRoom() float64 {
	if s.rooms <= 0 {
		return s.heapMB
	}
	return s.heapMB / float64(s.rooms)
}

func main() {
	minutes := flag.Float64("minutes", 60, "số phút GAME mô phỏng (spec §13.9: 60)")
	churn := flag.Int("churn", 6, "số người vào/ra cùng lúc")
	everySecs := flag.Float64("every", 8, "mỗi người ở lại bao nhiêu giây rồi rời")
	tol := flag.Float64("tol", 10, "ngưỡng tăng cho phép, %")
	verbose := flag.Bool("v", false, "in mốc mỗi phút game")
	seed := flag.Int64("seed", 7, "random seed")
	heapProf := flag.String("heapprofile", "", "ghi hồ sơ heap sau khi đo")
	exact := flag.Bool("exact", false, "hồ sơ heap chính xác từng lần cấp phát (chậm)")
	flag.Parse()
	if *exact {
		runtime.MemProfileRate = 1
	}

	cfg := config.MustLoadDefault()
	// Trần phòng phải rộng hơn hẳn `churn`: từ §7.1 một người mất kết nối vẫn GIỮ SUẤT
	// thêm 20 giây (nhân vật còn đứng trong thế giới). Với `churn` người ở lại ~8 giây,
	// số suất đang giữ ở trạng thái ổn định là churn × (8+20)/8 ≈ 3,5 × churn. Đặt trần
	// sát quá thì phòng thứ hai mọc ra giữa bài đo và heap nhảy một bậc — bậc đó là một
	// THẾ GIỚI MỚI, không phải rò rỉ, nhưng nó che mất thứ cần nhìn.
	cfg.Room.MaxHumans = max(cfg.Room.MaxHumans, *churn*6+10)
	m := room.NewManager(cfg, nil, nil)
	m.Start = func(*room.Room) {} // tự tick, không chạy goroutine riêng
	rng := rand.New(rand.NewSource(*seed))

	// Một người ở lại suốt, để phòng không bao giờ đóng: bài này đo rò rỉ TRONG một phòng
	// sống lâu. Phòng đóng rồi mở lại thì mọi thứ về 0 và bài test luôn xanh mà vô nghĩa.
	anchor := room.NewClient(nullConn{}, false)
	m.Join(anchor, netproto.Join{Name: "NeoTru", Char: "a_hon"})
	r := m.AllRooms()[0]
	dt := 1 / float64(cfg.Net.TickHz)
	// Tick MỌI phòng, không chỉ phòng đầu.
	//
	// Ngoài đời mỗi phòng có goroutine riêng (`Manager.Start`). Ở đây `Start` là hàm rỗng
	// để tick bằng đồng hồ giả, nên nếu chỉ tick phòng đầu thì phòng thứ hai (sinh ra khi
	// phòng đầu đầy) đứng im mãi: chỗ giữ §7.1 không bao giờ hết hạn, phòng không bao giờ
	// đóng, và bài test báo "rò rỉ" cho một thứ chỉ tồn tại trong chính bài test.
	step := func() {
		for _, rm := range m.AllRooms() {
			rm.Step(dt)
		}
	}
	// Rời phòng qua CHÍNH phòng của người đó.
	//
	// Không được giữ một con trỏ phòng rồi gọi `r.RemoveClient` cho mọi người: phòng có
	// thể đóng giữa chừng (người cuối chết) và người mới sẽ vào một phòng khác. Gọi nhầm
	// phòng thì `RemoveClient` không tìm thấy ai và im lặng không làm gì — nhân vật nằm
	// lại trong phòng thật, suất người không được trả, phòng đầy dần và server mở thêm
	// phòng mới. Bài đo sẽ thấy heap tăng đều và báo "rò rỉ" cho một lỗi của chính nó.
	leave := func(c *room.Client) bool {
		rm := c.Room()
		if rm == nil {
			return false
		}
		rm.RemoveClient(c)
		return true
	}
	// Chạy ấm: bot lấp đủ sàn mọi khu, quái và vật phẩm sinh đủ. Lấy mốc nền TRƯỚC khi ấm
	// thì phần lớn "mức tăng" đo được chỉ là phòng đang dựng xong chính nó.
	warm := 120.0
	for t := 0.0; t < warm; t += dt {
		step()
	}
	base := measure("nền (sau khi ấm)")
	fmt.Println(base)
	dumpZones(r, "nền")

	var marks []sample
	trendBad := false
	type live struct {
		c      *room.Client
		leaves float64
	}
	var people []live
	total := *minutes * 60
	joined, left := 0, 0
	nextMark := 60.0
	for t := 0.0; t < total; t += dt {
		// Người rời.
		keep := people[:0]
		for _, p := range people {
			if t >= p.leaves {
				if leave(p.c) {
					left++
				}
				continue
			}
			keep = append(keep, p)
		}
		// Xoá phần đuôi: `people[:0]` chỉ hạ `len`, mảng nền vẫn giữ con trỏ Client —
		// đúng cái bẫy vừa sửa trong `itemGrid.remove`, và một bài đo rò rỉ mà tự rò rỉ
		// thì vô dụng.
		for i := len(keep); i < len(people); i++ {
			people[i] = live{}
		}
		people = keep
		// Người neo có thể bị bot chém chết (họ là người chơi thật trong một phòng đầy
		// bot). Chết là rời phòng (§7 v6.29), nên phải cho vào lại — không thì phòng rỗng,
		// đóng lại, và phần còn lại của bài đo chạy trên một phòng khác.
		if anchor.Room() == nil {
			anchor = room.NewClient(nullConn{}, false)
			m.Join(anchor, netproto.Join{Name: "NeoTru", Char: "a_hon"})
		}
		// Người vào, giữ đủ `churn` người ngoài người neo.
		for len(people) < *churn {
			c := room.NewClient(nullConn{}, false)
			m.Join(c, netproto.Join{Name: fmt.Sprintf("Khach%04d", joined), Char: cfg.Characters[rng.Intn(len(cfg.Characters))]})
			joined++
			people = append(people, live{c: c, leaves: t + *everySecs*(0.5+rng.Float64())})
		}
		step()
		if t >= nextMark {
			nextMark += 60
			mark := measure(fmt.Sprintf("phút %.0f", t/60))
			mark.rooms = len(m.AllRooms())
			marks = append(marks, mark)
			if !*verbose {
				continue
			}
			gc, sl, it, qo := 0, 0, 0, 0
			for _, z := range r.Zones() {
				a := z.Game().GridAudit()
				st, _, _ := z.Game().SlotAudit()
				o, _ := z.Game().QueueAudit()
				gc += a.Cap
				sl += st
				it += len(z.Game().Items)
				qo += o
			}
			res := ""
			for _, rm := range m.AllRooms() {
				res += fmt.Sprintf(" %s:%d/%d", rm.ID, rm.Reserved(), rm.HeldCount())
			}
			_ = res
			fmt.Printf("%s · mã %d · phòng%s · lướicap %d · item %d · chờ %d\n",
				mark, m.ResumeCount(), res, gc, it, qo)
			_ = sl
		}
	}

	// Mọi người rời. Người neo rời SAU CÙNG — còn một người thì phòng còn sống, và ta muốn
	// đo cả phần dọn của phòng đang sống (mã phiên §7.1, chỗ giữ, tên bot).
	for _, p := range people {
		if leave(p.c) {
			left++
		}
	}
	people = nil
	// §7.1 giữ nhân vật 20 giây sau khi mất kết nối, và §13.9 quét mồ côi mỗi 60 giây.
	// Chạy đủ 3 lượt quét để mọi đường dọn đều được chạy ít nhất một lần.
	for t := 0.0; t < 3*room.SweepEvery.Seconds()+room.HoldSecs+5; t += dt {
		step()
	}
	leave(anchor)
	for t := 0.0; t < room.HoldSecs+room.SweepEvery.Seconds()+5; t += dt {
		step()
	}
	// Thả MỌI tham chiếu tới phòng trước khi đo lần cuối.
	//
	// Chính biến `r` của bài đo cũng giữ cả một thế giới (5 khu, ~28 000 viên EXP, 5 lưới
	// A*) sống mãi sau khi phòng đã đóng. Quên dòng này thì bài đo luôn báo "còn 8 MB" và
	// 8 MB đó là do nó cầm.
	dumpZones(r, "trước khi thả")
	r = nil
	anchor = nil
	after := measure("sau khi mọi người rời")
	fmt.Println(after)
	if *heapProf != "" {
		f, err := os.Create(*heapProf)
		if err == nil {
			_ = pprof.WriteHeapProfile(f)
			_ = f.Close()
			fmt.Printf("hồ sơ heap: %s\n", *heapProf)
		}
	}

	// Xu hướng TRONG lúc chơi: so phút thứ 2 (thế giới đã ổn định) với phút cuối.
	//
	// Phép so "nền ↔ sau khi mọi người rời" của spec bắt được thứ KHÔNG ĐƯỢC THU HỒI, nhưng
	// nó mù với thứ tăng đều rồi được thu hồi hết ở cuối — mà đó mới là dạng giết server
	// thật: một giờ tăng 7 MB/phút là 420 MB, tiến trình chết trước khi kịp "mọi người rời".
	if len(marks) >= 4 {
		// So NỬA SAU với mốc cuối, không so với phút 2.
		//
		// Vài phút đầu thế giới còn đang lớn lên (bot lấp đủ sàn từng khu, quái và vật
		// phẩm sinh dần, phòng thứ hai mở ra khi phòng đầu đầy). Lấy phút 2 làm mốc là
		// tính cả phần "đang dựng xong chính nó" vào mức tăng, và bài đo đỏ vì một thứ
		// đúng. Từ nửa sau trở đi thì mọi thứ đã ổn định, và bất kỳ mức tăng nào cũng là
		// mức tăng thật.
		a, b := marks[len(marks)/2], marks[len(marks)-1]
		av, bv := a.perRoom(), b.perRoom()
		gr := 0.0
		if av > 0 {
			gr = (bv - av) / av * 100
		}
		perMin := (bv - av) / float64(len(marks)-1-len(marks)/2)
		ok := gr <= *tol || bv-av <= noiseFloor("heap MB")
		fmt.Printf("\nxu hướng khi đang chơi (heap / phòng): %s %.2f MB → %s %.2f MB = %+.1f%% (%+.2f MB/phút)  %s\n",
			a.label, av, b.label, bv, gr, perMin, map[bool]string{true: "ĐẠT", false: "TRƯỢT"}[ok])
		if !ok {
			trendBad = true
		}
	}

	fmt.Printf("\n%d lượt vào · %d lượt rời · %.0f phút game · %d phòng\n", joined, left, *minutes, len(m.AllRooms()))
	if n := m.ResumeCount(); n != 0 {
		fmt.Printf("TRƯỢT: còn %d mã phiên §7.1 trong sổ sau khi mọi người rời\n", n)
		os.Exit(1)
	}
	held := 0
	for _, rm := range m.AllRooms() {
		held += rm.HeldCount()
	}
	_ = held
	if n := held; n != 0 {
		fmt.Printf("TRƯỢT: còn %d chỗ giữ §7.1 sau khi mọi người rời\n", n)
		os.Exit(1)
	}

	bad := trendBad
	check := func(name string, from, to float64) {
		// Nền rất nhỏ thì phần trăm vô nghĩa (1 → 2 goroutine là +100%). Bỏ qua khi mức
		// tăng tuyệt đối còn dưới một ngưỡng tiếng ồn.
		grow := to - from
		pct := 0.0
		if from > 0 {
			pct = grow / from * 100
		}
		ok := pct <= *tol || grow <= noiseFloor(name)
		if !ok {
			bad = true
		}
		fmt.Printf("%-12s %8.2f → %8.2f  (%+.1f%%)  %s\n", name, from, to, pct,
			map[bool]string{true: "ĐẠT", false: "TRƯỢT"}[ok])
	}
	fmt.Println()
	check("heap MB", base.heapMB, after.heapMB)
	check("goroutine", float64(base.goroutines), float64(after.goroutines))
	check("object", float64(base.objects), float64(after.objects))

	if bad {
		fmt.Printf("\nTRƯỢT: vượt ngưỡng %.0f%% của §13.9.\n", *tol)
		os.Exit(1)
	}
	fmt.Printf("\nĐẠT: mọi chỉ số về lại mức nền trong ±%.0f%% (§13.9).\n", *tol)
}

// dumpZones in số entity từng khu — để phân biệt "rò rỉ" với "thế giới đang lớn lên".
func dumpZones(r *room.Room, label string) {
	for i, z := range r.Zones() {
		g := z.Game()
		byKind := map[string]int{}
		for _, it := range g.Items {
			byKind[it.Kind]++
		}
		ga := g.GridAudit()
		st, sf, su := g.SlotAudit()
		qo, qi := g.QueueAudit()
		fmt.Printf("  [%s] khu %d: người %d · quái %d · vật phẩm %d %v\n", label, i,
			len(g.Players), len(g.Monsters), len(g.Items), byKind)
		fmt.Printf("        lưới ô %d len %d cap %d SÓT %d · slot %d/trống %d/dùng %d · hàng chờ viên %d vật phẩm %d\n",
			ga.Cells, ga.Len, ga.Cap, ga.Stale, st, sf, su, qo, qi)
	}
}

// noiseFloor: mức tăng tuyệt đối coi như tiếng ồn, theo từng chỉ số.
//
// Heap của một tiến trình Go nhảy vài trăm KB chỉ vì bộ cấp phát giữ lại span; goroutine
// thì một hai con là chuyện của runtime (bộ dọn rác, timer). Không có sàn này thì bài test
// đỏ ngẫu nhiên, và một bài test đỏ ngẫu nhiên sẽ bị tắt đi trong vòng một tuần.
func noiseFloor(name string) float64 {
	switch name {
	case "heap MB":
		return 1.5
	case "goroutine":
		return 2
	default:
		return 20000
	}
}
