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

// Client is a connected human. The room goroutine owns everything except the input fields.
type Client struct {
	Conn     Conn
	GuestID  string
	Debug    bool
	PlayerID uint32
	room     *Room

	knownPlayers map[uint32]bool
	knownItems   map[uint32]bool

	inMu   sync.Mutex
	seq    uint32
	angle  float64
	moving bool
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
	c.moving = in.Moving
	c.dirty = true
}

func (c *Client) takeInput() (uint32, float64, bool, bool) {
	c.inMu.Lock()
	defer c.inMu.Unlock()
	d := c.dirty
	c.dirty = false
	return c.seq, c.angle, c.moving, d
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

// playerState converts a game player to its wire form.
func playerState(g *game.Game, p *game.Player, withName bool) netproto.PlayerState {
	s := netproto.PlayerState{
		ID: p.ID, X: float32(p.Pos.X), Y: float32(p.Pos.Y), Facing: float32(p.Facing),
		Level: p.Level, Count: p.Ring.Count, Phase: float32(p.Ring.Angle), Dir: int8(p.Ring.Dir),
		Omega: float32(g.Omega(p)), RingR: float32(p.Geom.Radius), SwordL: float32(p.Geom.Length),
		SwordW: float32(p.Geom.Width), Scale: float32(p.Scale), Mood: p.Mood, Crack: float32(p.Ring.Crack),
		Exp: float32(p.TotalExp), Flags: playerFlags(g, p),
	}
	for b := 0; b < game.BuffCount; b++ {
		if p.HasBuff(b, g.Time) {
			s.Buffs |= 1 << b
		}
	}
	if withName {
		s.Name, s.Char = p.Name, p.Char
	}
	return s
}

func playerFlags(g *game.Game, p *game.Player) uint8 {
	var f uint8
	set := func(cond bool, bit uint8) {
		if cond {
			f |= bit
		}
	}
	set(p.Invulnerable(g.Time), netproto.FlagInvuln)
	set(p.Bush >= 0, netproto.FlagBush)
	set(p.HasBuff(game.BuffStealth, g.Time), netproto.FlagStealth)
	set(p.InTransit(g.Time), netproto.FlagTransit)
	set(p.Dead, netproto.FlagDead)
	set(p.IsBot, netproto.FlagBot)
	set(p.RevealedUntil > g.Time, netproto.FlagRevealed)
	return f
}
