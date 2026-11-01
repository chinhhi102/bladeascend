package main

// --duel: đấu từng cặp nhân vật — SỐ THAM KHẢO cho báo cáo (spec v6.30 §6.3 bỏ hàng rào
// tỉ lệ thắng tay đôi; cân bằng theo số liệu người chơi thật, xem /api/admin/characters).
//
//	cd server && go run ./cmd/sim --duel [--rounds 20] [--secs 60]
//
// Mỗi cặp đánh `rounds` trận trên bản đồ TRỐNG (để kết quả là "kỹ năng nào mạnh", không trộn
// với "ai may đứng sau tảng đá"), hai bên cùng cảnh giới, **đủ trần kiếm**, kỹ năng **hồi theo
// bảng** (vào trận 1 lượt, tự hồi — không cấp sẵn lượt). Đổi bên trái/phải mỗi trận.
//
// Thắng = chém trúng thân trước (v6.26: chém trúng là chết). Hết giờ chưa ai trúng → bên còn
// nhiều kiếm hơn thắng; bằng nhau → hoà. Luôn thoát mã 0.
//
// Lịch sử: bỏ ở Q25, v6.25 đưa lại làm hàng rào 35–65%, v6.30 (Q36) hạ xuống số tham khảo.

import (
	"fmt"
	"math"
	"math/rand"
	"sort"
	"sync"

	"bladeascend/server/internal/bot"
	"bladeascend/server/internal/config"
	"bladeascend/server/internal/game"
	"bladeascend/server/internal/world"
)

// duelLo / duelHi: chỉ để tô ✅/⚠ trong bảng tham khảo (không chặn gì).
const (
	duelLo = 0.35
	duelHi = 0.65
)

type duelOpts struct {
	rounds int
	secs   float64
	seed   int64
	realm  int // cảnh giới hai bên (mặc định Kim Đan)
}

// duelTally: kết quả cả loạt — win[a][b] = số trận a thắng b; casts = số lần bấm kỹ năng.
type duelTally struct {
	mu     sync.Mutex
	win    map[string]map[string]int
	draws  int
	played int
	casts  map[string]int
}

func (t *duelTally) add(a, b string, res int, casts map[string]int) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.played++
	switch res {
	case 1:
		t.win[a][b]++
	case -1:
		t.win[b][a]++
	default:
		t.draws++
	}
	for k, v := range casts {
		t.casts[k] += v
	}
}

func runDuels(cfg *config.Config, o duelOpts) *duelTally {
	chars := append([]string(nil), cfg.Characters...)
	sort.Strings(chars)
	t := &duelTally{win: map[string]map[string]int{}, casts: map[string]int{}}
	for _, a := range chars {
		t.win[a] = map[string]int{}
	}
	var wg sync.WaitGroup
	sem := make(chan struct{}, 8)
	for i, a := range chars {
		for j := i + 1; j < len(chars); j++ {
			b := chars[j]
			for r := 0; r < o.rounds; r++ {
				wg.Add(1)
				sem <- struct{}{}
				go func(i, j, r int, a, b string) {
					defer func() { <-sem; wg.Done() }()
					seed := o.seed + int64(i*10000+j*100+r)
					res, casts := duelOnce(cfg, a, b, seed, o, r%2 == 1)
					t.add(a, b, res, casts)
				}(i, j, r, a, b)
			}
		}
	}
	wg.Wait()
	return t
}

