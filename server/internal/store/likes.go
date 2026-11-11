package store

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"sync"
	"time"
)

// Lượt thích nhân vật (GAME_SPEC §6.6, v6.48).
//
// Mỗi người chơi (theo `guest_id`) có ĐÚNG MỘT lượt thích cho một nhân vật: thích người
// khác thì lượt chuyển sang, bấm lại thì bỏ thích. Một dòng mỗi `guest_id` — khoá chính là
// `guest_id`, nên "đổi ý" chỉ là một `UPSERT` chứ không phải đếm lại gì.
//
// Vì sao không lưu lịch sử: lượt thích là một ý kiến HIỆN TẠI ("tôi thích nhân vật này
// nhất"), không phải một sự kiện đã xảy ra. Giữ lịch sử thì phải định nghĩa "lượt thích
// đang hiệu lực" ở mọi câu truy vấn, và chỉ cần sót một chỗ là số đếm sai.

// likeTable: xem `schema` ở store.go — chạy cùng lúc, cùng một `db.Exec`.
const likeTable = `
CREATE TABLE IF NOT EXISTS character_likes (
	guest_id  TEXT PRIMARY KEY,
	character TEXT NOT NULL,
	ip_hash   TEXT NOT NULL,
	at        INTEGER NOT NULL
);
CREATE INDEX IF NOT EXISTS likes_char ON character_likes(character);
CREATE INDEX IF NOT EXISTS likes_ip ON character_likes(ip_hash, at);`

// likeCache giữ bảng đếm 60 giây (§6.6: "GET /api/likes (đếm, cache 60s)").
type likeCache struct {
	mu    sync.Mutex
	at    time.Time
	count map[string]int
}

// HashIP băm địa chỉ IP trước khi lưu (luật 15 của CLAUDE.md: repo công khai, KHÔNG BAO GIỜ
// lưu IP thật ở chỗ có thể lọt ra ngoài).
//
// Băm có muối cố định theo tiến trình là đủ cho việc duy nhất mục đích của nó — đếm "mấy
// lượt thích từ cùng một nhà" trong 24 giờ. Không cần đảo ngược được, và không nên đảo
// ngược được.
func HashIP(ip string) string {
	sum := sha256.Sum256([]byte("bladeascend-like:" + ip))
	return hex.EncodeToString(sum[:8])
}

// LikeStatus là câu trả lời của `GET /api/likes`.
type LikeStatus struct {
	// Counts: số lượt thích theo tên nhân vật.
	Counts map[string]int `json:"counts"`
	// Top: nhân vật nhiều lượt thích nhất ("" khi chưa ai thích) — nhãn "Được thích nhất".
	Top string `json:"top"`
	// Mine: nhân vật mà chính người gọi đang thích ("" = chưa thích ai).
	Mine string `json:"mine"`
	// CanLike: người gọi đã đủ điều kiện thích chưa (§6.6: đã chơi ≥ 1 mạng ≥ 60 s).
	CanLike bool `json:"can_like"`
}

