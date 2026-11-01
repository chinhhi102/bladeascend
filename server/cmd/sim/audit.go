package main

// --audit: đo EXP theo NGUỒN cho docs/EXP_AUDIT.md (CHANGES v6.21, bước 1).
//
// Hai phép đo, đều là farm thuần (không ai chém được ai):
//
//  1. Tốc độ: một người farm (não bot) GHIM ở Luyện Khí 1 / Luyện Hư 1 / Đại Thừa 1, có và
//     không Nam Châm. Ghim nghĩa là lên cấp thì đặt lại về đúng cấp đó — để con số là EXP/s
//     của CẢNH GIỚI đó, không bị trộn với cảnh giới kế tiếp.
//  2. Tái hiện lỗi 1/11: bắt đầu Luyện Hư 1, không ghim, đo thời gian tới từng cảnh giới,
//     tới Đại Thừa tầng 10 và tới Chân Tiên.
//
// Không đổi luật nào: chỉ gắn `Game.ExpTrace` để biết từng lượng EXP từ đâu ra.

import (
	"fmt"
	"math/rand"
	"sort"
	"strings"

	"bladeascend/server/internal/bot"
	"bladeascend/server/internal/config"
	"bladeascend/server/internal/game"
)

// auditOpts: một lượt đo.
type auditOpts struct {
	realm    int
	magnet   bool
	pin      bool
	seconds  float64
	seed     int64
	mapID    string
	rivals   int
	stopTien bool
}

// auditResult: EXP theo nguồn + mốc thời gian (lần đầu chạm từng cảnh giới).
type auditResult struct {
	bySrc     [game.ExpSourceCount]float64
	events    [game.ExpSourceCount]int
	reached   map[string]float64
	simulated float64
	prog      *game.Progression
}

func runAudit(cfg *config.Config, o auditOpts) auditResult {
	m := pickMap(cfg, o.mapID, o.seed)
	g := game.New(cfg, m, o.seed, 0)
	rng := rand.New(rand.NewSource(o.seed))
	startL := o.realm * g.Prog.Tiers()
	me := g.AddPlayer(game.SpawnOptions{Name: "audit", Char: cfg.Characters[0],
		Level: startL, Swords: g.Prog.MaxSwords(o.realm)})
	brains := []*bot.Brain{bot.New(me.ID, rng.Int63(), g)}
	// Đối thủ farm cùng khu (Luyện Khí, như người mới) — để viên trên bản đồ không phải của
	// riêng một mình người được đo.
	for i := 0; i < o.rivals; i++ {
		p := g.AddPlayer(game.SpawnOptions{Name: fmt.Sprintf("rival%d", i),
			Char: cfg.Characters[(i+1)%len(cfg.Characters)], IsBot: true, Swords: 1})
		brains = append(brains, bot.New(p.ID, rng.Int63(), g))
	}
	res := auditResult{reached: map[string]float64{}, prog: g.Prog}
	g.ExpTrace = func(p *game.Player, src game.ExpSource, amount float64) {
		if p.ID == me.ID {
			res.bySrc[src] += amount
			res.events[src]++
		}
	}
	dt := 1 / float64(cfg.Net.TickHz)
	daiThua10 := g.Prog.Threshold(len(cfg.Cultivation.Realms) * g.Prog.Tiers())
	for i := 0; float64(i)*dt < o.seconds; i++ {
		if o.magnet {
			me.Buffs[game.BuffMagnet] = g.Time + 1
		}
		for _, b := range brains {
			b.Update(g)
		}
		g.Tick(dt)
		g.DrainEvents()
		if o.pin && me.Level != startL {
			me.TotalExp = g.Prog.Threshold(startL)
			g.ForceState(me, startL, me.Ring.Count, me.Pos)
		}
		noteReached(g, me, res.reached, daiThua10)
		if o.stopTien && res.reached["Chân Tiên"] > 0 {
			break
		}
	}
	res.simulated = g.Time
	return res
}

// noteReached ghi lần đầu người được đo chạm mỗi cảnh giới, Đại Thừa 10 đầy, Chân Tiên.
func noteReached(g *game.Game, p *game.Player, reached map[string]float64, daiThua10 float64) {
	name := g.Prog.RealmName(g.Prog.Realm(p.Level))
	if _, ok := reached[name]; !ok {
		reached[name] = g.Time
	}
	if _, ok := reached["ĐT10 đầy"]; !ok && p.TotalExp >= daiThua10-0.5 {
		reached["ĐT10 đầy"] = g.Time
	}
}

