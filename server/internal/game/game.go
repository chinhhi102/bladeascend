// Package game is the authoritative rules engine: entities, tick, collisions, EXP, swords, buffs.
package game

import (
	"math"
	"math/rand"
	"sort"

	"bladeascend/server/internal/config"
	"bladeascend/server/internal/world"
)

// Event is sent to clients (spec §13.3). Kind names match the spec's event list.
type Event struct {
	Kind string  `msgpack:"k"`
	A    uint32  `msgpack:"a,omitempty"`
	B    uint32  `msgpack:"b,omitempty"`
	X    float32 `msgpack:"x"`
	Y    float32 `msgpack:"y"`
	N    int     `msgpack:"n,omitempty"`
	S    string  `msgpack:"s,omitempty"`
	// F: một số thực đi kèm — ví dụ số giây của một trạng thái (spec §6.3).
	F      float32 `msgpack:"f,omitempty"`
	Global bool    `msgpack:"-"` // broadcast to the whole room regardless of AOI
	To     uint32  `msgpack:"-"` // when set, only this player receives it
}

// Game owns one room's world state.
type Game struct {
	Cfg  *config.Config
	Prog *Progression
	Map  *world.Map
	Rng  *rand.Rand
	Time float64
	// ExpTrace (tuỳ chọn): gọi mỗi lần ai đó nhận EXP, kèm nguồn. Chỉ công cụ đo gắn vào
	// (`cmd/sim --audit`); nil trên server thật.
	ExpTrace func(p *Player, src ExpSource, amount float64)

	altars   []altarState
	Players  map[uint32]*Player
	Monsters map[uint32]*Monster
	Items    map[uint32]*Item
	// ZoneIndex: khu này là khu thứ mấy của phòng (spec §8.0). Phòng đặt lúc dựng.
	ZoneIndex int
	// GateRouter / MoveToZone: do phòng cài, để cổng dịch chuyển đi được sang khu khác.
	// Khu không biết gì về khu khác — mọi việc qua ranh giới đều đi qua hai hàm này.
	GateRouter func(fromZone int, gate world.Gate) (GateDest, bool)
	MoveToZone func(p *Player, zone int, at Vec)
	// itemSlots: vật phẩm theo chỉ số dày đặc; freeSlots là các ô trống để cấp lại.
	itemSlots []*Item
	freeSlots []int
	// bladeBuf dùng lại giữa các lần so kiếm — spec §13.4 cấm cấp phát trong vòng nóng.
	bladeBuf []blade
	// obsBuf: danh sách id vật cản gần, dùng lại để khỏi cấp phát mỗi bước di chuyển.
	obsBuf []int
	// slicks: vệt trơn do Gánh Lướt rơi lại (spec §6.3).
	slicks    []slick
	Obstacles []ObstacleState
	// RoomRank trả hạng Uy Danh của một người trong CẢ PHÒNG (5 khu), 1 là cao nhất; 0 nếu
	// không biết. `Room` cắm vào mỗi tick (GAME_SPEC v6.19 §11.3).
	//
	// Là một HÀM chứ không phải map: `Game` không được biết gì về `Room`, và chép map sang
	// năm khu mỗi tick là năm lần cấp phát cho một thứ đọc xong là bỏ.
	RoomRank func(id uint32) int
	Grid     *world.Grid
	Events   []Event

	nextID uint32
	// idBase: bit cao của id, riêng cho từng khu (xem SetIDBase + newID).
	idBase       uint32
	order        []uint32 // player ids in join order, for deterministic iteration
	cooldowns    map[uint64]float64
	obsIndex     *staticIndex
	gridDirty    bool
	chestNextAt  float64
	chestAt      *Vec
	poolCount    int
	itemRespawns []float64
	orbRespawns  []pendingOrb
	itemGrid     *itemGrid
	// stalls: dãy sạp tạm của Cô Ba (spec v6.10 §6.3). Xem `stall.go` để biết vì sao chúng
	// là một tầng va chạm riêng chứ không nằm trong `Map.Obstacles`.
	stalls   []tempStall
	stallSeq uint32
	orbCount int
	// expiring: id các viên CÓ HẠN (Tiên khí). Giữ riêng để `expireItems` khỏi quét cả
	// bản đồ vật phẩm mỗi tick — xem ghi chú ở đó.
	expiring  []uint32
	bossCount int
}

