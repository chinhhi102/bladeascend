package main

import (
	"context"
	"log"
	"net/http"
	"slices"
	"time"

	"github.com/coder/websocket"

	"kiemxoay/server/internal/netproto"
	"kiemxoay/server/internal/room"
)

const (
	sendQueue    = 256
	maxMsgBytes  = 4096
	writeTimeout = 5 * time.Second
)

// wsConn adapts a WebSocket to room.Conn with a buffered writer goroutine.
type wsConn struct {
	out chan []byte
}

func (c *wsConn) Send(b []byte) bool {
	select {
	case c.out <- b:
		return true
	default:
		return false // slow client: drop rather than stall the room
	}
}

func (s *server) handleWS(w http.ResponseWriter, r *http.Request) {
	ip := clientIP(r)
	if !s.limiter.acquire(ip) {
		http.Error(w, "too many connections", http.StatusTooManyRequests)
		return
	}
	defer s.limiter.release(ip)
	ws, err := websocket.Accept(w, r, &websocket.AcceptOptions{InsecureSkipVerify: true})
	if err != nil {
		return
	}
	ws.SetReadLimit(maxMsgBytes)
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()
	conn := &wsConn{out: make(chan []byte, sendQueue)}
	go writeLoop(ctx, ws, conn.out, cancel)
	client := room.NewClient(conn, s.debug)
	s.readLoop(ctx, ws, client)
	if rm := client.Room(); rm != nil {
		rm.RemoveClient(client)
	}
	_ = ws.Close(websocket.StatusNormalClosure, "")
}

func writeLoop(ctx context.Context, ws *websocket.Conn, out <-chan []byte, cancel func()) {
	defer cancel()
	for {
		select {
		case <-ctx.Done():
			return
		case b := <-out:
			wctx, done := context.WithTimeout(ctx, writeTimeout)
			err := ws.Write(wctx, websocket.MessageBinary, b)
			done()
			if err != nil {
				return
			}
		}
	}
}

// readLoop decodes client messages and enforces the per-connection rate limit.
func (s *server) readLoop(ctx context.Context, ws *websocket.Conn, c *room.Client) {
	windowStart, count := time.Now(), 0
	for {
		_, data, err := ws.Read(ctx)
		if err != nil {
			return
		}
		if time.Since(windowStart) > time.Second {
			windowStart, count = time.Now(), 0
		}
		if count++; count > s.cfg.Net.MaxMsgsPerSec {
			continue
		}
		var env netproto.Envelope
		if netproto.Decode(data, &env) != nil {
			continue
		}
		s.dispatch(c, env.T, data)
	}
}

func (s *server) dispatch(c *room.Client, t string, data []byte) {
	switch t {
	case netproto.TInput:
		var in netproto.Input
		if netproto.Decode(data, &in) == nil {
			c.SetInput(in)
		}
	case netproto.TJoin:
		s.handleJoin(c, data)
	case netproto.TPing:
		var p netproto.Ping
		if netproto.Decode(data, &p) == nil {
			if b, err := netproto.Encode(netproto.Pong{T: "pong", Client: p.Client, Server: float64(time.Now().UnixMilli())}); err == nil {
				c.Conn.Send(b)
			}
		}
	case netproto.TDebug:
		var d netproto.Debug
		if s.debug && netproto.Decode(data, &d) == nil {
			if rm := c.Room(); rm != nil {
				rm.Debug(c, d)
			}
		}
	default:
	}
}

func (s *server) handleJoin(c *room.Client, data []byte) {
	if c.Room() != nil {
		return // already playing
	}
	var j netproto.Join
	if netproto.Decode(data, &j) != nil {
		return
	}
	j.Name = s.names.Clean(j.Name)
	if j.Name == "" {
		sendErr(c, "bad_name")
		return
	}
	if !slices.Contains(s.cfg.Characters, j.Char) {
		j.Char = s.cfg.Characters[0]
	}
	if len(j.GuestID) > 64 {
		j.GuestID = j.GuestID[:64]
	}
	c.GuestID = j.GuestID
	rm := s.mgr.Join(c, j)
	log.Printf("join %q (%s) → room %s", j.Name, j.Char, rm.ID)
}

func sendErr(c *room.Client, msg string) {
	if b, err := netproto.Encode(netproto.Error{T: "error", Msg: msg}); err == nil {
		c.Conn.Send(b)
	}
}
