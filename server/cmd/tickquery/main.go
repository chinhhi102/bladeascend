// tickquery: đọc telemetry.db và in số liệu TICK SERVER trong những lúc NGƯỜI THẬT đang chơi.
//
// Vì sao cần, khi `docs/perf/real/` đã có dòng S6 tự động: dòng S6 gộp tick của CẢ khoảng,
// kể cả những phút không ai nối và những phút chỉ có bài đo Playwright. Tick chỉ có nghĩa
// khi đặt cạnh "khu nào, đông bao nhiêu, ai đang nối, máy có bận không" — đó là việc của
// công cụ này. Chỉ đọc, không ghi gì vào DB. Báo cáo: `docs/TICK_NGUOI_THAT.md`.
package main

import (
	"database/sql"
	"flag"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

type srv struct {
	TS       int64
	Zone     int
	ZoneName string
	Avg, P95 float64
	Humans   int
	Bots     int
	Kbps     float64
	CPU, RSS float64
	Build    string
}

type iv struct {
	from, to int64
	auto     bool
	id       string
}

func pctl(v []float64, p float64) float64 {
	if len(v) == 0 {
		return math.NaN()
	}
	s := append([]float64(nil), v...)
	sort.Float64s(s)
	i := int(float64(len(s)-1)*p + 0.5)
	return s[i]
}
func mean(v []float64) float64 {
	if len(v) == 0 {
		return math.NaN()
	}
	t := 0.0
	for _, x := range v {
		t += x
	}
	return t / float64(len(v))
}
func maxf(v []float64) float64 {
	m := math.NaN()
	for _, x := range v {
		if math.IsNaN(m) || x > m {
			m = x
		}
	}
	return m
}

func main() {
	db_ := flag.String("db", "", "đường dẫn telemetry.db")
	days := flag.Int("days", 0, "chỉ lấy N ngày gần nhất (0 = tất cả)")
	wins := flag.String("win", "", `so các khoảng, ngăn bằng dấu phẩy: "2026-11-12 09:00..09:31,2026-11-12 09:31..09:50"`)
	flag.Parse()

	db, err := sql.Open("sqlite", *db_+"?_pragma=busy_timeout(5000)")
	must(err)
	defer db.Close()

	vn := time.FixedZone("ICT", 7*3600)

	// 1. Khoảng hoạt động của từng phiên, theo mẫu client thật sự gửi về.
	var ivs []iv
	rows, err := db.Query(`SELECT s.session_id, COALESCE(ss.automated,0), MIN(s.server_ts), MAX(s.server_ts)
		FROM samples s LEFT JOIN sessions ss ON ss.session_id=s.session_id GROUP BY s.session_id`)
	must(err)
	for rows.Next() {
		var id string
		var auto int
		var lo, hi int64
		must(rows.Scan(&id, &auto, &lo, &hi))
		ivs = append(ivs, iv{from: lo - 20, to: hi + 20, auto: auto != 0, id: id})
	}
	rows.Close()

	// 2. Mẫu server.
	where := ""
	var args []any
	if *days > 0 {
		where = " WHERE ts >= ?"
		args = append(args, time.Now().AddDate(0, 0, -*days).Unix())
	}
	var all []srv
	rows, err = db.Query(`SELECT ts,zone,zone_name,COALESCE(tick_avg,0),COALESCE(tick_p95,0),
		COALESCE(humans,0),COALESCE(bots,0),COALESCE(kbps_out,0),COALESCE(cpu_pct,0),COALESCE(rss_mb,0),build_id
		FROM server_samples`+where+` ORDER BY ts`, args...)
	must(err)
	for rows.Next() {
		var s srv
		must(rows.Scan(&s.TS, &s.Zone, &s.ZoneName, &s.Avg, &s.P95, &s.Humans, &s.Bots, &s.Kbps, &s.CPU, &s.RSS, &s.Build))
		all = append(all, s)
	}
	rows.Close()

	if len(all) == 0 {
		fmt.Println("không có mẫu server nào")
		return
	}

	// 3. Gắn nhãn mỗi mốc thời gian: có người thật / có máy / không ai.
	type lab struct{ real, auto int }
	labs := map[int64]lab{}
	tsSet := map[int64]bool{}
	for _, s := range all {
		tsSet[s.TS] = true
	}
	for ts := range tsSet {
		l := lab{}
		for _, v := range ivs {
			if ts >= v.from && ts <= v.to {
				if v.auto {
					l.auto++
				} else {
					l.real++
				}
			}
		}
		labs[ts] = l
	}

	var realOnly, mixed, autoOnly, empty []srv
	for _, s := range all {
		l := labs[s.TS]
		switch {
		case l.real > 0 && l.auto == 0:
			realOnly = append(realOnly, s)
		case l.real > 0 && l.auto > 0:
			mixed = append(mixed, s)
		case l.auto > 0:
			autoOnly = append(autoOnly, s)
		default:
			empty = append(empty, s)
		}
	}

	fmt.Printf("# TICK SERVER — người chơi thật\n\n")
	fmt.Printf("DB: %s · mẫu server: %d dòng (mỗi dòng = 1 khu × 10 s)\n", *db_, len(all))
	fmt.Printf("Khoảng: %s → %s\n\n",
		time.Unix(all[0].TS, 0).In(vn).Format("2006-01-02 15:04"),
		time.Unix(all[len(all)-1].TS, 0).In(vn).Format("2006-01-02 15:04"))
	fmt.Printf("Phân loại mốc thời gian theo ai đang nối:\n")
	fmt.Printf("  chỉ NGƯỜI THẬT : %6d dòng\n", len(realOnly))
	fmt.Printf("  người + máy    : %6d dòng\n", len(mixed))
	fmt.Printf("  chỉ MÁY (bài đo): %6d dòng\n", len(autoOnly))
	fmt.Printf("  không ai nối   : %6d dòng\n\n", len(empty))

	show := func(name string, rs []srv) {
		if len(rs) == 0 {
			fmt.Printf("%-22s  (không có dòng nào)\n", name)
			return
		}
		var a, p []float64
		over33, over35 := 0, 0
		for _, s := range rs {
			a = append(a, s.Avg)
			p = append(p, s.P95)
			if s.P95 > 33.3 {
				over33++
			}
			if s.P95 > 3.5 {
				over35++
			}
		}
		fmt.Printf("%-22s n=%-6d avg TB %.2f · p95: TB %.2f  trung vị %.2f  p95-của-p95 %.2f  max %.2f  | >3,5ms %d (%.1f%%) · >33,3ms %d\n",
			name, len(rs), mean(a), mean(p), pctl(p, 0.5), pctl(p, 0.95), maxf(p), over35, 100*float64(over35)/float64(len(rs)), over33)
	}

	fmt.Printf("## 1. Tổng quan tick (ms), theo nhóm\n\n")
	show("CHỈ NGƯỜI THẬT", realOnly)
	show("người + máy", mixed)
	show("chỉ máy (bài đo)", autoOnly)
	show("không ai nối", empty)
	fmt.Println()

	human := append(append([]srv{}, realOnly...), mixed...)

	// Chỉ khu CÓ người thật trong khu đó (humans > 0) — khu rỗng thì tick gần 0, gộp vào làm loãng.
	var occupied []srv
	for _, s := range human {
		if s.Humans > 0 {
			occupied = append(occupied, s)
		}
	}
	fmt.Printf("## 2. Chỉ những KHU ĐANG CÓ NGƯỜI (humans>0), lúc có người thật online\n\n")
	show("khu có người", occupied)
	fmt.Println()

	// Theo khu.
	fmt.Printf("## 3. Theo khu (chỉ dòng có người thật online, humans>0)\n\n")
	byZone := map[int][]srv{}
	for _, s := range occupied {
		byZone[s.Zone] = append(byZone[s.Zone], s)
	}
	var zs []int
	for z := range byZone {
		zs = append(zs, z)
	}
	sort.Ints(zs)
	for _, z := range zs {
		show(fmt.Sprintf("khu %d %s", z, byZone[z][0].ZoneName), byZone[z])
	}
	fmt.Println()

	// Theo số nhân vật trong khu (người + bot).
	fmt.Printf("## 4. Theo số nhân vật trong khu (người + bot), lúc có người thật online\n\n")
	buckets := []struct {
		name   string
		lo, hi int
	}{{"0", 0, 0}, {"1–5", 1, 5}, {"6–10", 6, 10}, {"11–20", 11, 20}, {"21–30", 21, 30}, {"31+", 31, 1 << 30}}
	for _, b := range buckets {
		var rs []srv
		for _, s := range human {
			n := s.Humans + s.Bots
			if n >= b.lo && n <= b.hi {
				rs = append(rs, s)
			}
		}
		show("nhân vật "+b.name, rs)
	}
	fmt.Println()

	// Theo số NGƯỜI trong khu.
	fmt.Printf("## 5. Theo số NGƯỜI trong khu\n\n")
	for _, b := range []struct {
		name   string
		lo, hi int
	}{{"1 người", 1, 1}, {"2–3", 2, 3}, {"4–6", 4, 6}, {"7+", 7, 1 << 30}} {
		var rs []srv
		for _, s := range human {
			if s.Humans >= b.lo && s.Humans <= b.hi {
				rs = append(rs, s)
			}
		}
		show(b.name, rs)
	}
	fmt.Println()

	// Theo build.
	fmt.Printf("## 6. Theo build (chỉ dòng có người thật online, humans>0)\n\n")
	byBuild := map[string][]srv{}
	for _, s := range occupied {
		byBuild[s.Build] = append(byBuild[s.Build], s)
	}
	var bs []string
	for b := range byBuild {
		bs = append(bs, b)
	}
	sort.Strings(bs)
	for _, b := range bs {
		n := b
		if n == "" {
			n = "(trống)"
		}
		show(n, byBuild[b])
	}
	fmt.Println()

	// Theo ngày.
	fmt.Printf("## 7. Theo ngày (chỉ dòng có người thật online, humans>0)\n\n")
	byDay := map[string][]srv{}
	for _, s := range occupied {
		d := time.Unix(s.TS, 0).In(vn).Format("2006-01-02")
		byDay[d] = append(byDay[d], s)
	}
	var ds []string
	for d := range byDay {
		ds = append(ds, d)
	}
	sort.Strings(ds)
	for _, d := range ds {
		show(d, byDay[d])
	}
	fmt.Println()

	// 20 tick xấu nhất lúc có người thật.
	fmt.Printf("## 8. 20 mốc tick p95 CAO NHẤT lúc có người thật online\n\n")
	sort.Slice(human, func(i, j int) bool { return human[i].P95 > human[j].P95 })
	fmt.Printf("%-17s %-6s %-18s %7s %7s %6s %5s %7s %7s %s\n",
		"thời điểm", "khu", "tên khu", "avg", "p95", "người", "bot", "KB/s", "CPU%", "RSS MB")
	for i, s := range human {
		if i >= 20 {
			break
		}
		fmt.Printf("%-17s %-6d %-18s %7.2f %7.2f %6d %5d %7.1f %7.1f %7.0f\n",
			time.Unix(s.TS, 0).In(vn).Format("01-02 15:04:05"), s.Zone, s.ZoneName, s.Avg, s.P95, s.Humans, s.Bots, s.Kbps, s.CPU, s.RSS)
	}
	fmt.Println()

	// CPU / RSS lúc có người thật.
	fmt.Printf("## 9. CPU và RAM tiến trình lúc có người thật online\n\n")
	var cpu, rss []float64
	seen := map[int64]bool{}
	for _, s := range human {
		if seen[s.TS] {
			continue
		}
		seen[s.TS] = true
		cpu = append(cpu, s.CPU)
		rss = append(rss, s.RSS)
	}
	fmt.Printf("CPU%%  TB %.1f · trung vị %.1f · p95 %.1f · max %.1f  (n=%d mốc)\n", mean(cpu), pctl(cpu, 0.5), pctl(cpu, 0.95), maxf(cpu), len(cpu))
	fmt.Printf("RSS   TB %.0f MB · trung vị %.0f · p95 %.0f · max %.0f MB\n\n", mean(rss), pctl(rss, 0.5), pctl(rss, 0.95), maxf(rss))

	// 10. Ghép client ↔ server: lúc client tụt FPS thì tick bao nhiêu.
	fmt.Printf("## 10. Lúc CLIENT thật tụt FPS thì tick server bao nhiêu\n\n")
	type cs struct {
		ts            int64
		fps, p1, ping float64
		zone          int
	}
	var low []cs
	crows, err := db.Query(`SELECT s.server_ts,s.fps_avg,s.fps_p1,s.ping_p95,COALESCE(s.zone,0)
		FROM samples s JOIN sessions ss ON ss.session_id=s.session_id
		WHERE COALESCE(ss.automated,0)=0 AND s.fps_avg IS NOT NULL`)
	must(err)
	var allFPS []cs
	for crows.Next() {
		var c cs
		must(crows.Scan(&c.ts, &c.fps, &c.p1, &c.ping, &c.zone))
		allFPS = append(allFPS, c)
		if c.fps < 45 {
			low = append(low, c)
		}
	}
	crows.Close()
	// tick p95 lớn nhất trong cùng khu ±10 s
	tickAt := func(ts int64, zone int) (float64, bool) {
		best, ok := 0.0, false
		for _, s := range all {
			if s.Zone == zone && s.TS >= ts-10 && s.TS <= ts+10 {
				if s.P95 > best {
					best = s.P95
				}
				ok = true
			}
		}
		return best, ok
	}
	var tickLow, tickAllv []float64
	for _, c := range low {
		if v, ok := tickAt(c.ts, c.zone); ok {
			tickLow = append(tickLow, v)
		}
	}
	for i, c := range allFPS {
		if i%5 != 0 {
			continue // lấy mẫu thưa cho nhanh
		}
		if v, ok := tickAt(c.ts, c.zone); ok {
			tickAllv = append(tickAllv, v)
		}
	}
	fmt.Printf("mẫu client thật: %d · trong đó FPS < 45: %d (%.1f%%)\n", len(allFPS), len(low), 100*float64(len(low))/math.Max(1, float64(len(allFPS))))
	fmt.Printf("tick p95 server CÙNG LÚC client tụt FPS<45 : TB %.2f ms · trung vị %.2f · max %.2f (n=%d)\n", mean(tickLow), pctl(tickLow, 0.5), maxf(tickLow), len(tickLow))
	fmt.Printf("tick p95 server ở mẫu client BẤT KỲ        : TB %.2f ms · trung vị %.2f · max %.2f (n=%d)\n", mean(tickAllv), pctl(tickAllv, 0.5), maxf(tickAllv), len(tickAllv))

	// 11. Số phiên người thật, theo ngày — để biết mẫu dày hay mỏng.
	fmt.Printf("\n## 11. Phiên người thật theo ngày\n\n")
	srows, err := db.Query(`SELECT started_at, COALESCE(dur_secs,0) FROM sessions WHERE COALESCE(automated,0)=0`)
	must(err)
	cnt := map[string]int{}
	dur := map[string]float64{}
	for srows.Next() {
		var st int64
		var d float64
		must(srows.Scan(&st, &d))
		k := time.Unix(st, 0).In(vn).Format("2006-01-02")
		cnt[k]++
		dur[k] += d
	}
	srows.Close()
	var ks []string
	for k := range cnt {
		ks = append(ks, k)
	}
	sort.Strings(ks)
	for _, k := range ks {
		fmt.Printf("%s  %3d phiên · %.2f giờ (tổng dur_secs đã có)\n", k, cnt[k], dur[k]/3600)
	}

	checks(db, all, vn)
	humanTS := map[int64]bool{}
	for _, s := range human {
		humanTS[s.TS] = true
	}
	controlled(all, humanTS, vn)
	autoWin := func(ts int64) bool {
		for _, v := range ivs {
			if v.auto && ts >= v.from-300 && ts <= v.to+300 {
				return true
			}
		}
		return false
	}
	quiet(all, humanTS, autoWin, vn)
	cpuVsTick(all, humanTS)
	idleZone(all, vn)
	ratio(all, vn)
	spread(all, humanTS)
	overlap(all, vn)
	recent(all, time.Now().Add(-26*time.Hour).Unix(), vn)
	if *wins != "" {
		fmt.Printf("## 21. So hai khoảng thời gian chỉ định\n\n")
		for _, w := range strings.Split(*wins, ",") {
			from, to, err := parseWindow(strings.TrimSpace(w), vn)
			if err != nil {
				fmt.Printf("%-28s LỖI: %v\n", w, err)
				continue
			}
			window(all, from, to, strings.TrimSpace(w), vn)
		}
		fmt.Println()
	}
}

// parseWindow đọc "2006-01-02 15:04..15:04" — hai mốc giờ trong CÙNG một ngày.
func parseWindow(s string, vn *time.Location) (int64, int64, error) {
	day, rest, ok := strings.Cut(s, " ")
	if !ok {
		return 0, 0, fmt.Errorf(`cần dạng "2026-11-12 09:00..09:31"`)
	}
	a, b, ok := strings.Cut(rest, "..")
	if !ok {
		return 0, 0, fmt.Errorf(`thiếu ".." giữa hai mốc giờ`)
	}
	from, err := time.ParseInLocation("2006-01-02 15:04", day+" "+a, vn)
	if err != nil {
		return 0, 0, err
	}
	to, err := time.ParseInLocation("2006-01-02 15:04", day+" "+b, vn)
	if err != nil {
		return 0, 0, err
	}
	return from.Unix(), to.Unix(), nil
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
