// Package room runs one goroutine per room: tick, bots, snapshots, matchmaking (spec §9, §13.3).
package room

import (
	"log"
	"math/rand"
	"time"

	"kiemxoay/server/internal/bot"
	"kiemxoay/server/internal/config"
	"kiemxoay/server/internal/game"
	"kiemxoay/server/internal/netproto"
	"kiemxoay/server/internal/store"
	"kiemxoay/server/internal/world"
)

// Recorder persists finished human lives (implemented by store.Store).
type Recorder interface {
	Record(l store.Life) error
}

type pendingRespawn struct {
	id uint32
	at float64
}

type Room struct {
	ID        string
	Created   time.Time
	cfg       *config.Config
	G         *game.Game
	mgr       *Manager
	inbox     chan func()
	clients   map[uint32]*Client
	brains    map[uint32]*bot.Brain
	respawns  []pendingRespawn
	pending   []game.Event
	recorder  Recorder
	rng       *rand.Rand
	tick      uint32
	botTimer  float64
	lbTimer   float64
	emptyFor  float64
	reserved  int // humans counted by the manager (guarded by Manager.mu)
	closed    bool
	snapEvery float64
	snapAcc   float64
	namesAt   float64
	namesDue  bool
}

func newRoom(id string, cfg *config.Config, m *world.Map, mgr *Manager, rec Recorder, seed int64) *Room {
	return &Room{
		ID: id, Created: time.Now(), cfg: cfg, G: game.New(cfg, m, seed), mgr: mgr,
		inbox: make(chan func(), 1024), clients: map[uint32]*Client{}, brains: map[uint32]*bot.Brain{},
		recorder: rec, rng: rand.New(rand.NewSource(seed + 1)),
		snapEvery: 1 / float64(cfg.Net.SnapshotHz),
	}
}

// Do runs fn on the room goroutine.
func (r *Room) Do(fn func()) {
	select {
	case r.inbox <- fn:
	default:
		log.Printf("room %s inbox full, dropping message", r.ID)
	}
}

// Run ticks the room until it closes (no humans for empty_close_secs).
func (r *Room) Run() {
	dt := 1 / float64(r.cfg.Net.TickHz)
	t := time.NewTicker(time.Duration(float64(time.Second) * dt))
	defer t.Stop()
	for range t.C {
		r.Step(dt)
		if r.closed {
			return
		}
	}
}

// Step advances one tick; exported so cmd/sim can run rooms without real time.
func (r *Room) Step(dt float64) {
	r.drainInbox()
	r.applyInputs()
	for _, b := range r.brains {
		b.Update(r.G)
	}
	r.G.Tick(dt)
	r.tick++
	r.pending = append(r.pending, r.G.DrainEvents()...)
	r.handleDeaths()
	r.updateBots(dt)
	r.snapAcc += dt
	if r.snapAcc >= r.snapEvery-1e-9 {
		r.snapAcc -= r.snapEvery
		r.broadcast()
	}
	r.lbTimer += dt
	if r.lbTimer >= 1/float64(r.cfg.Net.LeaderboardHz) {
		r.lbTimer = 0
		r.sendLeaderboards()
	}
	r.checkEmpty(dt)
}

func (r *Room) drainInbox() {
	for {
		select {
		case fn := <-r.inbox:
			fn()
		default:
			return
		}
	}
}

func (r *Room) applyInputs() {
	for id, c := range r.clients {
		if seq, ang, mov, burn, ok := c.takeInput(); ok {
			r.G.SetInput(id, seq, ang, mov, burn)
		}
	}
}

