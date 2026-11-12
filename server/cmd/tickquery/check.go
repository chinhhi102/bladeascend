package main

import (
	"database/sql"
	"fmt"
	"sort"
	"time"
)

// checks: những phép kiểm tra tính lành mạnh của chính số đo, chạy sau phần thống kê.
func checks(db *sql.DB, all []srv, vn *time.Location) {
	fmt.Printf("\n## 12. Kiểm tra chính số đo\n\n")

	// a) Cùng một mốc ts mà có nhiều build_id → nhiều tiến trình server cùng ghi.
	byTS := map[int64]map[string]bool{}
	for _, s := range all {
		if byTS[s.TS] == nil {
			byTS[s.TS] = map[string]bool{}
		}
		byTS[s.TS][s.Build] = true
	}
	multi := 0
	for _, m := range byTS {
		if len(m) > 1 {
			multi++
		}
	}
	fmt.Printf("a) mốc có >1 build_id cùng ghi (2 server chạy chồng): %d / %d mốc\n", multi, len(byTS))

	// b) Khoảng cách giữa các mốc — sampler đáng lẽ đúng 10 s.
	var ts []int64
	for t := range byTS {
		ts = append(ts, t)
	}
	sort.Slice(ts, func(i, j int) bool { return ts[i] < ts[j] })
	gaps := map[int64]int{}
	for i := 1; i < len(ts); i++ {
		d := ts[i] - ts[i-1]
		if d > 60 {
			d = 999 // gộp "server tắt"
		}
		gaps[d]++
	}
	var gk []int64
	for g := range gaps {
		gk = append(gk, g)
	}
	sort.Slice(gk, func(i, j int) bool { return gaps[gk[i]] > gaps[gk[j]] })
	fmt.Printf("b) khoảng cách giữa hai mốc (giây → số lần), 8 giá trị hay gặp nhất: ")
	for i, g := range gk {
		if i >= 8 {
			break
		}
		name := fmt.Sprintf("%d", g)
		if g == 999 {
			name = ">60"
		}
		fmt.Printf("%s s:%d  ", name, gaps[g])
	}
	fmt.Println()

	// c) avg > p95: cửa sổ có vài tick khổng lồ → server bị treo, không phải luật game nặng.
	froze := 0
	for _, s := range all {
		if s.Avg > s.P95 && s.P95 > 0 {
			froze++
		}
	}
	fmt.Printf("c) dòng có avg > p95 (vài tick khổng lồ kéo trung bình, dấu hiệu TREO): %d / %d\n", froze, len(all))

	// d) Dòng lặp y hệt (avg,p95 không đổi qua nhiều mốc) → vòng game không chạy giữa hai mốc.
	type key struct {
		zone     int
		avg, p95 float64
	}
	last := map[int]key{}
	repeat := 0
	for _, s := range all {
		k := key{s.Zone, s.Avg, s.P95}
		if last[s.Zone] == k {
			repeat++
		}
		last[s.Zone] = k
	}
	fmt.Printf("d) dòng trùng y hệt dòng trước CÙNG KHU (vòng game đứng im): %d / %d\n\n", repeat, len(all))
}

// quiet: so sánh CÓ KIỂM SOÁT nhưng CHỈ những mốc máy rảnh — không có phiên do MÁY điều
// khiển (bài đo Playwright) trong vòng ±5 phút. Bài đo trình duyệt ăn CPU của chính cái máy
// đang chạy server, nên tick lúc đó không nói lên điều gì về luật game (quy tắc 16).
func quiet(all []srv, human map[int64]bool, autoWin func(int64) bool, vn *time.Location) {
	fmt.Printf("## 14. Như §13 nhưng LOẠI mọi mốc có bài đo trình duyệt chạy trong ±5 phút\n\n")
	var rs []srv
	for _, s := range all {
		n := s.Humans + s.Bots
		if s.Zone == 0 && n >= 21 && n <= 30 && human[s.TS] && !autoWin(s.TS) {
			rs = append(rs, s)
		}
	}
	byB := map[string][]srv{}
	for _, s := range rs {
		byB[s.Build] = append(byB[s.Build], s)
	}
	var bs []string
	for b := range byB {
		bs = append(bs, b)
	}
	sort.Strings(bs)
	fmt.Printf("%-24s %6s %9s %9s %9s %9s %7s %7s %7s\n", "build", "n", "avg TB", "p95 TB", "p95 t.vị", "p95 max", "CPU%", "n.vật", "KB/s")
	for _, b := range bs {
		if len(byB[b]) < 20 {
			continue
		}
		var a, p, c, n, k []float64
		for _, s := range byB[b] {
			a = append(a, s.Avg)
			p = append(p, s.P95)
			c = append(c, s.CPU)
			n = append(n, float64(s.Humans+s.Bots))
			k = append(k, s.Kbps)
		}
		fmt.Printf("%-24s %6d %9.2f %9.2f %9.2f %9.2f %7.1f %7.1f %7.1f\n",
			b, len(byB[b]), mean(a), mean(p), pctl(p, 0.5), maxf(p), mean(c), mean(n), mean(k))
	}
	fmt.Printf("(tổng %d dòng còn lại sau khi loại)\n\n", len(rs))
}

