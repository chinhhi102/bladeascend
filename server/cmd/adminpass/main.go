// Command adminpass sinh hash bcrypt cho mật khẩu trang quản trị (ADMIN_SPEC §1).
//
//	go run ./cmd/adminpass
//
// Nhập mật khẩu, nhận lại một dòng `ADMIN_PASSWORD_HASH=...` để dán vào biến môi trường của
// server. Mật khẩu KHÔNG bao giờ nằm trong code hay config — chỉ cái hash này.
package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"golang.org/x/crypto/bcrypt"

	"kiemxoay/server/internal/admin"
)

func main() {
	fmt.Fprint(os.Stderr, "Mật khẩu quản trị: ")
	// Đọc từ stdin thay vì cờ dòng lệnh: cờ nằm lại trong lịch sử shell và trong `ps`.
	sc := bufio.NewScanner(os.Stdin)
	if !sc.Scan() {
		fmt.Fprintln(os.Stderr, "không đọc được mật khẩu")
		os.Exit(1)
	}
	pw := strings.TrimSpace(sc.Text())
	if len(pw) < 12 {
		fmt.Fprintln(os.Stderr, "mật khẩu phải dài ít nhất 12 ký tự")
		os.Exit(1)
	}
	h, err := bcrypt.GenerateFromPassword([]byte(pw), admin.MinBcryptCost)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Printf("ADMIN_PASSWORD_HASH=%s\n", h)
}
