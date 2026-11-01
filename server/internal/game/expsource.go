package game

// ExpSource: EXP này từ đâu ra (CHANGES v6.21 — EXP_AUDIT). Dùng để đo, không đổi luật nào.
//
// Viên EXP nằm trên đất mang theo nguồn của nó (`Item.Src`): viên môi trường, viên rơi từ
// quái, viên văng ra khi chém người đều là `orb` giống hệt nhau lúc nhặt, nên phải ghi nguồn
// lúc SINH viên thì sau mới tách được.
type ExpSource uint8

const (
	ExpOrbAmbient    ExpSource = iota // viên môi trường §6.1 (giá trị 0 = mặc định của Item)
	ExpOrbMonster                     // viên rơi khi hạ quái §14
	ExpOrbHitDrop                     // viên văng ra khi chém trúng người (drop_share)
	ExpHitPlayer                      // cướp EXP khi chém người thật (attacker_share)
	ExpHitBot                         // cướp EXP khi chém bot
	ExpKillPlayer                     // hạ người thật ở Luyện Khí (+ kill_bonus_exp)
	ExpKillBot                        // hạ bot ở Luyện Khí
	ExpOvercapSword                   // nhặt kiếm khi đã đủ kiếm tối đa
	ExpDebug                          // phím debug (chỉ server -debug)
	ExpOther
	ExpSourceCount
)

// ExpSourceNames: tên ngắn cho báo cáo (`cmd/sim --audit`).
var ExpSourceNames = [ExpSourceCount]string{
	"orb_ambient", "orb_monster", "orb_hit_drop", "hit_player", "hit_bot",
	"kill_player", "kill_bot", "overcap_sword", "debug", "other",
}

// hitSource / killSource: nguồn khi chém trúng / hạ `b`.
func hitSource(b *Player) ExpSource {
	if b.IsBot {
		return ExpHitBot
	}
	return ExpHitPlayer
}

func killSource(b *Player) ExpSource {
	if b.IsBot {
		return ExpKillBot
	}
	return ExpKillPlayer
}