// addHuman spawns the client's player at Luyện Khí 1 with 1 sword (spec §7).
func (r *Room) addHuman(c *Client, j netproto.Join) {
	p := r.G.AddPlayer(game.SpawnOptions{Name: j.Name, Char: j.Char, GuestID: j.GuestID, Swords: r.cfg.Sword.StartCount})
	c.PlayerID = p.ID
	c.knownPlayers, c.knownItems = map[uint32]bool{}, map[uint32]bool{}
	r.clients[p.ID] = c
	c.setRoom(r)
	w := netproto.Welcome{T: "welcome", ID: p.ID, Room: r.ID, Map: r.G.Map.ID, TickHz: r.cfg.Net.TickHz, Time: r.G.Time, Debug: c.Debug}
	for id, st := range r.G.Obstacles {
		if st.Broken || st.HP != r.G.Map.Obstacles[id].MaxHP {
			w.Obstacles = append(w.Obstacles, netproto.ObstacleState{ID: id, HP: st.HP, Broken: st.Broken})
		}
	}
	c.send(w)
}

// RemoveClient is called when a connection closes.
func (r *Room) RemoveClient(c *Client) {
	r.Do(func() {
		if _, ok := r.clients[c.PlayerID]; !ok {
			return
		}
		if p, ok := r.G.Players[c.PlayerID]; ok && !p.Dead {
			r.record(p)
		}
		r.detach(c)
	})
}

func (r *Room) detach(c *Client) {
	r.G.RemovePlayer(c.PlayerID)
	delete(r.clients, c.PlayerID)
	c.PlayerID = 0
	c.setRoom(nil)
	r.mgr.release(r)
}

// handleDeaths sends the Tổng kết screen after the death animation and records the life.
func (r *Room) handleDeaths() {
	for _, p := range r.G.PlayerList() {
		if !p.Dead || r.G.Time-p.DiedAt < r.cfg.Player.DeathAnimSecs {
			continue
		}
		if c, ok := r.clients[p.ID]; ok {
			rank := r.rankOf(p)
			r.record(p)
			c.send(netproto.Dead{T: "dead", KillerName: p.KillerName, KillerChar: p.KillerChar, PeakLevel: p.PeakLevel,
				PeakExp: p.PeakExp, Kills: p.Kills, SurviveSec: p.DiedAt - p.SpawnedAt, Rank: rank,
				Score: p.Score, Hits: p.Hits, ScoreBy: p.ScoreBy, BurnedExp: p.BurnedExp,
				DaiThuaSecs: p.DaiThuaSecs, Top1Secs: p.Top1Secs})
			r.detach(c)
			continue
		}
		if b, ok := r.brains[p.ID]; ok && !r.respawnQueued(b.ID) {
			r.onBotDeath(p)
		}
	}
}

func (r *Room) record(p *game.Player) {
	if p.IsBot || r.recorder == nil {
		return
	}
	end := p.DiedAt
	if !p.Dead {
		end = r.G.Time
	}
	err := r.recorder.Record(store.Life{GuestID: p.GuestID, Name: p.Name, Character: p.Char,
		Score: p.Score, Hits: p.Hits, PeakExp: p.PeakExp, PeakLevel: p.PeakLevel, Kills: p.Kills,
		DaiThuaSecs: p.DaiThuaSecs, Top1Secs: p.Top1Secs,
		SurviveSecs: end - p.SpawnedAt, Map: r.G.Map.ID, EndedAt: time.Now().Unix()})
	if err != nil {
		log.Printf("record life: %v", err)
	}
}

// rankOf is the player's 1-based rank in the room board (spec §11.2: Uy Danh, then exp).
func (r *Room) rankOf(p *game.Player) int { return r.G.RankOf(p) }

func (r *Room) checkEmpty(dt float64) {
	if len(r.clients) > 0 || r.mgr.humansOf(r) > 0 {
		r.emptyFor = 0
		return
	}
	r.emptyFor += dt
	if r.emptyFor >= r.cfg.Room.EmptyCloseSecs && r.mgr.tryClose(r) {
		r.closed = true
	}
}

// Humans returns connected human players (room goroutine only).
func (r *Room) Humans() int { return len(r.clients) }

// Bots returns the number of bot brains, including bots waiting to respawn.
func (r *Room) Bots() int { return len(r.brains) }
