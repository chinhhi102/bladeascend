package main

import (
	"image"
	"image/color"
	"testing"
)

// paint tô các hình chữ nhật đặc lên ảnh trong suốt 128×128.
func paint(rects ...image.Rectangle) *image.NRGBA {
	img := image.NewNRGBA(image.Rect(0, 0, 128, 128))
	for _, r := range rects {
		for y := r.Min.Y; y < r.Max.Y; y++ {
			for x := r.Min.X; x < r.Max.X; x++ {
				img.SetNRGBA(x, y, color.NRGBA{255, 200, 0, 255})
			}
		}
	}
	return img
}

func TestIconCheckAcceptsCrownShape(t *testing.T) {
	// Vương miện: bè ngang, kèm một viên ngọc nhỏ rời — chi tiết nhỏ không tính là vật thứ hai.
	crown := paint(image.Rect(14, 44, 114, 92), image.Rect(60, 30, 66, 36))
	if err := checkIcon(crown, iconRules["plate_crown_small"]); err != nil {
		t.Fatalf("hình vương miện bị từ chối: %v", err)
	}
}

func TestIconCheckRejectsPersonShape(t *testing.T) {
	// Người đứng (đầu + thân liền nhau): cao gấp đôi bề ngang — đúng kiểu ảnh hỏng 1/11.
	person := paint(image.Rect(48, 6, 80, 38), image.Rect(40, 38, 88, 122))
	if err := checkIcon(person, iconRules["plate_crown_small"]); err == nil {
		t.Fatal("dáng người lọt qua bài kiểm icon vương miện")
	}
}

func TestIconCheckRejectsTwoObjects(t *testing.T) {
	two := paint(image.Rect(4, 40, 60, 80), image.Rect(68, 40, 124, 80))
	if err := checkIcon(two, iconRules["plate_crown_small"]); err == nil {
		t.Fatal("hai vật thể lớn lọt qua bài kiểm icon")
	}
}
