package game

import "kiemxoay/server/internal/geom"

type Vec = geom.Vec

// Buff kinds (spec §6). Their order is the bit order sent to clients.
const (
	BuffSpin = iota
	BuffHaste
	BuffMagnet
	BuffLong
	BuffShield
	BuffExp2
	BuffStealth
	BuffCount
)

var BuffNames = [BuffCount]string{"spin", "haste", "magnet", "long", "shield", "exp2", "stealth"}

// BuffByName maps item kind → buff index.
func BuffByName(name string) (int, bool) {
	for i, n := range BuffNames {
		if n == name {
			return i, true
		}
	}
	return 0, false
}

// Mood names = animation names of spec §12.
const (
	MoodIdle         = "idle"
	MoodRun          = "run"
	MoodChase        = "chase"
	MoodFlee         = "flee"
	MoodHit          = "hit"
	MoodLevelup      = "levelup"
	MoodBreakthrough = "breakthrough"
	MoodRealmDrop    = "realm_drop"
	MoodTeleportOut  = "teleport_out"
	MoodTeleportIn   = "teleport_in"
	MoodDeath        = "death"
	MoodRespawn      = "respawn"
)

// Transit tracks a player moving through a gate (spec §8.4).
type Transit struct {
	Dest     Vec
	DestGate int
	OutAt    float64 // time the player reappears at Dest
	DoneAt   float64 // time transit (immunity from combat) ends
	Arrived  bool
	WarnSent bool
}

type Player struct {
	ID      uint32
	Name    string
	Char    string
	GuestID string
	IsBot   bool

	Pos        Vec
	Facing     float64
	InputAngle float64
	Moving     bool
	LastSeq    uint32

	TotalExp  float64
	Level     int
	Ring      Ring
	Geom      RingGeom
	Scale     float64

	PeakExp   float64
	PeakLevel int
	Kills     int
	SpawnedAt float64

	InvulnUntil float64
	Buffs       [BuffCount]float64 // expiry time; 0 = inactive
	KnockVel    Vec
	KnockUntil  float64

	Dead       bool
	DiedAt     float64
	KilledBy   uint32
	KillerName string
	KillerChar string

	Mood      string
	MoodUntil float64 // forced moods (hit, levelup…) hold until this time
	MoodSince float64

	InBush        bool
	RevealedUntil float64 // bush reveal after being hit
	GateReadyAt   float64
	lastTeleportIn float64
	Transit       *Transit
	EmoteReady    map[string]float64
}

func (p *Player) HasBuff(b int, now float64) bool { return p.Buffs[b] > now }

func (p *Player) Invulnerable(now float64) bool { return p.InvulnUntil > now }

// InTransit is true while the player is inside a gate and immune to combat.
func (p *Player) InTransit(now float64) bool { return p.Transit != nil && now < p.Transit.DoneAt }

// Hidden reports whether the player is out of sight because they are mid-teleport.
func (p *Player) Hidden(now float64) bool { return p.Transit != nil && !p.Transit.Arrived }

type Monster struct {
	ID        uint32
	Kind      string
	Pos       Vec
	R         float64
	HP, MaxHP int
	Exp       float64
	Heading   float64
	TurnAt    float64
	FleeUntil float64
	FleeFrom  Vec
}

// Item kinds: "sword", buff names, "orb" (EXP orb, Value = EXP), "chest".
type Item struct {
	ID    uint32
	Kind  string
	Pos   Vec
	Value float64
}

// ObstacleState is the dynamic state of a destructible obstacle (spec §8.2).
type ObstacleState struct {
	HP       int
	Broken   bool
	BrokenAt float64
	Warned   bool
}
