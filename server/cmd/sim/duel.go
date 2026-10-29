package main

import (
	"fmt"
	"math"
	"math/rand"
	"sort"

	"kiemxoay/server/internal/bot"
	"kiemxoay/server/internal/config"
	"kiemxoay/server/internal/game"
	"kiemxoay/server/internal/world"
)

// Đấu từng cặp nhân vật để đo cân bằng kỹ năng (spec §6.3, CHANGES_v5 mục 14).
//
//	cd server && go run ./cmd/sim --duel
//
// Mỗi cặp đánh `rounds` trận trên bản đồ trống, hai bên cùng cảnh giới, cùng số kiếm, **mỗi
// bên 2 lượt kỹ năng** như spec yêu cầu. Thắng = bên còn kiếm nhiều hơn khi hết giờ, hoặc
// bên còn sống nếu có người chết.
//
// Bản đồ TRỐNG là cố ý: thêm vật cản vào thì kết quả trộn lẫn giữa "kỹ năng mạnh" và "ai
// may mắn đứng sau tảng đá", mà ta đang cần đo đúng một thứ.

type duelResult struct {
	char  string
	wins  int
	games int
}

func runDuels(cfg *config.Config, seed int64, rounds int, secs float64) {
	chars := append([]string(nil), cfg.Characters...)
	sort.Strings(chars)
	// win[a][b] = số trận a thắng b.
	win := map[string]map[string]int{}
	for _, a := range chars {
		win[a] = map[string]int{}
	}

	for i, a := range chars {
		for j, b := range chars {
			if j <= i {
				continue
			}
			for r := 0; r < rounds; r++ {
				switch duelOnce(cfg, a, b, seed+int64(i*100+j*10+r), secs) {
				case 1:
					win[a][b]++
				case -1:
					win[b][a]++
				}
			}
		}
	}

	// Bảng chéo 10×10.
	fmt.Printf("Đấu từng cặp — mỗi cặp %d trận, mỗi bên 2 lượt kỹ năng, %.0f giây/trận\n\n", rounds, secs)
	fmt.Printf("%-14s", "")
	for _, b := range chars {
		fmt.Printf("%6s", short(b))
	}
	fmt.Println("   tỉ lệ thắng")
	totals := make([]duelResult, 0, len(chars))
	for _, a := range chars {
		fmt.Printf("%-14s", short(a))
		w, n := 0, 0
		for _, b := range chars {
			if a == b {
				fmt.Printf("%6s", "—")
				continue
			}
			games := rounds
			fmt.Printf("%6d", win[a][b])
			w += win[a][b]
			n += games
		}
		rate := 0.0
		if n > 0 {
			rate = float64(w) / float64(n) * 100
		}
		flag := ""
		if rate > 60 {
			flag = "  ← > 60%, spec nói phải giảm"
		}
		fmt.Printf("   %5.1f%%%s\n", rate, flag)
		totals = append(totals, duelResult{char: a, wins: w, games: n})
	}

	sort.Slice(totals, func(i, j int) bool {
		return float64(totals[i].wins)/math.Max(float64(totals[i].games), 1) >
			float64(totals[j].wins)/math.Max(float64(totals[j].games), 1)
	})
	fmt.Println("\nXếp theo tỉ lệ thắng:")
	for _, t := range totals {
		fmt.Printf("  %-16s %5.1f%% (%d/%d)\n", t.char, float64(t.wins)/math.Max(float64(t.games), 1)*100, t.wins, t.games)
	}
}

// duelOnce chạy một trận. Trả 1 nếu a thắng, -1 nếu b thắng, 0 nếu hoà.
func duelOnce(cfg *config.Config, charA, charB string, seed int64, secs float64) int {
	c := *cfg
	c.Room.MonsterCount, c.Items.CountPerRoom = 0, 0
	g := game.New(&c, world.EmptyMap(c.Room.MapSize), seed)
	rng := rand.New(rand.NewSource(seed))

	mid := c.Room.MapSize / 2
	level := c.Cultivation.TiersPerRealm * 3 // Kim Đan: đủ kiếm để có chuyện xảy ra
	swords := g.Prog.MaxSwords(g.Prog.Realm(level)) / 2

	mk := func(char string, dx float64) (*game.Player, *bot.Brain) {
		p := g.AddPlayer(game.SpawnOptions{Name: char, Char: char, IsBot: true, Level: level, Swords: swords})
		p.Pos = game.Vec{X: mid + dx, Y: mid}
		p.InvulnUntil = 0
		// Mỗi bên đúng 2 lượt kỹ năng (spec v5 mục 14).
		for k := 0; k < c.Skills.MaxCharges; k++ {
			g.AddSkillCharge(p)
		}
		return p, bot.New(p.ID, rng.Int63(), g)
	}
	pa, ba := mk(charA, -600)
	pb, bb := mk(charB, 600)

	dt := 1 / float64(c.Net.TickHz)
	for i := 0; i < int(secs/dt); i++ {
		ba.Update(g)
		bb.Update(g)
		g.Tick(dt)
		g.DrainEvents()
		if pa.Dead || pb.Dead {
			break
		}
	}
	switch {
	case pb.Dead && !pa.Dead:
		return 1
	case pa.Dead && !pb.Dead:
		return -1
	case pa.Ring.Count > pb.Ring.Count:
		return 1
	case pb.Ring.Count > pa.Ring.Count:
		return -1
	}
	return 0
}

// short cắt tên nhân vật cho vừa cột bảng.
func short(s string) string {
	if len(s) <= 5 {
		return s
	}
	return s[:5]
}