// New dựng thế giới của MỘT khu.
//
// `zoneIndex` quyết định khoảng id của khu này — xem `newID`. Truyền vào đây, KHÔNG đặt
// bằng một setter gọi sau: `New` sinh hàng nghìn viên Linh khí ngay trong thân hàm
// (`seedOrbs`), nên một setter gọi sau là đã muộn và mọi viên đó mang id của khu 0.
//
// Tôi đã viết đúng lỗi đó: thêm `SetIDBase` gọi sau `New`, kèm một comment khẳng định
// "`New` chưa sinh entity nào". Comment sai, và bài kiểm bắt được ngay — quái thì đúng
// khoảng, còn vật phẩm vẫn trùng. Tham số bắt buộc thì không có cách nào gọi sai thứ tự.
func New(cfg *config.Config, m *world.Map, seed int64, zoneIndex int) *Game {
	g := &Game{
		idBase: uint32(zoneIndex) << 24,
		Cfg:    cfg, Prog: NewProgression(cfg.Cultivation, cfg.Sword.MaxSwords), Map: m,
		Rng:       rand.New(rand.NewSource(seed)),
		Players:   map[uint32]*Player{},
		Monsters:  map[uint32]*Monster{},
		Items:     map[uint32]*Item{},
		cooldowns: map[uint64]float64{},
	}
	// Mọi vật cản bắt đầu ở trạng thái còn nguyên. Không còn máu để khởi tạo (v6.7 §8.2).
	g.Obstacles = make([]ObstacleState, len(m.Obstacles))
	g.obsIndex = newStaticIndex(m, 500)
	g.Grid = world.NewGrid(m, cfg.Bots.PathGrid)
	g.Grid.Rebuild(g.ObstacleSolid)
	g.chestNextAt = cfg.Items.ChestEverySecs
	g.itemGrid = newItemGrid(m.Size, 500)
	g.seedOrbs()
	g.resetAltars()
	return g
}

// newID cấp id cho một entity mới.
//
// Id phải duy nhất trong CẢ PHÒNG, không chỉ trong khu. Trước đây `nextID` bắt đầu từ 0 ở
// MỌI khu, mà năm khu sinh ra cùng số quái và cùng số vật phẩm theo config — nên khu 0 và
// khu 1 cấp ra đúng những con số giống nhau. Đo thật: **79 id trùng** trong một phòng 5 khu
// ngay ở tick đầu tiên (và hơn 2 000 nếu đếm cả vật phẩm).
//
// Hậu quả không hề im lặng cho lắm, chỉ là khó lần ra:
//   - `Room.rankAt[p.ID]` là bảng CẢ PHÒNG → hai người khác khu cùng id ghi đè hạng của
//     nhau, nên danh hiệu phát sai (đúng thứ GAME_SPEC v6.19 §11.3 nói tới);
//   - bảng xếp hạng gửi `id`, client so `e.id === myId` để tô vàng dòng của mình → tô
//     nhầm người;
//   - người đi qua cổng mang theo id của mình sang khu mới, va vào một con quái đang dùng
//     đúng id đó.
//
// Cách sửa: mỗi khu một KHOẢNG id riêng, đặt ở bit cao. 5 khu × 16 triệu id mỗi khu, mà
// một khu đông nhất cũng chỉ ~7 000 entity — không có đường nào chạm trần.
func (g *Game) newID() uint32 {
	g.nextID++
	return g.idBase | g.nextID
}

func (g *Game) emit(e Event) { g.Events = append(g.Events, e) }

// DrainEvents returns and clears this tick's events.
func (g *Game) DrainEvents() []Event {
	ev := g.Events
	g.Events = nil
	return ev
}

// PlayerList returns players in join order.
func (g *Game) PlayerList() []*Player {
	out := make([]*Player, 0, len(g.order))
	for _, id := range g.order {
		if p, ok := g.Players[id]; ok {
			out = append(out, p)
		}
	}
	return out
}

// SpawnOptions controls how a new player enters the world.
type SpawnOptions struct {
	Name, Char, GuestID string
	IsBot               bool
	Level               int
	Swords              int
}