// cpuVsTick: tick p95 theo %CPU của tiến trình server cùng lúc.
func cpuVsTick(all []srv, human map[int64]bool) {
	fmt.Printf("## 15. Tick p95 theo %%CPU tiến trình server (chỉ mốc có người thật)\n\n")
	type b struct {
		name   string
		lo, hi float64
	}
	for _, bk := range []b{{"<10%", 0, 10}, {"10–25%", 10, 25}, {"25–50%", 25, 50}, {"50–75%", 50, 75}, {"≥75%", 75, 1e9}} {
		var p []float64
		for _, s := range all {
			if human[s.TS] && s.Humans > 0 && s.CPU >= bk.lo && s.CPU < bk.hi {
				p = append(p, s.P95)
			}
		}
		if len(p) == 0 {
			fmt.Printf("%-8s (không có)\n", bk.name)
			continue
		}
		fmt.Printf("%-8s n=%-6d p95: TB %.2f  trung vị %.2f  max %.2f\n", bk.name, len(p), mean(p), pctl(p, 0.5), maxf(p))
	}
	fmt.Println()
}

// idleZone: khu KHÔNG có người, chỉ 19–21 bot — khối lượng việc gần như KHÔNG ĐỔI giữa các
// build. Nếu tick của khu này cũng tăng cùng lúc với khu 0 thì thủ phạm là cái MÁY (hoặc một
// thay đổi toàn cục), không phải "người chơi làm nặng khu 0".
func idleZone(all []srv, vn *time.Location) {
	fmt.Printf("## 16. Khu ĐỐI CHỨNG: 0 người, 19–21 bot (khối lượng việc không đổi)\n\n")
	byDay := map[string][]float64{}
	byBuild := map[string][]float64{}
	for _, s := range all {
		if s.Humans != 0 || s.Bots < 19 || s.Bots > 21 || s.Zone == 0 {
			continue
		}
		d := time.Unix(s.TS, 0).In(vn).Format("2006-01-02")
		byDay[d] = append(byDay[d], s.P95)
		byBuild[s.Build] = append(byBuild[s.Build], s.P95)
	}
	var ds []string
	for d := range byDay {
		ds = append(ds, d)
	}
	sort.Strings(ds)
	fmt.Printf("%-12s %7s %9s %9s %9s\n", "ngày", "n", "p95 TB", "p95 t.vị", "p95 max")
	for _, d := range ds {
		p := byDay[d]
		fmt.Printf("%-12s %7d %9.2f %9.2f %9.2f\n", d, len(p), mean(p), pctl(p, 0.5), maxf(p))
	}
	fmt.Printf("\n%-24s %7s %9s %9s %9s\n", "build (≥30 dòng)", "n", "p95 TB", "p95 t.vị", "p95 max")
	var bs []string
	for b := range byBuild {
		if len(byBuild[b]) >= 30 {
			bs = append(bs, b)
		}
	}
	sort.Strings(bs)
	for _, b := range bs {
		p := byBuild[b]
		fmt.Printf("%-24s %7d %9.2f %9.2f %9.2f\n", b, len(p), mean(p), pctl(p, 0.5), maxf(p))
	}
	fmt.Println()
}

// window: tick trong MỘT khoảng thời gian cụ thể — để so hai đoạn cùng build, cùng người
// chơi, chỉ khác mức bận của cái máy.
func window(all []srv, from, to int64, name string, vn *time.Location) {
	var p, a []float64
	occ := 0
	for _, s := range all {
		if s.TS < from || s.TS > to || s.Humans == 0 {
			continue
		}
		occ++
		p = append(p, s.P95)
		a = append(a, s.Avg)
	}
	if occ == 0 {
		fmt.Printf("%-28s (không có dòng nào)\n", name)
		return
	}
	over := 0
	for _, v := range p {
		if v > 3.5 {
			over++
		}
	}
	fmt.Printf("%-28s n=%-4d avg TB %.2f · p95: TB %.2f  trung vị %.2f  max %.2f  | >3,5ms %.0f%%\n",
		name, occ, mean(a), mean(p), pctl(p, 0.5), maxf(p), 100*float64(over)/float64(occ))
}

