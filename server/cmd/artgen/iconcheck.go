package main

import (
	"fmt"
	"image"
	"log"
	"os"
	"path/filepath"
	"sort"
)

// iconRule: một icon UI phải trông thế nào sau khi xoá nền (chủ dự án chốt 1/11).
//
// Sinh ra vì `ui/crown_small.png` — đáng lẽ là "tiny golden crown icon" — về tay là CẢ MỘT
// NHÂN VẬT đội vương miện (prompt dùng nhầm `style`, phong cách vẽ người biếm hoạ). Thu về
// 24px trên bảng xếp hạng thì thành một hình người tí hon trên đầu chân dung, và không bài
// kiểm nào bắt được vì ảnh vẫn "hợp lệ": có nền trong suốt, đúng kích thước.
//
// Phép thử rẻ mà bắt đúng loại lỗi này: icon là MỘT vật thể (cắt tự động ra đúng một mảng
// chính), và dáng nó không cao hơn rộng quá `maxHW` — người đứng thì cao gấp rưỡi bề ngang,
// còn vương miện thì bè ra.
type iconRule struct {
	maxHW float64 // cao / rộng tối đa của vật thể chính
}

// iconRules theo tên job. Chỉ ghi icon có hình dáng biết trước; băng rôn, viên thuốc… đã có
// khung cố định nên không cần.
var iconRules = map[string]iconRule{
	"plate_crown_small": {maxHW: 1.0},
}

// majorBlobFrac: mảng nhỏ hơn ngần này so với mảng lớn nhất là chi tiết rời (viên ngọc,
// đốm lấp lánh) — không tính là "vật thể thứ hai".
const majorBlobFrac = 0.2

// checkIcon trả lỗi (tiếng Việt, để in ra cho người đọc) nếu ảnh không phải một icon đơn.
func checkIcon(img *image.NRGBA, rule iconRule) error {
	_, blobs := labelBand(img, img.Bounds())
	if len(blobs) == 0 {
		return fmt.Errorf("ảnh trống sau khi xoá nền")
	}
	sort.Slice(blobs, func(i, j int) bool { return blobs[i].area > blobs[j].area })
	major := 0
	for _, b := range blobs {
		if float64(b.area) >= majorBlobFrac*float64(blobs[0].area) {
			major++
		}
	}
	if major != 1 {
		return fmt.Errorf("cắt ra %d vật thể lớn, icon phải đúng 1", major)
	}
	r := blobs[0].rect
	if hw := float64(r.Dy()) / float64(r.Dx()); hw > rule.maxHW {
		return fmt.Errorf("vật thể cao gấp %.2f bề ngang (tối đa %.2f) — dáng người/nhân vật, không phải icon", hw, rule.maxHW)
	}
	return nil
}

// checkIconJob kiểm một job icon sau khi xoá nền; hỏng thì đánh dấu sinh lại.
func checkIconJob(j Job, img *image.NRGBA) {
	rule, ok := iconRules[j.Name]
	if !ok {
		return
	}
	if err := checkIcon(img, rule); err != nil {
		miscut[j.Name] = true
		log.Printf("  HỎNG %s: %v — cần sinh lại", j.Name, err)
	}
}

// iconCheckAll kiểm các icon ĐÃ CÓ trong assets/src (lệnh `artgen iconcheck`). Icon hỏng vào
// danh sách sinh lại, nên `artgen run` lần sau tự vẽ lại. Trả về số icon hỏng.
func iconCheckAll(jobs []Job, assets string) int {
	bad := 0
	for _, j := range jobs {
		rule, ok := iconRules[j.Name]
		if !ok {
			continue
		}
		img, err := loadImage(filepath.Join(assets, j.Out))
		if err != nil {
			fmt.Printf("[ ] %s: chưa có ảnh (%v)\n", j.Name, err)
			continue
		}
		prepped[j.Name] = true // đã kiểm lại: qua thì gỡ khỏi danh sách sinh lại (writeMiscut)
		if err := checkIcon(img, rule); err != nil {
			bad++
			miscut[j.Name] = true
			fmt.Printf("[✗] %s (%s): %v\n", j.Name, j.Out, err)
			continue
		}
		fmt.Printf("[✓] %s\n", j.Name)
	}
	if err := writeMiscut(assets); err != nil {
		fmt.Fprintln(os.Stderr, "ghi danh sách sinh lại:", err)
	}
	return bad
}
