// Package netproto defines the MessagePack messages exchanged with the browser client (spec §13.3).
// Every message is a map with a "t" (type) key; short field names keep packets small.
package netproto

import (
	"github.com/vmihailenco/msgpack/v5"

	"kiemxoay/server/internal/game"
)

// Client → server message types.
const (
	TJoin  = "join"
	TInput = "in"
	TPing  = "ping"
	TDebug = "dbg"
)

// Envelope is decoded first to find the message type.
type Envelope struct {
	T string `msgpack:"t"`
}

type Join struct {
	T       string `msgpack:"t"`
	Name    string `msgpack:"name"`
	Char    string `msgpack:"char"`
	GuestID string `msgpack:"guest_id"`
}

// Input carries the movement direction quantised to 256 steps (angle_u8).
type Input struct {
	T      string `msgpack:"t"`
	Seq    uint32 `msgpack:"s"`
	Angle  uint8  `msgpack:"a"`
	Moving bool   `msgpack:"m"`
}

type Ping struct {
	T      string  `msgpack:"t"`
	Client float64 `msgpack:"c"`
}

// Debug commands are only honoured when the server runs with -debug.
type Debug struct {
	T   string `msgpack:"t"`
	Cmd string `msgpack:"cmd"`
	N   int    `msgpack:"n"`
}

// ObstacleState is sent in Welcome for obstacles that are not intact.
type ObstacleState struct {
	ID     int  `msgpack:"id"`
	HP     int  `msgpack:"hp"`
	Broken bool `msgpack:"b"`
}

type Welcome struct {
	T         string          `msgpack:"t"`
	ID        uint32          `msgpack:"id"`
	Room      string          `msgpack:"room"`
	Map       string          `msgpack:"map"`
	TickHz    int             `msgpack:"tick_hz"`
	Time      float64         `msgpack:"time"`
	Obstacles []ObstacleState `msgpack:"obs"`
	Debug     bool            `msgpack:"debug"`
}

// Player flags.
const (
	FlagInvuln = 1 << iota
	FlagBush
	FlagStealth
	FlagTransit
	FlagDead
	FlagBot
	FlagRevealed
)

type PlayerState struct {
	ID     uint32  `msgpack:"id"`
	X      float32 `msgpack:"x"`
	Y      float32 `msgpack:"y"`
	Facing float32 `msgpack:"f"`
	Level  int     `msgpack:"L"`
	Count  int     `msgpack:"n"`
	Phase  float32 `msgpack:"ph"`
	Dir    int8    `msgpack:"d"`
	Omega  float32 `msgpack:"w"`
	Tilt   float32 `msgpack:"tl"` // blade lean (rad); ring geometry is derived from scale, count and buffs (spec §13.3)
	Scale  float32 `msgpack:"sc"`
	Mood   string  `msgpack:"md"`
	Buffs  uint8   `msgpack:"bf"`
	Flags  uint8   `msgpack:"fl"`
	Crack  float32 `msgpack:"ck,omitempty"`
	Exp    float32 `msgpack:"xp"`
	Title  int     `msgpack:"ti,omitempty"` // room rank badge 1..3 (spec §11.3), 0 = none
	Name   string  `msgpack:"nm,omitempty"` // only the first time a client sees this player
	Char   string  `msgpack:"ch,omitempty"`
}

type MonsterState struct {
	ID   uint32  `msgpack:"id"`
	X    float32 `msgpack:"x"`
	Y    float32 `msgpack:"y"`
	Kind string  `msgpack:"k"`
	HP   int     `msgpack:"hp"`
	Max  int     `msgpack:"mx"`
	R    float32 `msgpack:"r"`
}

type ItemState struct {
	ID    uint32  `msgpack:"id"`
	X     float32 `msgpack:"x"`
	Y     float32 `msgpack:"y"`
	Kind  string  `msgpack:"k"`
	Value float32 `msgpack:"v,omitempty"`
}

// Snapshot is sent at snapshot_hz. Items are diffed per client (added / removed).
type Snapshot struct {
	T            string         `msgpack:"t"`
	Tick         uint32         `msgpack:"tk"`
	Time         float64        `msgpack:"tm"`
	Ack          uint32         `msgpack:"ack"`
	Players      []PlayerState  `msgpack:"p"`
	Monsters     []MonsterState `msgpack:"m"`
	ItemsAdded   []ItemState    `msgpack:"ia,omitempty"`
	ItemsRemoved []uint32       `msgpack:"ir,omitempty"`
	Events       []game.Event   `msgpack:"ev,omitempty"`
}

type LeaderEntry struct {
	ID    uint32  `msgpack:"id"`
	Name  string  `msgpack:"nm"`
	Exp   float32 `msgpack:"xp"`
	Level int     `msgpack:"L"`
}

type Leaderboard struct {
	T     string        `msgpack:"t"`
	Top   []LeaderEntry `msgpack:"top"`
	Rank  int           `msgpack:"rank"`
	Total int           `msgpack:"total"`
}

// Feed is a kill-feed line with names resolved on the server (spec §7).
type Feed struct {
	T     string `msgpack:"t"`
	Kind  string `msgpack:"k"`
	A     string `msgpack:"a"`
	B     string `msgpack:"b,omitempty"`
	Realm int    `msgpack:"r"`
}

// Dead is the Tổng kết screen payload (spec §7).
type Dead struct {
	T          string  `msgpack:"t"`
	KillerName string  `msgpack:"killer"`
	KillerChar string  `msgpack:"killer_char"`
	PeakLevel  int     `msgpack:"peak_level"`
	PeakExp    float64 `msgpack:"peak_exp"`
	Kills      int     `msgpack:"kills"`
	SurviveSec float64 `msgpack:"survive"`
	Rank       int     `msgpack:"rank"`
}

type Pong struct {
	T      string  `msgpack:"t"`
	Client float64 `msgpack:"c"`
	Server float64 `msgpack:"s"`
}

type Error struct {
	T   string `msgpack:"t"`
	Msg string `msgpack:"msg"`
}

func Encode(v any) ([]byte, error) { return msgpack.Marshal(v) }

func Decode(b []byte, v any) error { return msgpack.Unmarshal(b, v) }