func (g *Game) AddPlayer(o SpawnOptions) *Player {
	p := &Player{ID: g.newID(), Name: o.Name, Char: o.Char, GuestID: o.GuestID, IsBot: o.IsBot,
		EmoteReady: map[string]float64{}, Bush: -1}
	g.Players[p.ID] = p
	g.order = append(g.order, p.ID)
	g.resetPlayer(p, o.Level, o.Swords)
	return p
}

// ForceState đặt thẳng cảnh giới, số kiếm và vị trí của một người. Chỉ dùng cho công cụ đo
// (cmd/loadtest, spec §13.6) để dựng kịch bản nặng nhất mà không phải chơi hàng giờ.
func (g *Game) ForceState(p *Player, level, swords int, at Vec) {
	p.Level = level
	p.PeakLevel = max(p.PeakLevel, level)
	g.seatTienKhi(p, level)
	p.Ring.Count = swords
	p.Pos = at
	p.InvulnUntil = g.Time + 1e9 // không để họ chém nhau chết giữa lúc đang đo
	g.refreshDerived(p)
}

// seatTienKhi đặt thanh Tiên khí đúng mốc khởi điểm khi một người được ĐẶT THẲNG vào một
// cảnh giới Tiên, thay vì tu lên tới đó.
//
// Phải gọi ở MỌI chỗ gán `p.Level` mà không đi qua `GainTienKhi`. Thiếu nó thì người đó có
// `Level` ở Tiên giới nhưng `TienKhi = 0`, và hai nguồn sự thật nói ngược nhau:
//
//	robOnKill:     Realm(Level) = 9 → IsTien → chia Tiên khí
//	               TienRealmOf → TienRealmForKhi(0) = -1 → return, KHÔNG LÀM GÌ
//
// Triệu chứng ngoài trận: **bot Chân Tiên bị chém mà không rớt cảnh giới** — vẫn bị đẩy
// lùi, vẫn hiện "Hit", nhưng đứng nguyên ở Chân Tiên mãi mãi. Bot dính nhiều nhất vì
// `randomBotLevel` chép thẳng `Level` của người hạng N, mà người đó có thể đang ở Tiên giới.
func (g *Game) seatTienKhi(p *Player, level int) {
	if r := g.Prog.Realm(level); g.Prog.IsTien(r) {
		p.TienKhi = math.Max(p.TienKhi, g.Prog.TienEntry(g.Prog.TienIndex(r)))
	}
}

// resetPlayer (re)spawns a player at a safe spot (spec §7).
func (g *Game) resetPlayer(p *Player, level, swords int) {
	p.TotalExp = g.Prog.Threshold(level)
	p.Level = level
	p.PeakExp, p.PeakLevel, p.Kills = p.TotalExp, level, 0
	// A new life starts from zero Uy Danh (spec §11.1).
	p.Score, p.Hits, p.Top1Secs, p.DaiThuaSecs, p.SkillCasts = 0, 0, 0, 0, 0
	p.Burning, p.BurnedExp, p.TienKhi = false, 0, 0
	// Đặt SAU khi xoá về 0: sinh thẳng ở một cảnh giới Tiên (bot chép cảnh giới người hạng
	// N, hoặc hồi sinh ở Tiên giới) thì thanh Tiên khí phải ở mốc khởi điểm của cảnh giới
	// đó, không phải 0. Xem `seatTienKhi`.
	g.seatTienKhi(p, level)
	p.ScoreBy, p.HitCooldown, p.ScoredRealms = map[string]float64{}, map[string]float64{}, map[int]bool{}
	p.Title, p.pendingTitle, p.titleSince = 0, 0, 0
	p.Ring = Ring{Count: max(g.Cfg.Sword.StartCount, swords), Dir: 1, Angle: g.Rng.Float64() * 2 * math.Pi}
	p.Ring.Tilt = RestTilt(g.Cfg.Sword, p.Ring.Dir)
	p.Buffs = [BuffCount]float64{}
	p.Dead, p.KilledBy, p.Transit = false, 0, nil
	p.Pos = g.safeSpawn()
	p.SpawnedAt = g.Time
	p.InvulnUntil = g.Time + g.Cfg.Player.InvulnOnSpawn
	g.setMood(p, MoodRespawn, g.Cfg.Mood.RespawnSecs)
	g.refreshDerived(p)
	// SAU refreshDerived: trần lượt phụ thuộc cảnh giới (Tiên giới +1) vừa tính xong.
	g.startSkillCharges(p)
}