// auditMain in cả hai phép đo dưới dạng bảng Markdown, dán thẳng vào EXP_AUDIT.md.
func auditMain(cfg *config.Config, mapID string, seed int64, seeds int) {
	cfg.Player.InvulnOnSpawn = 1e9 // farm thuần: không ai chém được ai (như --farm)
	fmt.Println("## Đo 1 — EXP/s theo nguồn, ghim cảnh giới (farm thuần, 60 s × " +
		fmt.Sprint(seeds) + " seed, 19 đối thủ Luyện Khí)")
	fmt.Println()
	fmt.Println("| Cảnh giới | Nam Châm | Tổng EXP/s | " + strings.Join(game.ExpSourceNames[:game.ExpOther], " | ") + " | EXP cần cho 1 cảnh giới | Giây / cảnh giới |")
	fmt.Println("|" + strings.Repeat("---|", int(game.ExpOther)+5))
	for _, r := range []int{0, 5, 8} {
		for _, mag := range []bool{false, true} {
			var sum auditResult
			for s := 0; s < seeds; s++ {
				res := runAudit(cfg, auditOpts{realm: r, magnet: mag, pin: true, seconds: 60,
					seed: seed + int64(s), mapID: mapID, rivals: 19})
				for i := range sum.bySrc {
					sum.bySrc[i] += res.bySrc[i]
					sum.events[i] += res.events[i]
				}
				sum.simulated += res.simulated
				sum.prog = res.prog
			}
			printRateRow(r, mag, sum)
		}
	}
	fmt.Println()
	fmt.Println("## Đo 2 — tái hiện: Luyện Hư 1 → … (farm thuần, không ghim, tối đa 30 phút)")
	fmt.Println()
	for _, mag := range []bool{true, false} {
		for s := 0; s < seeds; s++ {
			res := runAudit(cfg, auditOpts{realm: 5, magnet: mag, seconds: 1800,
				seed: seed + int64(s), mapID: mapID, rivals: 19, stopTien: true})
			printRepro(mag, s, res)
		}
	}
}

func printRateRow(r int, mag bool, sum auditResult) {
	prog := sum.prog
	total := 0.0
	cols := make([]string, 0, game.ExpOther)
	for i := game.ExpSource(0); i < game.ExpOther; i++ {
		total += sum.bySrc[i]
		cols = append(cols, fmt.Sprintf("%.1f", sum.bySrc[i]/sum.simulated))
	}
	perSec := total / sum.simulated
	need := prog.Threshold((r+1)*prog.Tiers()) - prog.Threshold(r*prog.Tiers())
	fmt.Printf("| %s | %s | **%.1f** | %s | %.0f | %s |\n", prog.RealmName(r), onOff(mag), perSec,
		strings.Join(cols, " | "), need, secsFor(need, perSec))
}

func printRepro(mag bool, seed int, res auditResult) {
	type mark struct {
		name string
		t    float64
	}
	marks := make([]mark, 0, len(res.reached))
	for k, v := range res.reached {
		marks = append(marks, mark{k, v})
	}
	sort.Slice(marks, func(i, j int) bool { return marks[i].t < marks[j].t })
	parts := make([]string, 0, len(marks))
	for _, mk := range marks {
		parts = append(parts, fmt.Sprintf("%s %.0fs", mk.name, mk.t))
	}
	total := 0.0
	for _, v := range res.bySrc {
		total += v
	}
	share := func(s game.ExpSource) string {
		if total == 0 {
			return "0%"
		}
		return fmt.Sprintf("%.0f%%", 100*res.bySrc[s]/total)
	}
	fmt.Printf("- Nam Châm %s · seed %d: %s · EXP: viên môi trường %s, quái %s, kiếm dư %s\n",
		onOff(mag), seed, strings.Join(parts, " → "),
		share(game.ExpOrbAmbient), share(game.ExpOrbMonster), share(game.ExpOvercapSword))
}

func secsFor(need, perSec float64) string {
	if perSec <= 0 {
		return "∞"
	}
	return fmt.Sprintf("%.0f", need/perSec)
}

func onOff(b bool) string {
	if b {
		return "có"
	}
	return "không"
}
