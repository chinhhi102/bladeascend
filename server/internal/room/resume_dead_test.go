package room

import (
	"testing"

	"bladeascend/server/internal/netproto"
)

// Chết trong lúc mất kết nối rồi NỐI LẠI NGAY, khi hoạt ảnh chết còn chưa chạy xong.
//
// Khác `TestBiChemChetTrongLucMatKetNoiThiThayTongKet` đúng một chỗ, và chính chỗ đó là lỗi:
// bài kia tick qua hết `death_anim_secs` rồi mới vào lại, nên nó luôn đi vào đường đã có
// `h.dead`. Bài này KHÔNG tick — đúng cảnh người chơi bị chém ngay lúc đang F5 và quay lại
// sau nửa giây.
//
// Trước 14/11, đường đó bỏ luôn cái hold: người chơi rơi vào một TRẬN MỚI và mất cả ván mà
// không hiểu vì sao — đúng điều §7.1 cấm. Lỗi chặn một lượt deploy, và chỉ phân biệt được
// với "bài đo chém hụt" nhờ dòng log của lệnh `killid`.
func TestVaoLaiNgaySauKhiChet_ThayTongKet(t *testing.T) {
	m, cfg := resumeManager(t)
	c, _, key := joinOne(t, m, "Teo")
	r := m.AllRooms()[0]
	z := r.Zone(0)
	id := c.PlayerID
	r.RemoveClient(c)
	settle(r, cfg)

	// Chém chết — rồi KHÔNG tick cho hết hoạt ảnh chết.
	r.Do(func() { z.G.DebugKill(z.G.Players[id]) })

	c2 := NewClient(&recConn{}, false)
	conn2 := c2.Conn.(*recConn)
	m.Join(c2, netproto.Join{Name: "Teo", Char: "a_hon", Resume: key})
	settle(r, cfg)

	dead := sentOfKind(t, conn2, "dead")
	if len(dead) == 0 {
		t.Fatal("§7.1: chết lúc mất kết nối rồi vào lại ngay → phải thấy màn Tổng kết")
	}
	if len(sentOfKind(t, conn2, "welcome")) != 0 {
		t.Fatal("chết rồi thì KHÔNG được ném vào trận mới (đúng lỗi 14/11)")
	}
	if c2.PlayerID != 0 {
		t.Fatal("không được gắn vào nhân vật nào")
	}
	// Màn Tổng kết phải CÓ SỐ LIỆU: gửi một màn rỗng thì người chơi vẫn mất cả ván, chỉ là
	// mất một cách im lặng hơn.
	if sv, _ := dead[0]["survive"].(float64); sv <= 0 {
		t.Fatalf("màn Tổng kết thiếu thời gian sống — số liệu ván không được dựng: %v", dead[0])
	}
	if _, ok := dead[0]["peak_exp"]; !ok {
		t.Fatalf("màn Tổng kết thiếu EXP đỉnh: %v", dead[0])
	}
	if _, still := z.G.Players[id]; still {
		t.Fatal("chết rồi thì nhân vật phải rời thế giới")
	}
}
