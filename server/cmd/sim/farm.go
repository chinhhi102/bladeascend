package main

// --farm: nhịp farm thuần so với bảng chuẩn spec v6.24 §6.0 (Q33-b, phương án A).
//
//	go run ./cmd/sim --farm                    # bảng thường, ±25 % mỗi cảnh giới
//	go run ./cmd/sim --farm --magnet-always    # Nam Châm suốt: không nhanh hơn bảng quá 1,5 lần
//
// Thoát mã 1 khi trượt — CI chạy cả hai (spec: "trượt khỏi khoảng cho phép → build fail").
// Bài thứ ba chạy cùng `--magnet-always`: Luyện Hư 1 → Chân Tiên phải ≥ 15 phút farm thuần.

import (
	"fmt"
	"math/rand"
	"sort"
	"sync"

	"bladeascend/server/internal/bot"
	"bladeascend/server/internal/config"
	"bladeascend/server/internal/game"
)

// farmTarget: phút farm qua từng cảnh giới, bảng §6.0 (Đại Thừa = đủ 10 tầng, đầy thanh).
var farmTarget = []float64{1.3, 1.5, 1.8, 2.1, 2.6, 3.3, 4.1, 5.2, 6.6}

const (
	farmTolerance  = 0.25 // ±25 %
	magnetMaxSpeed = 1.5  // Nam Châm không được nhanh hơn bảng quá 1,5 lần
	tienMinMinutes = 15.0 // Luyện Hư → Chân Tiên có Nam Châm ≥ 15 phút
)

type farmOpts struct {
	bots     int
	minutes  float64
	magnet   bool
	seed     int64
	mapID    string
	startRlm int // cảnh giới khởi đầu của mọi bot (0 = Luyện Khí 1)
	seeds    int // số khu chạy song song (seed liên tiếp), gộp mọi bot lại
}

// farmLife: mốc thời gian (giây kể từ lúc sinh) một bot chạm từng cảnh giới; index 9 =
// Đại Thừa 10 đầy thanh, index 10 = Chân Tiên.
type farmLife struct{ at [11]float64 }

// runFarmSeeds chạy `seeds` khu song song (mỗi khu một goroutine, game không chung trạng
// thái) và gộp mọi bot: một khu 12 bot cho trung vị nhảy ±40 %, bốn khu thì ổn định.
func runFarmSeeds(cfg *config.Config, o farmOpts) []farmLife {
	cfg.Player.InvulnOnSpawn = 1e9
	n := max(1, o.seeds)
	res := make([][]farmLife, n)
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			oi := o
			oi.seed = o.seed + int64(i)
			res[i] = runFarm(cfg, oi)
		}(i)
	}
	wg.Wait()
	var all []farmLife
	for _, r := range res {
		all = append(all, r...)
	}
	return all
}

func runFarm(cfg *config.Config, o farmOpts) []farmLife {
	m := pickMap(cfg, o.mapID, o.seed)
	g := game.New(cfg, m, o.seed, 0)
	rng := rand.New(rand.NewSource(o.seed))
	startL := o.startRlm * g.Prog.Tiers()
	// Slice, không phải map: duyệt map Go theo thứ tự ngẫu nhiên, nên cùng một seed cho hai
	// kết quả khác nhau — bài CI sát ngưỡng sẽ đỏ/xanh tuỳ lượt chạy.
	var brains []*bot.Brain
	lives := map[uint32]*farmLife{}
	for i := 0; i < o.bots; i++ {
		p := g.AddPlayer(game.SpawnOptions{Name: fmt.Sprintf("farm%d", i),
			Char: cfg.Characters[i%len(cfg.Characters)], IsBot: true, Level: startL,
			Swords: max(1, g.Prog.MaxSwords(o.startRlm)*min(o.startRlm, 1))})
		brains = append(brains, bot.NewFarmer(p.ID, rng.Int63(), g))
		lives[p.ID] = &farmLife{}
	}
	dt := 1 / float64(cfg.Net.TickHz)
	realms := len(cfg.Cultivation.Realms)
	for i := 0; float64(i)*dt < o.minutes*60; i++ {
		for _, b := range brains {
			if o.magnet {
				g.Players[b.ID].Buffs[game.BuffMagnet] = g.Time + 1
			}
			b.Update(g)
		}
		g.Tick(dt)
		g.DrainEvents()
		if noteFarm(g, lives, realms) {
			break
		}
	}
	out := make([]farmLife, 0, len(lives))
	for _, l := range lives {
		out = append(out, *l)
	}
	return out
}

// noteFarm ghi mốc mới; true khi mọi bot đã tới Chân Tiên (dừng sớm cho nhanh).
func noteFarm(g *game.Game, lives map[uint32]*farmLife, realms int) bool {
	done := true
	for id, l := range lives {
		p := g.Players[id]
		r := min(g.Prog.Realm(p.Level), realms)
		for k := 1; k <= r && k < realms; k++ {
			if l.at[k] == 0 {
				l.at[k] = g.Time
			}
		}
		if l.at[realms] == 0 && g.BanksTienKhi(p) {
			l.at[realms] = g.Time
		}
		if l.at[realms+1] == 0 && g.TienRealmOf(p) >= 0 {
			l.at[realms+1] = g.Time
		}
		done = done && l.at[realms+1] > 0
	}
	return done
}

