package room

import (
	"math"
	"sync"

	"kiemxoay/server/internal/game"
	"kiemxoay/server/internal/netproto"
)

// Conn is the transport for one browser tab; implemented by the WebSocket handler.
type Conn interface {
	// Send queues a message without blocking; it returns false when the queue is full.
	Send(b []byte) bool
}

// LossyConn là Conn biết bỏ gói CŨ khi hàng đợi đầy (spec mục 27).
//
// Mạng yếu thì hàng đợi đầy, và một snapshot cũ nằm trong hàng là thứ vô giá trị: client chỉ
// cần bản MỚI NHẤT. Giữ cái cũ rồi vứt cái mới là cách chắc chắn nhất để người chơi thấy
// nhân vật giật lùi.
type LossyConn interface {
	SendLossy(b []byte) bool
}

// Client is a connected human. The room goroutine owns everything except the input fields.
type Client struct {
	Conn     Conn
	GuestID  string
	Debug    bool
	PlayerID uint32
	room     *Room

	knownPlayers map[uint32]bool
	knownItems    map[uint32]bool
	// knownMonsters: đã gửi phần tĩnh (loại, máu tối đa, bán kính) của con quái này chưa.
	knownMonsters map[uint32]bool

	inMu   sync.Mutex
	seq    uint32
	angle  float64
	moving bool
	burn   bool
	dirty  bool
}

func NewClient(c Conn, debug bool) *Client {
	return &Client{Conn: c, Debug: debug}
}

// SetInput is called from the connection goroutine; the room applies it on its next tick.
func (c *Client) SetInput(in netproto.Input) {
	c.inMu.Lock()
	defer c.inMu.Unlock()
	c.seq = in.Seq
	c.angle = float64(in.Angle) / 256 * 2 * math.Pi
	c.moving, c.burn = in.Moving, in.Burn
	c.dirty = true
}

func (c *Client) takeInput() (seq uint32, angle float64, moving, burn, ok bool) {
	c.inMu.Lock()
	defer c.inMu.Unlock()
	d := c.dirty
	c.dirty = false
	return c.seq, c.angle, c.moving, c.burn, d
}

// Room returns the room the client is currently in (nil between matches).
func (c *Client) Room() *Room {
	c.inMu.Lock()
	defer c.inMu.Unlock()
	return c.room
}

func (c *Client) setRoom(r *Room) {
	c.inMu.Lock()
	defer c.inMu.Unlock()
	c.room = r
}

func (c *Client) send(v any) {
	b, err := netproto.Encode(v)
	if err == nil {
		c.Conn.Send(b)
	}
}

// sendSnapshot gửi gói ảnh trạng thái: được phép bỏ bản cũ khi nghẽn, khác với `welcome` hay
// `leaderboard` vốn mất là mất luôn.
func (c *Client) sendSnapshot(v any) {
	b, err := netproto.Encode(v)
	if err != nil {
		return
	}
	if lossy, ok := c.Conn.(LossyConn); ok {
		lossy.SendLossy(b)
		return
	}
	c.Conn.Send(b)
}

// playerState converts a game player to its wire form as seen by one viewer. Ẩn Thân is
// per-viewer: far away the target is flagged stealth (drawn faded, no name, no title),
// up close they look normal (spec §6.2).
func playerState(g *game.Game, viewer, p *game.Player, withName bool) netproto.PlayerState {
	faded := g.StealthFaded(viewer, p)
	s := netproto.PlayerState{
		ID: p.ID, X: float32(p.Pos.X), Y: float32(p.Pos.Y), Facing: float32(p.Facing),
		Level: p.Level, Count: p.Ring.Count, Phase: float32(p.Ring.Angle), Dir: int8(p.Ring.Dir),
		Tilt: float32(p.Ring.Tilt), Scale: float32(p.Scale), Mood: game.MoodIndex(p.Mood),
		Crack: float32(p.Ring.Crack), Flags: playerFlags(g, viewer, p),
	}
	// Omega bỏ hẳn khỏi gói tin: client tự tính được từ rev_secs và buff Tốc Kiếm mà nó đã
	// có sẵn (spec mục 27). Gửi lại một con số suy ra được là lãng phí 7 byte mỗi người mỗi
	// gói, 20 lần mỗi giây.
	if !faded {
		s.Title = p.Title
	}
	if p == viewer {
		s.Exp = float32(p.TotalExp)
		have, need := g.TienProgress(p)
		s.TienHave, s.TienNeed = float32(have), float32(need)
	}
	for b := 0; b < game.BuffCount; b++ {
		if p.HasBuff(b, g.Time) {
			s.Buffs |= 1 << b
		}
	}
	if withName && !faded {
		s.Name, s.Char = p.Name, p.Char
	}
	return s
}

func playerFlags(g *game.Game, viewer, p *game.Player) uint8 {
	var f uint8
	set := func(cond bool, bit uint8) {
		if cond {
			f |= bit
		}
	}
	set(p.Invulnerable(g.Time), netproto.FlagInvuln)
	set(p.Bush >= 0, netproto.FlagBush)
	// Your own stealth shows whenever the buff is up (so you can tell); other people's only
	// once you are far enough that they fade (spec §6.2).
	set(p == viewer && p.HasBuff(game.BuffStealth, g.Time) || g.StealthFaded(viewer, p), netproto.FlagStealth)
	set(p.InTransit(g.Time), netproto.FlagTransit)
	set(p.Dead, netproto.FlagDead)
	set(p.IsBot, netproto.FlagBot)
	set(p.RevealedUntil > g.Time, netproto.FlagRevealed)
	set(g.Burning(p), netproto.FlagBurning)
	return f
}
