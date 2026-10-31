// Package bot drives AI players (spec §10). Bots only use what a player at their position could see.
package bot

import (
	"math"
	"math/rand"

	"kiemxoay/server/internal/game"
	"kiemxoay/server/internal/geom"
)

type Vec = geom.Vec

// States of spec §10.
const (
	StateFarm   = "farm"
	StateHunt   = "hunt"
	StateFlee   = "flee"
	StateAmbush = "ambush"
	StateWander = "wander"
)

const (
	farmSearchR   = 1800.0
	pathMaxExpand = 2500
	repathSecs    = 1.5
	waypointReach = 60.0
	// cutDelaySecs: §8.2 — bot tính vật cản mình chém được là mất thêm 0.3 giây.
	cutDelaySecs = 0.3
)

// Brain is one bot's AI state and personality.
type Brain struct {
	ID         uint32
	State      string
	rng        *rand.Rand
	nextDecide float64
	reaction   float64
	aimNoise   float64
	aggression float64

	goal        Vec
	hasGoal     bool
	stopNear    float64 // stop moving within this distance of the goal
	path        []Vec
	pathFor     Vec
	pathAt      float64
	ambushUntil float64
	// gateUntil: đang trên đường ra cổng tới khi nào. Không có mốc này thì bot đổi ý ngay
	// ở tick sau và đứng rung trước cổng mãi.
	gateUntil float64
	// fearless: không bao giờ bỏ chạy (chỉ dùng trong cmd/sim --duel).
	fearless bool
}

// trySkill: bot dùng kỹ năng của nhân vật mình theo luật đơn giản (spec §6.3).
//
// Mỗi nhân vật một điều kiện riêng, cố ý giữ THÔ: mục đích là để kỹ năng xuất hiện trong
// trận và đo được cân bằng (cmd/sim --duel), không phải để bot chơi giỏi. Bot tính toán
// quá khéo thì người chơi thật thấy ức chế, mà cũng không nói lên điều gì về cân bằng
// giữa người với người.
func (b *Brain) trySkill(g *game.Game, me *game.Player, threat, prey *game.Player) {
	if !g.CanCastSkill(me) {
		return
	}
	near := threat
	if near == nil {
		near = prey
	}
	reach := g.Geom(me).Reach
	switch me.Char {
	case "cu_rua": // Rụt Cổ: chỉ còn một vòng mà địch đã sát bên
		if g.Geom(me).Rings <= 1 && near != nil && near.Pos.Dist(me.Pos) < reach*1.5 {
			g.CastSkill(me)
		}
	case "ba_tam_loa", "map_tra_sua": // kỹ năng hình nón: địch phải ở phía trước
		if near != nil && b.inFront(me, near, 60) {
			g.CastSkill(me)
		}
	case "a_hon": // Bánh Bao: khi đã mất kha khá kiếm
		if me.Ring.Count*2 < g.Prog.MaxSwords(g.Prog.Realm(me.Level)) {
			g.CastSkill(me)
		}
	case "be_ti": // Gồng: khi có ai đó sát bên, dù mình đang đuổi hay đang bị đuổi
		if near != nil && near.Pos.Dist(me.Pos) < reach*2 {
			g.CastSkill(me)
		}
	case "shipper", "co_ba_ganh": // chạy/lướt: khi cần rút ngắn hoặc kéo giãn khoảng cách
		if near != nil {
			g.CastSkill(me)
		}
	default: // Ru Ngủ, Lưới Mì, Gọi Đàn Em: cứ có địch gần là dùng
		if near != nil && near.Pos.Dist(me.Pos) < reach*2.5 {
			g.CastSkill(me)
		}
	}
}

// inFront: mục tiêu có nằm trong nón `deg` độ trước mặt không.
func (b *Brain) inFront(me, target *game.Player, deg float64) bool {
	to := target.Pos.Sub(me.Pos)
	if to.Len() < 1e-6 {
		return false
	}
	d := math.Abs(math.Atan2(math.Sin(to.Angle()-me.Facing), math.Cos(to.Angle()-me.Facing)))
	return d*180/math.Pi <= deg/2
}

func New(id uint32, seed int64, g *game.Game) *Brain {
	r := rand.New(rand.NewSource(seed))
	bc := g.Cfg.Bots
	return &Brain{
		ID: id, State: StateWander, rng: r,
		reaction:   bc.ReactionMin + r.Float64()*(bc.ReactionMax-bc.ReactionMin),
		aimNoise:   r.Float64() * 0.35,
		aggression: 0.3 + r.Float64()*0.7,
	}
}

