package main

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"

	"bladeascend/server/internal/store"
)

// Lượt thích nhân vật — GAME_SPEC §6.6 (v6.48).
//
//	GET  /api/likes?guest=<id>  → {counts, top, mine, can_like}
//	POST /api/likes             → {guest, character} ("" = bỏ thích)
//
// Lượt thích KHÔNG ảnh hưởng luật chơi: nó chỉ là một ý kiến để chủ dự án biết nhân vật nào
// được ưa. Nên mọi lỗi ở đây đều phải "hỏng êm" — người chơi không bao giờ bị chặn vào trận
// vì bảng lượt thích trục trặc.

// likeMinSecs: phải có ít nhất một mạng sống ngần này giây mới được thích (§6.6, chống spam).
const likeMinSecs = 60

// likeIPMax: tối đa ngần này lượt thích CÒN HIỆU LỰC từ cùng một địa chỉ IP trong 24 giờ.
const likeIPMax = 5

// maxLikeBody: thân POST chỉ có hai chuỗi ngắn.
const maxLikeBody = 1 << 10

func (s *server) handleLikesGet(w http.ResponseWriter, r *http.Request) {
	s.writeLikes(w, r.URL.Query().Get("guest"))
}

// writeLikes trả bảng đếm kèm phần riêng của `guest`.
//
// Tách khỏi `handleLikesGet` vì câu trả lời của POST phải đọc `guest` từ THÂN GÓI, không
// phải từ query string — lần chạy đầu trả về `mine: ""` ngay sau khi vừa thích xong, và
// client vẽ cái tim rỗng dù lượt thích đã vào sổ.
func (s *server) writeLikes(w http.ResponseWriter, guest string) {
	if s.store == nil {
		writeJSON(w, store.LikeStatus{Counts: map[string]int{}})
		return
	}
	counts, err := s.store.LikeCounts(time.Now())
	if err != nil {
		http.Error(w, "likes", http.StatusInternalServerError)
		return
	}
	out := store.LikeStatus{Counts: counts, Top: store.TopLiked(counts)}
	if guest != "" {
		out.Mine, _ = s.store.MyLike(guest)
		out.CanLike, _ = s.store.CanLike(guest, likeMinSecs)
	}
	writeJSON(w, out)
}

func (s *server) handleLikesPost(w http.ResponseWriter, r *http.Request) {
	if s.store == nil {
		http.Error(w, "no store", http.StatusServiceUnavailable)
		return
	}
	body, err := io.ReadAll(io.LimitReader(r.Body, maxLikeBody))
	if err != nil {
		http.Error(w, "bad body", http.StatusBadRequest)
		return
	}
	var in struct {
		Guest     string `json:"guest"`
		Character string `json:"character"`
	}
	if json.Unmarshal(body, &in) != nil || in.Guest == "" {
		http.Error(w, "bad json", http.StatusBadRequest)
		return
	}
	// Tên nhân vật phải có thật trong config: không thì bảng đếm đầy những cái tên bịa và
	// nhãn "Được thích nhất" có thể rơi vào một nhân vật không tồn tại.
	if in.Character != "" && !s.knownCharacter(in.Character) {
		http.Error(w, "bad character", http.StatusBadRequest)
		return
	}
	err = s.store.SetLike(in.Guest, in.Character, store.HashIP(clientIP(r)), time.Now(), likeIPMax, likeMinSecs)
	switch {
	case errors.Is(err, store.ErrLikeNotAllowed):
		// 403 + mã riêng: client nói "chơi xong một mạng rồi thích nhé", không phải "lỗi".
		writeJSONStatus(w, http.StatusForbidden, map[string]string{"error": "need_play"})
		return
	case errors.Is(err, store.ErrLikeIPLimit):
		writeJSONStatus(w, http.StatusForbidden, map[string]string{"error": "ip_limit"})
		return
	case err != nil:
		http.Error(w, "likes", http.StatusInternalServerError)
		return
	}
	s.writeLikes(w, in.Guest)
}

func (s *server) knownCharacter(ch string) bool {
	for _, c := range s.cfg.Characters {
		if c == ch {
			return true
		}
	}
	return false
}

// writeJSONStatus: như `writeJSON` (http.go) nhưng đặt mã trạng thái trước khi ghi thân —
// `WriteHeader` sau khi đã ghi byte đầu tiên là một no-op im lặng.
func writeJSONStatus(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}