// Respawn puts a dead player back at Luyện Khí 1 with 1 sword.
func (g *Game) Respawn(p *Player, level, swords int) {
	g.onDeathClear(p)
	g.resetPlayer(p, level, swords)
}

func (g *Game) RemovePlayer(id uint32) {
	delete(g.Players, id)
	for i, o := range g.order {
		if o == id {
			g.order = append(g.order[:i], g.order[i+1:]...)
			break
		}
	}
}

// Detach gỡ nhân vật khỏi thế giới này mà KHÔNG xoá trạng thái của họ — dùng khi chuyển
// sang khu khác (spec §8.0). Khác `RemovePlayer` ở chỗ người gọi giữ lại con trỏ Player và
// gắn nó vào khu mới, nên kiếm, buff, nứt, Tiên khí, Uy Danh đều nguyên vẹn.
func (g *Game) Detach(id uint32) *Player {
	p, ok := g.Players[id]
	if !ok {
		return nil
	}
	g.RemovePlayer(id)
	return p
}

// Attach nhận một nhân vật từ khu khác sang. Hình học vòng kiếm và khung nhìn tính lại theo
// bản đồ mới; vị trí do người gọi đặt trước.
func (g *Game) Attach(p *Player) {
	g.Players[p.ID] = p
	g.order = append(g.order, p.ID)
	p.Pos = g.resolveBlocking(p.Pos, g.BodyR(p))
	g.refreshDerived(p)
}

// SetInput applies a client input (already validated by the caller).
func (g *Game) SetInput(id uint32, seq uint32, angle float64, moving, burning bool) {
	p, ok := g.Players[id]
	if !ok {
		return
	}
	p.InputAngle, p.Moving, p.Burning = angle, moving, burning
	if seq > p.LastSeq {
		p.LastSeq = seq
	}
}

// refreshDerived recomputes scale and ring geometry from level, sword count and buffs.
func (g *Game) refreshDerived(p *Player) {
	p.Scale = g.Prog.Scale(p.Level)
	// Spec v4.3 bỏ Trường Kiếm: tầm chém chỉ tăng theo số vòng và cảnh giới (§4.1a).
	p.Geom = ComputeRingGeom(g.Cfg.Sword, g.Cfg.Player.BodyR, p.Scale, p.Ring.Count)
	g.refreshView(p)
	// Trần lượt Bí Kíp tụt khi rớt khỏi Tiên giới (spec v6.5 §6.3). Đặt ở ĐÂY vì đây là
	// chỗ duy nhất mọi đường đổi cảnh giới đều đi qua — nhặt EXP, bị chém, Thiêu Đốt,
	// `ForceState` của phím test. Cắt ở từng chỗ gọi thì chắc chắn sót một đường.
	g.clampSkillCharges(p)
}

// refreshView tính sẵn nửa khung nhìn. Nó chỉ phụ thuộc Scale và tầm kiếm — hai thứ vừa
// tính ở trên — nên tính một lần ở đây thay vì tính lại cho TỪNG vật phẩm, TỪNG client,
// TỪNG tick như trước. pprof đo được: `math.Pow` trong hàm này một mình ăn 35% CPU server,
// vì nó là `exp(log())` và bị gọi hàng triệu lần mỗi giây trong vòng lọc tầm nhìn.
func (g *Game) refreshView(p *Player) {
	pc := g.Cfg.Player
	view := math.Max(pc.CameraViewH*math.Pow(p.Scale, pc.CameraExp), pc.CameraReachMul*p.Geom.Reach)
	p.ViewHH = view/2 + g.Cfg.Net.AOIMargin
	p.ViewHW = p.ViewHH * g.Cfg.Net.AOIAspect
}

func (g *Game) BodyR(p *Player) float64 { return g.Cfg.Player.BodyR * p.Scale }