// NewFighter là bot LUÔN lao vào đánh, không bao giờ bỏ chạy — dùng cho `cmd/sim --duel`.
//
// Bot thường cân nhắc sức mạnh rồi bỏ chạy khi thấy đối thủ ngang cơ. Trong bài đấu từng
// cặp thì hai bên luôn ngang cơ, nên cả hai cùng chạy và 83% số trận kết thúc hoà — bảng
// số đẹp mà chẳng nói lên điều gì về kỹ năng. Bot hiếu chiến buộc trận đấu phải xảy ra.
func NewFighter(id uint32, seed int64, g *game.Game) *Brain {
	b := New(id, seed, g)
	b.aggression = 1
	b.fearless = true
	return b
}

// Update chooses an input for this tick and applies it to the game.
func (b *Brain) Update(g *game.Game) {
	me, ok := g.Players[b.ID]
	if !ok || me.Dead {
		return
	}
	if g.Time >= b.nextDecide {
		bc := g.Cfg.Bots
		b.decide(g, me)
		b.nextDecide = g.Time + bc.DecideMin + b.rng.Float64()*(bc.DecideMax-bc.DecideMin) + b.reaction*0.3
	}
	angle, moving := b.steer(g, me)
	// Bot chỉ Thiêu Đốt Tinh Huyết khi đang chạy trốn — đúng lúc đáng đánh đổi EXP (§4.1c).
	g.SetInput(b.ID, 0, angle, moving, b.State == StateFlee)
}

func (b *Brain) setGoal(p Vec, stopNear float64) {
	b.goal, b.hasGoal, b.stopNear = p, true, stopNear
}

func (b *Brain) decide(g *game.Game, me *game.Player) {
	bc := g.Cfg.Bots
	myPow := g.Power(me)
	threat, prey := b.scanPlayers(g, me, myPow)
	if b.fearless {
		threat = nil // không bao giờ bỏ chạy
	}
	b.trySkill(g, me, threat, prey)
	switch {
	case b.gateUntil > g.Time && b.headToGate(g, me):
		// Đang chạy ra cổng: giữ ý định, đừng đổi mỗi tick.
	case me.Ring.Count == 0 && threat != nil:
		// Spec §5.2 (v3.5): bị tước sạch kiếm thì chỉ còn nước chạy.
		b.State = StateFlee
		b.flee(g, me, threat)
	case me.Ring.Count == 0:
		b.farm(g, me) // không có kiếm thì đi nhặt đồ thay vì săn người
	case threat != nil:
		b.State = StateFlee
		b.flee(g, me, threat)
	case threat == nil && prey == nil && b.wantsGate(g, me) && b.headToGate(g, me):
		// Khu đang vắng mà cũng chẳng có gì nhặt: đi cổng sang khu khác tìm người. Không
		// có luật này thì bot nào rơi vào khu vắng sẽ quanh quẩn ở đó tới hết trận, và cả
		// một khu của thế giới thành vùng chết (spec v5 mục 8: "bot biết dùng cổng").
		b.gateUntil = g.Time + 20
	case prey != nil && b.rng.Float64() < 0.5+b.aggression/2:
		b.State = StateHunt
		// Còn lá chắn (§5.2) thì chém thân không ăn: đứng trong vùng mài — xa hơn tầm kiếm
		// của đối thủ nên không bị chém thân, nhưng đủ gần để hai vòng ngoài chạm nhau.
		// Hết lá chắn thì lao thẳng vào thân.
		stop := prey.Geom.Reach * 0.5
		if g.ShieldR(prey) > 0 {
			stop = prey.Geom.Reach + me.Geom.Reach*0.5
		}
		b.setGoal(prey.Pos, stop)
	case b.State == StateAmbush && g.Time < b.ambushUntil:
		// keep waiting in the bush
	case b.shouldAmbush(g, me):
		b.State = StateAmbush
		b.ambushUntil = g.Time + bc.AmbushSecs
	default:
		b.farm(g, me)
	}
}

// wantsGate: bot có nên đi cổng không. Chỉ khi quanh đây **không còn ai** — bot đang đánh
// nhau hay đang có mồi thì bỏ đi là vô lý.
func (b *Brain) wantsGate(g *game.Game, me *game.Player) bool {
	if me.InTransit(g.Time) || g.Time < me.GateReadyAt || len(g.Map.Gates) == 0 {
		return false
	}
	near := 0
	for _, o := range g.PlayerList() {
		if o.ID != me.ID && !o.Dead && o.Pos.Dist(me.Pos) < g.Geom(me).Reach*6 {
			near++
		}
	}
	return near == 0 && b.rng.Float64() < 0.02 // ~2% mỗi lần quyết định, không dồn cục
}