// LikeCounts đếm lượt thích theo nhân vật, cache 60 giây.
func (s *Store) LikeCounts(now time.Time) (map[string]int, error) {
	s.likes.mu.Lock()
	defer s.likes.mu.Unlock()
	if s.likes.count != nil && now.Sub(s.likes.at) < likeCacheTTL {
		return s.likes.count, nil
	}
	rows, err := s.db.Query(`SELECT character, COUNT(*) FROM character_likes GROUP BY character`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[string]int{}
	for rows.Next() {
		var ch string
		var n int
		if err := rows.Scan(&ch, &n); err != nil {
			return nil, err
		}
		out[ch] = n
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	s.likes.count, s.likes.at = out, now
	return out, nil
}

const likeCacheTTL = 60 * time.Second

// TopLiked: nhân vật nhiều lượt thích nhất. Hoà thì lấy tên nhỏ hơn theo thứ tự chữ — một
// kết quả ỔN ĐỊNH, không nhảy qua nhảy lại mỗi lần cache hết hạn.
func TopLiked(counts map[string]int) string {
	best, bn := "", 0
	names := make([]string, 0, len(counts))
	for ch := range counts {
		names = append(names, ch)
	}
	sort.Strings(names)
	for _, ch := range names {
		if counts[ch] > bn {
			best, bn = ch, counts[ch]
		}
	}
	return best
}

// MyLike: nhân vật mà `guestID` đang thích ("" nếu chưa thích ai).
func (s *Store) MyLike(guestID string) (string, error) {
	var ch string
	err := s.db.QueryRow(`SELECT character FROM character_likes WHERE guest_id = ?`, guestID).Scan(&ch)
	if err != nil {
		if isNoRows(err) {
			return "", nil
		}
		return "", err
	}
	return ch, nil
}

// CanLike: đã chơi ít nhất một mạng ≥ `minSecs` giây chưa (§6.6, chống spam).
//
// Đọc từ chính bảng `lives` thay vì một cờ riêng: ai chơi thật thì bảng đó đã có dòng của
// họ, còn một cờ riêng là thêm một thứ phải nhớ ghi đúng lúc.
func (s *Store) CanLike(guestID string, minSecs float64) (bool, error) {
	var n int
	err := s.db.QueryRow(`SELECT COUNT(*) FROM lives WHERE guest_id = ? AND survive_secs >= ?`,
		guestID, minSecs).Scan(&n)
	return n > 0, err
}

// SetLike ghi/đổi/bỏ lượt thích. `char` rỗng = bỏ thích.
//
// Trả về lỗi `ErrLikeNotAllowed` khi chưa đủ điều kiện, `ErrLikeIPLimit` khi một địa chỉ IP
// đã có quá `ipMax` lượt thích còn hiệu lực trong 24 giờ. Hai lỗi riêng vì client nói hai
// câu khác nhau với người chơi.
func (s *Store) SetLike(guestID, char, ipHash string, now time.Time, ipMax int, minSecs float64) error {
	ok, err := s.CanLike(guestID, minSecs)
	if err != nil {
		return err
	}
	if !ok {
		return ErrLikeNotAllowed
	}
	defer s.invalidateLikes()
	if char == "" {
		_, err := s.db.Exec(`DELETE FROM character_likes WHERE guest_id = ?`, guestID)
		return err
	}
	// Trần theo IP tính trên những lượt CÒN HIỆU LỰC trong 24 giờ, KHÔNG tính dòng của chính
	// người này: đổi ý từ nhân vật A sang B không được coi là một lượt thích mới.
	var n int
	since := now.Add(-24 * time.Hour).Unix()
	if err := s.db.QueryRow(`SELECT COUNT(*) FROM character_likes WHERE ip_hash = ? AND at >= ? AND guest_id <> ?`,
		ipHash, since, guestID).Scan(&n); err != nil {
		return err
	}
	if n >= ipMax {
		return ErrLikeIPLimit
	}
	_, err = s.db.Exec(`INSERT INTO character_likes (guest_id, character, ip_hash, at) VALUES (?,?,?,?)
		ON CONFLICT(guest_id) DO UPDATE SET character = excluded.character, ip_hash = excluded.ip_hash, at = excluded.at`,
		guestID, char, ipHash, now.Unix())
	return err
}

func (s *Store) invalidateLikes() {
	s.likes.mu.Lock()
	s.likes.count = nil
	s.likes.mu.Unlock()
}

// Lỗi của `SetLike` — người gọi (HTTP) đổi thành mã trạng thái và câu chữ cho người chơi.
var (
	ErrLikeNotAllowed = fmt.Errorf("chưa chơi đủ một mạng")
	ErrLikeIPLimit    = fmt.Errorf("quá nhiều lượt thích từ cùng một mạng")
)

// isNoRows tránh phải import `database/sql` chỉ để so một lỗi ở nhiều chỗ.
func isNoRows(err error) bool { return err != nil && err.Error() == "sql: no rows in result set" }
