package store

import (
	"errors"
	"testing"
	"time"
)

// GAME_SPEC §6.6 (v6.48) — lượt thích nhân vật.

func likeStore(t *testing.T) *Store {
	t.Helper()
	s, err := Open(":memory:", time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = s.Close() })
	return s
}

// played ghi một mạng đủ dài để người này được quyền thích.
func played(t *testing.T, s *Store, guest string, secs float64) {
	t.Helper()
	if err := s.Record(Life{GuestID: guest, Name: "N", Character: "a_hon", SurviveSecs: secs,
		Map: "m", EndedAt: time.Now().Unix()}); err != nil {
		t.Fatal(err)
	}
}

// "Mỗi người chơi có 1 lượt thích cho 1 nhân vật. Đổi ý: thích nhân vật khác → lượt chuyển
// sang (bỏ ở nhân vật cũ). Bấm lại để bỏ thích."
func TestLikeIsOnePerGuestAndMoves(t *testing.T) {
	s := likeStore(t)
	now := time.Now()
	played(t, s, "g1", 90)

	if err := s.SetLike("g1", "cu_rua", "ip1", now, 5, 60); err != nil {
		t.Fatal(err)
	}
	if c, _ := s.LikeCounts(now); c["cu_rua"] != 1 {
		t.Fatalf("đếm sau lần thích đầu: %v", c)
	}

	// Đổi ý: lượt CHUYỂN sang, không cộng thêm.
	if err := s.SetLike("g1", "be_ti", "ip1", now, 5, 60); err != nil {
		t.Fatal(err)
	}
	c, _ := s.LikeCounts(now.Add(2 * time.Minute))
	if c["be_ti"] != 1 || c["cu_rua"] != 0 {
		t.Fatalf("đổi ý xong: %v — muốn be_ti 1, cu_rua 0", c)
	}
	if mine, _ := s.MyLike("g1"); mine != "be_ti" {
		t.Fatalf("đang thích %q, muốn be_ti", mine)
	}

	// Bỏ thích.
	if err := s.SetLike("g1", "", "ip1", now, 5, 60); err != nil {
		t.Fatal(err)
	}
	if c, _ := s.LikeCounts(now.Add(4 * time.Minute)); c["be_ti"] != 0 {
		t.Fatalf("bỏ thích rồi vẫn đếm: %v", c)
	}
}

// "Chỉ người đã chơi ít nhất 1 mạng ≥ 60s mới thích được (chống spam)."
func TestLikeNeedsOneRealLife(t *testing.T) {
	s := likeStore(t)
	now := time.Now()

	if err := s.SetLike("g2", "a_hon", "ip", now, 5, 60); !errors.Is(err, ErrLikeNotAllowed) {
		t.Fatalf("chưa chơi mạng nào mà thích được: %v", err)
	}
	played(t, s, "g2", 59) // chơi nhưng chưa đủ 60 giây
	if err := s.SetLike("g2", "a_hon", "ip", now, 5, 60); !errors.Is(err, ErrLikeNotAllowed) {
		t.Fatalf("mạng 59 giây mà thích được: %v", err)
	}
	played(t, s, "g2", 61)
	if err := s.SetLike("g2", "a_hon", "ip", now, 5, 60); err != nil {
		t.Fatalf("chơi đủ rồi mà vẫn không thích được: %v", err)
	}
}

// "Giới hạn thêm theo IP: tối đa 5 lượt thích hoạt động / IP / 24 giờ."
func TestLikeIPLimit(t *testing.T) {
	s := likeStore(t)
	now := time.Now()
	for i := 0; i < 6; i++ {
		g := string(rune('a'+i)) + "-guest"
		played(t, s, g, 90)
		err := s.SetLike(g, "a_hon", "same-ip", now, 5, 60)
		if i < 5 && err != nil {
			t.Fatalf("lượt thứ %d từ cùng IP bị chặn sớm: %v", i+1, err)
		}
		if i == 5 && !errors.Is(err, ErrLikeIPLimit) {
			t.Fatalf("lượt thứ 6 từ cùng IP phải bị chặn, nhận: %v", err)
		}
	}
	// Người thứ 6 đổi sang IP khác thì thích được.
	if err := s.SetLike("f-guest", "a_hon", "other-ip", now, 5, 60); err != nil {
		t.Fatalf("IP khác vẫn bị chặn: %v", err)
	}
	// Và người đã thích rồi thì ĐỔI Ý được, dù IP đã đầy trần: đổi ý không phải lượt mới.
	if err := s.SetLike("a-guest", "be_ti", "same-ip", now, 5, 60); err != nil {
		t.Fatalf("đổi ý bị tính là lượt thích mới: %v", err)
	}
}

// Nhãn "Được thích nhất" + cache 60 giây.
func TestTopLikedAndCacheTTL(t *testing.T) {
	s := likeStore(t)
	now := time.Now()
	for i, ch := range []string{"a_hon", "a_hon", "be_ti"} {
		g := string(rune('a'+i)) + "-g"
		played(t, s, g, 90)
		if err := s.SetLike(g, ch, "ip"+g, now, 5, 60); err != nil {
			t.Fatal(err)
		}
	}
	c, _ := s.LikeCounts(now)
	if got := TopLiked(c); got != "a_hon" {
		t.Fatalf("được thích nhất = %q, muốn a_hon (%v)", got, c)
	}
	// Chưa ai thích thì không có nhãn.
	if got := TopLiked(map[string]int{}); got != "" {
		t.Errorf("phòng chưa ai thích mà vẫn có nhãn %q", got)
	}
	// Hoà thì lấy tên nhỏ hơn theo chữ cái — kết quả ỔN ĐỊNH, không nhảy mỗi lần cache hết hạn.
	tie := map[string]int{"zz": 3, "aa": 3}
	if got := TopLiked(tie); got != "aa" {
		t.Errorf("hoà thì phải ra tên nhỏ nhất theo chữ, nhận %q", got)
	}
}

// IP không bao giờ được lưu nguyên dạng (CLAUDE.md luật 15 — repo công khai).
func TestIPIsHashed(t *testing.T) {
	h := HashIP("203.0.113.9")
	if h == "203.0.113.9" || len(h) != 16 {
		t.Fatalf("băm IP ra %q — phải là chuỗi băm 16 ký tự hex", h)
	}
	if HashIP("203.0.113.9") != h {
		t.Error("băm cùng một IP hai lần ra hai kết quả — không đếm theo IP được")
	}
	if HashIP("203.0.113.10") == h {
		t.Error("hai IP khác nhau băm ra cùng một chuỗi")
	}
}