// headToGate đặt đích là cổng gần nhất. Trả về false khi không có cổng nào để đi.
//
// Ưu tiên **Thiên Môn** khi bot đã Đại Thừa trở lên (spec §8.4a): ở cảnh giới đó thứ đáng
// giá nhất là Thiên Tiên Đài, mà nó chỉ có ở Thiên Đình. Dưới cảnh giới ấy thì Thiên Môn
// chỉ ném chúng vào giữa ổ kiến lửa.
func (b *Brain) headToGate(g *game.Game, me *game.Player) bool {
	if me.InTransit(g.Time) {
		return true // đang dịch chuyển, cứ đứng yên chờ
	}
	wantHeaven := g.Prog.Realm(me.Level) >= g.Prog.Realm(g.Prog.TopMortalLevel())
	best, bd := Vec{}, math.MaxFloat64
	for _, gate := range g.Map.Gates {
		if gate.Dock {
			continue // bến chỉ là chỗ đặt chân, đi tới đó không dịch chuyển được
		}
		if wantHeaven != gate.Heaven && g.Map.HeavenGate() != nil {
			continue // đã Đại Thừa thì chỉ nhắm Thiên Môn; chưa thì tránh nó ra
		}
		at := Vec{X: gate.X, Y: gate.Y}
		if d := me.Pos.Dist(at); d < bd {
			best, bd = at, d
		}
	}
	if bd == math.MaxFloat64 {
		return false
	}
	b.State = StateFarm
	b.setGoal(best, 0) // đi thẳng vào tâm cổng thì mới kích hoạt
	return true
}

// scanPlayers returns the nearest visible threat and the nearest visible prey.
func (b *Brain) scanPlayers(g *game.Game, me *game.Player, myPow float64) (*game.Player, *game.Player) {
	bc := g.Cfg.Bots
	var threat, prey *game.Player
	bestT, bestP := math.Inf(1), math.Inf(1)
	// Bị Phun Trân Châu thì nhìn gần hẳn lại (spec §6.3).
	mySight := g.BlindSightMult(me)
	for _, o := range g.Players {
		if o == me || !g.CanSee(me, o) {
			continue
		}
		d := o.Pos.Dist(me.Pos)
		if b.fearless {
			if d > bc.HuntR*me.Scale*mySight {
				continue // bị bịt mắt thì cũng không thấy xa được
			}
			// Bot đấu tay đôi: ai gần nhất là mồi, bất kể mạnh yếu.
			//
			// Luật thường đòi đối thủ phải YẾU HƠN mới coi là mồi — mà trong bài đấu từng
			// cặp hai bên luôn ngang sức, nên không ai coi ai là mồi và 91% số trận kết
			// thúc hoà vì hai bot chưa từng gặp nhau.
			if d < bestP && !o.Invulnerable(g.Time) {
				prey, bestP = o, d
			}
			continue
		}
		pow := g.Power(o)
		sight := g.BotSightMult(o) // Ẩn Thân: bot chỉ phát hiện trong 50% tầm (spec §10)
		if pow > myPow*bc.FleeRatio && d < bc.FleeR*me.Scale*sight*mySight && d < bestT {
			threat, bestT = o, d
		}
		if pow < myPow*bc.HuntRatio && d < bc.HuntR*me.Scale*sight*mySight && d < bestP && !o.Invulnerable(g.Time) {
			prey, bestP = o, d
		}
	}
	return threat, prey
}

func (b *Brain) flee(g *game.Game, me, threat *game.Player) {
	away := me.Pos.Sub(threat.Pos).Norm()
	// Prefer a bush or a gate roughly away from the threat (spec §10 Flee).
	if spot, ok := b.nearestCover(g, me, g.Cfg.Bots.AmbushR*2); ok && spot.Sub(me.Pos).Norm().Dot(away) > 0.2 {
		b.setGoal(spot, 0)
		return
	}
	b.setGoal(me.Pos.Add(away.Mul(800)), 0)
}

// nearestCover finds the closest bush centre or gate within r.
func (b *Brain) nearestCover(g *game.Game, me *game.Player, r float64) (Vec, bool) {
	best, bestD := Vec{}, r
	found := false
	for _, z := range g.Map.Zones {
		if z.Kind != "bush" {
			continue
		}
		if c := z.Shape.Center(); c.Dist(me.Pos) < bestD {
			best, bestD, found = c, c.Dist(me.Pos), true
		}
	}
	for _, gt := range g.Map.Gates {
		c := Vec{X: gt.X, Y: gt.Y}
		if c.Dist(me.Pos) < bestD && g.Time >= me.GateReadyAt {
			best, bestD, found = c, c.Dist(me.Pos), true
		}
	}
	return best, found
}

func (b *Brain) shouldAmbush(g *game.Game, me *game.Player) bool {
	if g.Prog.Realm(me.Level) > 0 || b.rng.Float64() > 0.15*(1-b.aggression+0.3) {
		return false
	}
	spot, ok := b.nearestCover(g, me, g.Cfg.Bots.AmbushR)
	if ok {
		b.setGoal(spot, 20)
	}
	return ok
}