// Omega is the spin rate: one revolution per rev_secs, × Tốc Kiếm.
func (g *Game) Omega(p *Player) float64 {
	w := 2 * math.Pi / math.Max(g.Cfg.Sword.RevSecs, 1e-3)
	if p.HasBuff(BuffSpin, g.Time) {
		w *= g.Cfg.Buffs["spin"].Mult
	}
	// v6.27: ngủ gục / choáng KHÔNG dừng vòng kiếm — chém trúng là chết, khoá kiếm = giết chắc.
	return w
}

// Tick advances the simulation by dt seconds in the order of spec §5.
func (g *Game) Tick(dt float64) {
	g.Time += dt
	g.updateTimers()
	g.moveMinions(dt)
	g.movePlayers(dt)
	g.moveMonsters(dt)
	for _, p := range g.PlayerList() {
		if !p.Dead {
			g.spinRing(p, dt)
		}
	}
	g.updateSkillState(dt)
	g.rechargeSkills(dt)
	g.resolveSwordBody()
	g.resolveSwordSword()
	g.resolveSwordWorld()
	g.flushSwordBreaks()
	g.resolvePickups()
	// Gộp SAU khi nhặt để sự kiện đi ngay trong tick này, không trễ một nhịp.
	g.flushOrbPickups()
	g.maintainSpawns()
	g.updateAltars()
	g.expireItems()
	g.updateMoods()
	g.updateTop1(dt)
	g.updateTitles()
	if g.gridDirty {
		g.Grid.Rebuild(g.ObstacleSolid)
		g.gridDirty = false
	}
	g.pruneCooldowns()
}

// spinRing advances the ring and records the arc it swept this tick (spec §4.1b).
func (g *Game) spinRing(p *Player, dt float64) {
	r := &p.Ring
	step := r.Dir * g.Omega(p) * dt
	r.SweepLo, r.Sweep = r.Angle, math.Abs(step)
	if step < 0 {
		r.SweepLo = r.Angle + step
	}
	r.Angle = math.Mod(r.Angle+step, 2*math.Pi)
	stepTilt(g.Cfg.Sword, r, dt)
}

// pairKey packs two ids and a tag into one cooldown key.
func pairKey(tag uint8, a, b uint32, sub int) uint64 {
	return uint64(tag)<<60 | uint64(sub&0xff)<<52 | uint64(a&0x3ffffff)<<26 | uint64(b&0x3ffffff)
}

// onCooldown checks a cooldown without claiming it; pair it with startCooldown when the
// action actually happens (cooldownReady does both at once).
func (g *Game) onCooldown(key uint64) bool {
	until, ok := g.cooldowns[key]
	return ok && until > g.Time
}

func (g *Game) startCooldown(key uint64, secs float64) { g.cooldowns[key] = g.Time + secs }

func (g *Game) cooldownReady(key uint64, secs float64) bool {
	if until, ok := g.cooldowns[key]; ok && until > g.Time {
		return false
	}
	g.cooldowns[key] = g.Time + secs
	return true
}

func (g *Game) pruneCooldowns() {
	if int(g.Time*10)%30 != 0 {
		return
	}
	for k, v := range g.cooldowns {
		if v <= g.Time {
			delete(g.cooldowns, k)
		}
	}
}

// Ranking returns living players sorted by total_exp descending.
func (g *Game) Ranking() []*Player {
	var out []*Player
	for _, p := range g.PlayerList() {
		if !p.Dead {
			out = append(out, p)
		}
	}
	// Spec §11.2: the room board ranks by Uy Danh — the board is about who is winning the
	// fight, not who farmed longest. total_exp only breaks ties.
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Score != out[j].Score {
			return out[i].Score > out[j].Score
		}
		return out[i].TotalExp > out[j].TotalExp
	})
	return out
}

// RankOf is p's 1-based place in the room board (same order as Ranking).
func (g *Game) RankOf(p *Player) int {
	rank := 1
	for _, o := range g.PlayerList() {
		if o == p || o.Dead {
			continue
		}
		if o.Score > p.Score || (o.Score == p.Score && o.TotalExp > p.TotalExp) {
			rank++
		}
	}
	return rank
}

// Vec is a convenience constructor for other packages' tests.
func (g *Game) Vec(x, y float64) Vec { return Vec{X: x, Y: y} }