// recent: từng giờ của build đang chạy — để biết đoạn nào có người và tick lúc đó ra sao.
func recent(all []srv, since int64, vn *time.Location) {
	fmt.Printf("## 20. Từng giờ kể từ %s (khu có người)\n\n", time.Unix(since, 0).In(vn).Format("02/01 15:04"))
	type agg struct {
		p     []float64
		occ   int
		rows  int
		maxH  int
		build string
	}
	m := map[string]*agg{}
	for _, s := range all {
		if s.TS < since {
			continue
		}
		k := time.Unix(s.TS, 0).In(vn).Format("02/01 15h")
		if m[k] == nil {
			m[k] = &agg{}
		}
		a := m[k]
		a.rows++
		a.build = s.Build
		if s.Humans > 0 {
			a.occ++
			a.p = append(a.p, s.P95)
			if s.Humans > a.maxH {
				a.maxH = s.Humans
			}
		}
	}
	var ks []string
	for k := range m {
		ks = append(ks, k)
	}
	sort.Strings(ks)
	fmt.Printf("%-12s %7s %9s %8s %10s %10s\n", "giờ", "dòng", "có người", "đông nh", "p95 t.vị", "p95 max")
	for _, k := range ks {
		a := m[k]
		if a.occ == 0 {
			fmt.Printf("%-12s %7d %9d %8s %10s %10s\n", k, a.rows, 0, "–", "–", "–")
			continue
		}
		fmt.Printf("%-12s %7d %9d %8d %10.2f %10.2f\n", k, a.rows, a.occ, a.maxH, pctl(a.p, 0.5), maxf(a.p))
	}
	fmt.Println()
}

// overlap: hai tiến trình server cùng chạy trên một máy thì tranh CPU của nhau — tick của
// cả hai đều xấu đi mà không liên quan gì tới luật game. Dấu hiệu: trong cùng một phút có
// hai build_id khác nhau, hoặc mốc cách nhau 5 s thay vì 10 s.
func overlap(all []srv, vn *time.Location) {
	fmt.Printf("## 19. Hai server chạy chồng trên cùng một máy\n\n")
	byMin := map[string]map[string]bool{}
	for _, s := range all {
		k := time.Unix(s.TS, 0).In(vn).Format("2006-01-02 15:04")
		if byMin[k] == nil {
			byMin[k] = map[string]bool{}
		}
		byMin[k][s.Build] = true
	}
	perDay := map[string]int{}
	totDay := map[string]int{}
	for k, m := range byMin {
		d := k[:10]
		totDay[d]++
		if len(m) > 1 {
			perDay[d]++
		}
	}
	var ds []string
	for d := range totDay {
		ds = append(ds, d)
	}
	sort.Strings(ds)
	fmt.Printf("%-12s %8s %10s\n", "ngày", "số phút", "phút có 2 build")
	for _, d := range ds {
		fmt.Printf("%-12s %8d %10d\n", d, totDay[d], perDay[d])
	}

	// Tick lúc chạy chồng so với lúc chỉ một server.
	two := map[string]bool{}
	for k, m := range byMin {
		if len(m) > 1 {
			two[k] = true
		}
	}
	var a, b []float64
	for _, s := range all {
		k := time.Unix(s.TS, 0).In(vn).Format("2006-01-02 15:04")
		if two[k] {
			a = append(a, s.P95)
		} else {
			b = append(b, s.P95)
		}
	}
	fmt.Printf("\ntick p95 lúc có 2 build cùng phút : trung vị %.2f ms (n=%d)\n", pctl(a, 0.5), len(a))
	fmt.Printf("tick p95 lúc chỉ một server       : trung vị %.2f ms (n=%d)\n\n", pctl(b, 0.5), len(b))
}

// ratio: chia tick khu đông cho tick khu đối chứng CÙNG BUILD. Nhiễu của cái máy đánh vào
// cả hai khu nên phép chia khử bớt nó; cái còn lại mới là "khu đông đắt hơn bao nhiêu lần".
func ratio(all []srv, vn *time.Location) {
	fmt.Printf("## 17. Khu 0 đông ÷ khu đối chứng, cùng build (khử nhiễu của máy)\n\n")
	hot, cold := map[string][]float64{}, map[string][]float64{}
	for _, s := range all {
		n := s.Humans + s.Bots
		switch {
		case s.Zone == 0 && n >= 21 && n <= 30 && s.Humans > 0:
			hot[s.Build] = append(hot[s.Build], s.P95)
		case s.Zone != 0 && s.Humans == 0 && s.Bots >= 19 && s.Bots <= 21:
			cold[s.Build] = append(cold[s.Build], s.P95)
		}
	}
	var bs []string
	for b := range hot {
		if len(hot[b]) >= 20 && len(cold[b]) >= 20 {
			bs = append(bs, b)
		}
	}
	sort.Strings(bs)
	fmt.Printf("%-24s %7s %9s %7s %9s %8s\n", "build", "n đông", "p95 đông", "n đ/c", "p95 đ/c", "tỉ lệ")
	for _, b := range bs {
		h, c := pctl(hot[b], 0.5), pctl(cold[b], 0.5)
		fmt.Printf("%-24s %7d %9.2f %7d %9.2f %8.2f×\n", b, len(hot[b]), h, len(cold[b]), c, h/c)
	}
	fmt.Println()
}

