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

// castCount đếm số lần mỗi nhân vật thật sự bấm được kỹ năng trong cả loạt đấu.
//
// Cần con số này để phân biệt hai chuyện khác hẳn nhau: "kỹ năng yếu" và "bot không bao giờ
// bấm". Nhìn bảng tỉ lệ thắng thì hai thứ đó giống hệt nhau.
var castCount map[string]int

// duelSwords là số kiếm mỗi bên khi vào trận; 0 = đầy trần cảnh giới. Xem ghi chú trong
// `duelOnce` về việc vì sao con số này đổi hẳn kết luận của bảng.
var duelSwords int

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
	draws, played := 0, 0
	castCount = map[string]int{}

	for i, a := range chars {
		for j, b := range chars {
			if j <= i {
				continue
			}
			for r := 0; r < rounds; r++ {
				played++
				switch duelOnce(cfg, a, b, seed+int64(i*100+j*10+r), secs) {
				case 1:
					win[a][b]++
				case -1:
					win[b][a]++
				default:
					draws++
				}
			}
		}
	}

	// Bảng chéo 10×10.
	swDesc := fmt.Sprintf("%d kiếm", duelSwords)
	if duelSwords <= 0 {
		swDesc = "đầy trần cảnh giới"
	}
	fmt.Printf("Đấu từng cặp — mỗi cặp %d trận, mỗi bên 2 lượt kỹ năng, %s, tối đa %.0f giây/trận\n",
		rounds, swDesc, secs)
	// Nói rõ tỉ lệ hoà: hai bên cùng cảnh giới cùng số kiếm thì mài nhau rất đều, và nếu
	// phần lớn trận hoà thì bảng tỉ lệ thắng bên dưới gần như vô nghĩa — phải thấy điều đó
	// chứ không được đọc nó như một kết luận.
	fmt.Printf("%d trận, %d hoà (%.0f%%)\n\n", played, draws, float64(draws)/math.Max(float64(played), 1)*100)
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
	fmt.Println("\nSố lần thật sự bấm được kỹ năng (mỗi bên có 2 lượt, tổng 144 lượt khả dụng/nhân vật):")
	for _, a := range chars {
		fmt.Printf("  %-16s %d\n", a, castCount[a])
	}
	reportCounters(chars, win, rounds)
}

// reportCounters đọc bảng theo kiểu **khắc chế lẫn nhau**, không theo kiểu ai mạnh ai yếu.
//
// Mục tiêu cân bằng không phải "ai cũng 50%", mà là mỗi nhân vật đều có người khắc mình và
// có người mình khắc — kéo búa bao. Một bảng toàn 50% mà không ai khắc ai thì chọn nhân vật
// nào cũng như nhau, và cả hệ kỹ năng thành trang trí.
func reportCounters(chars []string, win map[string]map[string]int, rounds int) {
	const strong, weak = 0.65, 0.35
	fmt.Println("\nKhắc chế (thắng ≥ 65% một đối thủ = khắc; ≤ 35% = bị khắc):")
	flat, bully, punching := 0, []string{}, []string{}
	for _, a := range chars {
		var beats, loses []string
		for _, b := range chars {
			if a == b {
				continue
			}
			rate := float64(win[a][b]) / float64(rounds)
			switch {
			case rate >= strong:
				beats = append(beats, short(b))
			case rate <= weak:
				loses = append(loses, short(b))
			}
		}
		if len(beats) == 0 && len(loses) == 0 {
			flat++
		}
		if len(loses) == 0 && len(beats) > 0 {
			bully = append(bully, a)
		}
		if len(beats) == 0 && len(loses) > 0 {
			punching = append(punching, a)
		}
		fmt.Printf("  %-16s khắc: %-28s bị khắc bởi: %s\n", a, joinOr(beats, "(không ai)"), joinOr(loses, "(không ai)"))
	}
	fmt.Println()
	if len(bully) > 0 {
		fmt.Printf("  ⚠ không ai khắc được: %v — cần giảm\n", bully)
	}
	if len(punching) > 0 {
		fmt.Printf("  ⚠ không khắc được ai: %v — cần tăng\n", punching)
	}
	if flat > 0 {
		fmt.Printf("  ⚠ %d nhân vật không khắc ai và cũng không bị ai khắc — chọn ai cũng như nhau\n", flat)
	}
	if len(bully) == 0 && len(punching) == 0 && flat == 0 {
		fmt.Println("  ✓ mọi nhân vật đều có người khắc mình và người mình khắc")
	}
}

func joinOr(xs []string, empty string) string {
	if len(xs) == 0 {
		return empty
	}
	out := ""
	for i, x := range xs {
		if i > 0 {
			out += " "
		}
		out += x
	}
	return out
}

// duelOnce chạy một trận. Trả 1 nếu a thắng, -1 nếu b thắng, 0 nếu hoà.
func duelOnce(cfg *config.Config, charA, charB string, seed int64, secs float64) int {
	c := *cfg
	c.Room.MonsterCount, c.Items.CountPerRoom = 0, 0
	g := game.New(&c, world.EmptyMap(c.Room.MapSize), seed)
	rng := rand.New(rand.NewSource(seed))

	mid := c.Room.MapSize / 2
	level := c.Cultivation.TiersPerRealm * 3 // Kim Đan: đủ kiếm để có chuyện xảy ra
	// Mặc định **đầy trần cảnh giới**, không phải một vòng.
	//
	// Bản trước cho mỗi bên đúng một vòng (10 kiếm) vì lúc ấy đo được 2 vòng trở lên là 85%
	// trận hoà. Nhưng con số đó đo TRƯỚC khi `clash_cooldown` được sửa về 0 (mài nhanh gấp
	// 9 lần); đo lại ngày 29/10 thì đầy kiếm chỉ **10% hoà**, còn ÍT hơn một vòng (18%).
	//
	// Và một vòng ở Kim Đan là 10 kiếm trên trần 100, tức **1/10 số kiếm thật** của người
	// chơi cùng cảnh giới — nên nó thổi phồng gấp mười mọi kỹ năng tính theo TRẦN cảnh giới.
	// Bánh Bao Thần (v5.5: hồi 20% trần = 20 kiếm) ra **100% thắng** với một vòng, và 67%
	// với đủ kiếm. Con số thứ hai mới là thứ người chơi gặp.
	//
	// `--duel-swords N` để quay lại kiểu cũ.
	swords := duelSwords
	if swords <= 0 {
		swords = g.Prog.MaxSwords(g.Prog.Realm(level))
	}

	mk := func(char string, dx float64) (*game.Player, *bot.Brain) {
		p := g.AddPlayer(game.SpawnOptions{Name: char, Char: char, IsBot: true, Level: level, Swords: swords})
		p.Pos = game.Vec{X: mid + dx, Y: mid}
		p.InvulnUntil = 0
		// Mỗi bên đúng 2 lượt kỹ năng (spec v5 mục 14).
		for k := 0; k < c.Skills.MaxCharges; k++ {
			g.AddSkillCharge(p)
		}
		return p, bot.NewFighter(p.ID, rng.Int63(), g)
	}
	pa, ba := mk(charA, -600)
	pb, bb := mk(charB, 600)

	dt := 1 / float64(c.Net.TickHz)
	for i := 0; i < int(secs/dt); i++ {
		ba.Update(g)
		bb.Update(g)
		g.Tick(dt)
		for _, e := range g.DrainEvents() {
			if e.Kind == "SkillCast" && castCount != nil {
				castCount[e.S]++
			}
		}
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