// farm heads to the nearest monster, item or weak obstacle (spec §10 Farm).
func (b *Brain) farm(g *game.Game, me *game.Player) {
	target, stop, ok := b.nearestFarmTarget(g, me)
	if !ok {
		b.State = StateWander
		if !b.hasGoal || me.Pos.Dist(b.goal) < 200 {
			ang := b.rng.Float64() * 2 * math.Pi
			b.setGoal(me.Pos.Add(geom.FromAngle(ang).Mul(1500)), 0)
		}
		return
	}
	b.State = StateFarm
	b.setGoal(target, stop)
}

func (b *Brain) nearestFarmTarget(g *game.Game, me *game.Player) (Vec, float64, bool) {
	best, stop, bestD := Vec{}, 0.0, farmSearchR*math.Sqrt(me.Scale)
	found := false
	for _, it := range g.Items {
		if d := it.Pos.Dist(me.Pos); d < bestD*0.8 && g.InView(me, it.Pos) {
			best, stop, bestD, found = it.Pos, 0, d/0.8, true
		}
	}
	for _, m := range g.Monsters {
		if d := m.Pos.Dist(me.Pos); d < bestD && g.InView(me, m.Pos) {
			best, stop, bestD, found = m.Pos, me.Geom.Reach*0.8, d, true
		}
	}
	// Vật cản bot CHÉM ĐƯỢC là mục tiêu đáng đi tới (spec v6.7 §8.2 mục 5): chém xong là
	// mở ra một lối đi, và sạp/xe còn rơi item. Bậc cao hơn cảnh giới thì bỏ qua — đi tới
	// đó chỉ để mài vào một bức tường không suy chuyển.
	for id, o := range g.Map.Obstacles {
		st := g.Obstacles[id]
		if st.Broken || !g.CanCutObstacle(me, o.Tier) {
			continue
		}
		c := o.Shape.Center()
		if d := c.Dist(me.Pos); d < bestD*1.2 {
			best, stop, bestD, found = c, me.Geom.Reach*0.9, d/1.2, true
		}
	}
	return best, stop, found
}

// steer turns the current goal into a movement angle, using A* when the way is blocked.
func (b *Brain) steer(g *game.Game, me *game.Player) (float64, bool) {
	if !b.hasGoal || b.State == StateAmbush && me.Pos.Dist(b.goal) < 40 {
		return 0, false
	}
	d := me.Pos.Dist(b.goal)
	if d <= b.stopNear {
		// Orbit slightly so the ring sweeps over the target.
		return b.goal.Sub(me.Pos).Angle() + math.Pi/2, true
	}
	next := b.nextWaypoint(g, me)
	ang := next.Sub(me.Pos).Angle() + (b.rng.Float64()-0.5)*b.aimNoise
	return ang, true
}

func (b *Brain) nextWaypoint(g *game.Game, me *game.Player) Vec {
	if b.clearLine(g, me.Pos, b.goal) {
		b.path = nil
		return b.goal
	}
	if b.path == nil || b.pathFor.Dist(b.goal) > 200 || g.Time-b.pathAt > repathSecs {
		// §8.2: bot coi vật cản MÌNH chém được là "đi qua được, tốn thêm 0.3s". Nếu không,
		// một Đạo Tổ đi vòng cả bản đồ để tránh cái rào tre mà họ chạm vào là tan.
		b.path = g.Grid.FindPathCutting(me.Pos, b.goal, pathMaxExpand, cutPenaltyCells(g, me),
			func(id int) bool { return g.CanCutObstacle(me, g.Map.Obstacles[id].Tier) })
		b.pathFor, b.pathAt = b.goal, g.Time
	}
	for len(b.path) > 0 && me.Pos.Dist(b.path[0]) < waypointReach*math.Sqrt(me.Scale) {
		b.path = b.path[1:]
	}
	if len(b.path) == 0 {
		return b.goal
	}
	return b.path[0]
}

// cutPenaltyCells đổi "0.3 giây" của §8.2 thành số ô lưới, theo tốc độ đi của chính bot:
// A* trong `world.Grid` đo quãng đường theo ô, nên hằng số thời gian phải quy đổi, không
// nhét thẳng 0.3 vào.
func cutPenaltyCells(g *game.Game, me *game.Player) float64 {
	speed := g.MoveSpeed(me)
	if speed <= 1 {
		return 0
	}
	return cutDelaySecs * speed / g.Grid.Cell
}

// clearLine samples the path grid between two points.
func (b *Brain) clearLine(g *game.Game, from, to Vec) bool {
	d := from.Dist(to)
	steps := int(d / (g.Grid.Cell / 2))
	for i := 1; i <= steps; i++ {
		if g.Grid.Blocked(from.Add(to.Sub(from).Mul(float64(i) / float64(steps)))) {
			return false
		}
	}
	return true
}
