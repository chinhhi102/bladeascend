package main

import (
	"strings"
)

// Thay `%BRAND_*%` trong HTML do server dựng — GAME_SPEC §20.
//
// Ba trang HTML sống trong server chứ không qua Vite: trang đăng nhập quản trị, trang số
// liệu `/stats`, và `client/admin/index.html` (tệp tĩnh, server đọc rồi trả). Cả ba đều
// hiện tên game, nên cả ba cần cùng một nguồn với `index.html` của game.
//
// Dùng cùng bộ token `%BRAND_*%` với plugin `kx-brand` bên Vite — một cú pháp để nhớ, và
// grep `%BRAND_` ra được mọi chỗ hiển thị tên game trong cả repo.
func (s *server) brandHTML(html string) string {
	b := s.cfg.Brand
	return strings.NewReplacer(
		"%BRAND_NAME%", b.Name,
		"%BRAND_NAME_VI%", b.NameVI,
		"%BRAND_TAGLINE_VI%", b.TaglineVI,
		"%BRAND_DOMAIN%", b.Domain,
	).Replace(html)
}