// spread: vài con số đuôi mà bảng §1 không nói — p99, phần vượt nửa ngân sách, số người
// đông nhất từng thấy trong một khu.
func spread(all []srv, human map[int64]bool) {
	fmt.Printf("## 18. Đuôi phân phối và mức đông nhất từng gặp\n\n")
	var p []float64
	maxH, maxN := 0, 0
	hist := map[int]int{}
	for _, s := range all {
		if !human[s.TS] || s.Humans == 0 {
			continue
		}
		p = append(p, s.P95)
		if s.Humans > maxH {
			maxH = s.Humans
		}
		if s.Humans+s.Bots > maxN {
			maxN = s.Humans + s.Bots
		}
		hist[s.Humans]++
	}
	over16 := 0
	for _, v := range p {
		if v > 16.65 {
			over16++
		}
	}
	fmt.Printf("khu có người, lúc có người thật: n=%d · p99 của p95 = %.2f ms · vượt nửa ngân sách (16,65 ms): %d (%.2f%%)\n",
		len(p), pctl(p, 0.99), over16, 100*float64(over16)/float64(len(p)))
	fmt.Printf("đông nhất từng ghi được: %d NGƯỜI trong một khu · %d nhân vật (người+bot) trong một khu\n", maxH, maxN)
	var hs []int
	for h := range hist {
		hs = append(hs, h)
	}
	sort.Ints(hs)
	fmt.Printf("số khu-mốc theo số người trong khu: ")
	for _, h := range hs {
		fmt.Printf("%d người:%d  ", h, hist[h])
	}
	fmt.Println()
	fmt.Println()
}

// controlled: so sánh CÓ KIỂM SOÁT — cùng khu 0, cùng mật độ nhân vật, theo ngày và theo build.
func controlled(all []srv, human map[int64]bool, vn *time.Location) {
	fmt.Printf("## 13. So sánh CÓ KIỂM SOÁT: khu 0 (Thiên Đình), 21–30 nhân vật, lúc có người thật\n\n")
	var rs []srv
	for _, s := range all {
		n := s.Humans + s.Bots
		if s.Zone == 0 && n >= 21 && n <= 30 && human[s.TS] {
			rs = append(rs, s)
		}
	}
	byDay := map[string][]srv{}
	for _, s := range rs {
		byDay[time.Unix(s.TS, 0).In(vn).Format("2006-01-02")] = append(byDay[time.Unix(s.TS, 0).In(vn).Format("2006-01-02")], s)
	}
	var ds []string
	for d := range byDay {
		ds = append(ds, d)
	}
	sort.Strings(ds)
	fmt.Printf("%-12s %6s %9s %9s %9s %9s\n", "ngày", "n", "avg TB", "p95 TB", "p95 t.vị", "p95 max")
	for _, d := range ds {
		var a, p []float64
		for _, s := range byDay[d] {
			a = append(a, s.Avg)
			p = append(p, s.P95)
		}
		fmt.Printf("%-12s %6d %9.2f %9.2f %9.2f %9.2f\n", d, len(byDay[d]), mean(a), mean(p), pctl(p, 0.5), maxf(p))
	}

	fmt.Printf("\nTheo build (chỉ build có ≥ 20 dòng):\n")
	byB := map[string][]srv{}
	for _, s := range rs {
		byB[s.Build] = append(byB[s.Build], s)
	}
	var bs []string
	for b := range byB {
		if len(byB[b]) >= 20 {
			bs = append(bs, b)
		}
	}
	sort.Strings(bs)
	fmt.Printf("%-24s %6s %9s %9s %9s %9s\n", "build", "n", "avg TB", "p95 TB", "p95 t.vị", "p95 max")
	for _, b := range bs {
		var a, p []float64
		for _, s := range byB[b] {
			a = append(a, s.Avg)
			p = append(p, s.P95)
		}
		fmt.Printf("%-24s %6d %9.2f %9.2f %9.2f %9.2f\n", b, len(byB[b]), mean(a), mean(p), pctl(p, 0.5), maxf(p))
	}
	fmt.Println()
}