// duelOnce: 1 nếu a thắng, -1 nếu b thắng, 0 nếu hoà. `swap` đổi bên đứng.
func duelOnce(cfg *config.Config, charA, charB string, seed int64, o duelOpts, swap bool) (int, map[string]int) {
	c := *cfg
	c.Room.MonsterCount, c.Items.CountPerRoom = 0, 0
	g := game.New(&c, world.EmptyMap(c.Room.MapSize), seed, 0)
	rng := rand.New(rand.NewSource(seed))
	mid := c.Room.MapSize / 2
	level := o.realm * g.Prog.Tiers()
	swords := g.Prog.MaxSwords(o.realm)
	side := 600.0
	if swap {
		side = -side
	}
	mk := func(char string, dx float64) (*game.Player, *bot.Brain) {
		p := g.AddPlayer(game.SpawnOptions{Name: char, Char: char, IsBot: true, Level: level, Swords: swords})
		p.Pos = game.Vec{X: mid + dx, Y: mid}
		p.InvulnUntil = 0
		return p, bot.NewFighter(p.ID, rng.Int63(), g)
	}
	pa, ba := mk(charA, -side)
	pb, bb := mk(charB, side)
	casts := map[string]int{}
	dt := 1 / float64(c.Net.TickHz)
	for i := 0; i < int(o.secs/dt); i++ {
		ba.Update(g)
		bb.Update(g)
		g.Tick(dt)
		for _, e := range g.DrainEvents() {
			switch {
			case e.Kind == "SkillCast":
				casts[e.S]++
			case e.Kind == "Death" && e.A == pa.ID && e.B == pb.ID:
				return 1, casts
			case e.Kind == "Death" && e.A == pb.ID && e.B == pa.ID:
				return -1, casts
			}
		}
		if pa.Dead != pb.Dead { // chết vì đường khác (đàn em không chém thân được, phòng hờ)
			return map[bool]int{true: -1, false: 1}[pa.Dead], casts
		}
	}
	switch {
	case pa.Ring.Count > pb.Ring.Count:
		return 1, casts
	case pb.Ring.Count > pa.Ring.Count:
		return -1, casts
	default:
		return 0, casts
	}
}

// duelMain in bảng 10×10 + tỉ lệ thắng + hồi chiêu đang dùng; trả false khi trượt hàng rào.
func duelMain(cfg *config.Config, o duelOpts) bool {
	t := runDuels(cfg, o)
	chars := append([]string(nil), cfg.Characters...)
	sort.Strings(chars)
	fmt.Printf("## cmd/sim --duel · %s · đủ trần kiếm · %d trận/cặp · tối đa %.0f s/trận · seed %d\n\n",
		cfg.Cultivation.Realms[o.realm], o.rounds, o.secs, o.seed)
	fmt.Printf("%d trận, %d hoà (%.0f%%)\n\n", t.played, t.draws, 100*float64(t.draws)/math.Max(float64(t.played), 1))
	fmt.Print("| thắng ↓ / thua → |")
	for _, b := range chars {
		fmt.Printf(" %s |", short(b))
	}
	fmt.Println(" Tỉ lệ thắng | Hồi · lượt | Bấm kỹ năng | 35–65% (tham khảo) |")
	fmt.Println("|" + repeat("---|", len(chars)+5))
	ok := true
	for _, a := range chars {
		fmt.Printf("| %s |", short(a))
		w, n := 0, 0
		for _, b := range chars {
			if a == b {
				fmt.Print(" — |")
				continue
			}
			fmt.Printf(" %d |", t.win[a][b])
			w += t.win[a][b]
			n += t.win[a][b] + t.win[b][a]
		}
		rate := float64(w) / math.Max(float64(n), 1)
		pass := rate >= duelLo && rate <= duelHi
		ok = ok && pass
		cd := cfg.Skills.Cooldowns[a]
		fmt.Printf(" %.0f%% | %.0fs · %.0f | %d | %s |\n", rate*100, cd[0], cd[1], t.casts[a], mark(pass))
	}
	fmt.Println("\n(Tỉ lệ thắng tính trên trận có kết quả, không tính hoà.)")
	return ok
}

func repeat(s string, n int) string {
	out := ""
	for i := 0; i < n; i++ {
		out += s
	}
	return out
}

// short cắt tên nhân vật cho vừa cột bảng.
func short(s string) string {
	if len(s) <= 6 {
		return s
	}
	return s[:6]
}