// medianDur: trung vị thời gian TỪNG bot đi từ mốc k tới mốc k+1 (giây). ok=false khi chưa
// tới nửa số bot qua được. Lấy trung vị của từng đoạn, không trừ hai trung vị cộng dồn —
// cách đó cho đoạn này 1,5 phút đoạn sau 4,4 phút chỉ vì hai trung vị rơi vào hai bot khác.
func medianDur(lives []farmLife, k int) (float64, bool) {
	var xs []float64
	for _, l := range lives {
		if l.at[k+1] > 0 && (k == 0 || l.at[k] > 0) {
			xs = append(xs, l.at[k+1]-l.at[k])
		}
	}
	if len(xs)*2 < len(lives) {
		return 0, false
	}
	sort.Float64s(xs)
	return xs[len(xs)/2], true
}

// farmMain chạy bài đo, in bảng so với §6.0, trả false nếu trượt.
func farmMain(cfg *config.Config, o farmOpts) bool {
	lives := runFarmSeeds(cfg, o)
	label := "thường"
	if o.magnet {
		label = "Nam Châm suốt"
	}
	fmt.Printf("## cmd/sim --farm (%s) · %d khu × %d bot · seed %d\n\n", label, max(1, o.seeds), o.bots, o.seed)
	fmt.Println("| Cảnh giới | Đo (phút) | Bảng §6.0 | Cho phép | Cộng dồn đo | Bot qua | Kết quả |")
	fmt.Println("|---|---|---|---|---|---|---|")
	ok := true
	cum := 0.0
	for r, want := range farmTarget {
		d, reached := medianDur(lives, r)
		lo, hi := want*(1-farmTolerance), want*(1+farmTolerance)
		if o.magnet {
			lo, hi = want/magnetMaxSpeed, 1e9
		}
		name := cfg.Cultivation.Realms[r]
		if !reached {
			fmt.Printf("| %s | — | %.1f | %s | — | %d/%d | ❌ |\n", name, want, rangeText(lo, hi), countAt(lives, r+1), len(lives))
			ok = false
			continue
		}
		got := d / 60
		cum += got
		pass := got >= lo && got <= hi
		ok = ok && pass
		fmt.Printf("| %s | %.2f | %.1f | %s | %.1f | %d/%d | %s |\n", name, got, want, rangeText(lo, hi), cum, countAt(lives, r+1), len(lives), mark(pass))
	}
	return ok
}

// tienMain: Luyện Hư 1 → Chân Tiên có Nam Châm suốt phải ≥ 15 phút.
func tienMain(cfg *config.Config, o farmOpts) bool {
	o.startRlm, o.magnet = 5, true
	lives := runFarmSeeds(cfg, o)
	// Chân Tiên cần thêm Tiên khí (Đài / Bí Kíp đủ lượt); farm thuần có thể chưa có — khi
	// đó lấy mốc Đại Thừa 10 đầy thanh, sớm hơn nên là phép kiểm CHẶT hơn.
	realms := len(cfg.Cultivation.Realms)
	at, reached := medianSince(lives, 5, realms+1)
	if !reached {
		at, reached = medianSince(lives, 5, realms)
	}
	got := at / 60
	pass := !reached || got >= tienMinMinutes
	fmt.Printf("\nLuyện Hư 1 → Chân Tiên (Nam Châm suốt): %s · cần ≥ %.0f phút %s\n",
		tienText(got, reached, o.minutes), tienMinMinutes, mark(pass))
	return pass
}

// medianSince: trung vị thời gian từ mốc `from` tới mốc `to` của từng bot.
func medianSince(lives []farmLife, from, to int) (float64, bool) {
	var xs []float64
	for _, l := range lives {
		if l.at[to] > 0 {
			xs = append(xs, l.at[to]-l.at[from])
		}
	}
	if len(xs)*2 < len(lives) {
		return 0, false
	}
	sort.Float64s(xs)
	return xs[len(xs)/2], true
}

func tienText(got float64, reached bool, limit float64) string {
	if !reached {
		return fmt.Sprintf("chưa tới sau %.0f phút", limit)
	}
	return fmt.Sprintf("%.1f phút", got)
}

func rangeText(lo, hi float64) string {
	if hi >= 1e8 {
		return fmt.Sprintf("≥ %.2f", lo)
	}
	return fmt.Sprintf("%.2f – %.2f", lo, hi)
}

func mark(ok bool) string {
	if ok {
		return "✅"
	}
	return "❌"
}

func countAt(lives []farmLife, k int) int {
	n := 0
	for _, l := range lives {
		if l.at[k] > 0 {
			n++
		}
	}
	return n
}
