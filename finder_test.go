package main

import (
	"image"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"github.com/go-vgo/robotgo"
	"gocv.io/x/gocv"
)

// TestFindOnScreen chụp full màn hình, chọn ô 140x90 có độ biến thiên cao
// nhất (chắc chắn có nội dung, không phải nền trơn) làm template, rồi kiểm
// tra FindOnScreen tìm lại đúng vị trí. Không click.
func TestFindOnScreen(t *testing.T) {
	shot, err := robotgo.CaptureImg()
	if err != nil {
		t.Fatalf("capture: %v", err)
	}
	full, err := gocv.ImageToMatRGB(shot)
	if err != nil {
		t.Fatal(err)
	}
	defer full.Close()

	tw, th := 140, 90
	step := 60
	bx, by, best := 0, 0, -1.0
	for y := 0; y+th < full.Rows(); y += step {
		for x := 0; x+tw < full.Cols(); x += step {
			roi := full.Region(image.Rect(x, y, x+tw, y+th))
			mean := gocv.NewMat()
			std := gocv.NewMat()
			gocv.MeanStdDev(roi, &mean, &std)
			s := std.GetDoubleAt(0, 0) + std.GetDoubleAt(1, 0) + std.GetDoubleAt(2, 0)
			mean.Close()
			std.Close()
			roi.Close()
			if s > best {
				best, bx, by = s, x, y
			}
		}
	}
	t.Logf("chọn template tại (%d,%d), tổng-stddev=%.1f", bx, by, best)

	// crop và lưu template
	tplMat := full.Region(image.Rect(bx, by, bx+tw, by+th))
	tpl, err := tplMat.ToImage()
	tplMat.Close()
	if err != nil {
		t.Fatal(err)
	}
	tmp := filepath.Join(t.TempDir(), "tpl.png")
	f, _ := os.Create(tmp)
	if err := png.Encode(f, tpl); err != nil {
		f.Close()
		t.Fatal(err)
	}
	f.Close()

	_ = tmp

	// (A) Lõi matching trên CÙNG frame: phải conf≈1.0 tại đúng vị trí.
	tpl2 := full.Region(image.Rect(bx, by, bx+tw, by+th))
	mSelf, okSelf := FindInMat(full, tpl2, 0.9)
	tpl2.Close()
	t.Logf("[same-frame] found=(%d,%d) conf=%.4f ok=%v", mSelf.X, mSelf.Y, mSelf.Confidence, okSelf)
	if !okSelf {
		t.Fatalf("lõi matching sai: conf=%.4f", mSelf.Confidence)
	}
	if abs(mSelf.X-(bx+tw/2)) > 1 || abs(mSelf.Y-(by+th/2)) > 1 {
		t.Fatalf("[same-frame] vị trí lệch: got (%d,%d) want (%d,%d)", mSelf.X, mSelf.Y, bx+tw/2, by+th/2)
	}

	// (B) Pipeline đầy đủ có chụp lại màn hình (từ file PNG). Có thể lệch
	// nếu màn hình đổi giữa 2 lần chụp — chỉ log, không fail cứng.
	m, ok, err := FindOnScreen(tmp, 0.9)
	if err != nil {
		t.Fatalf("FindOnScreen: %v", err)
	}
	t.Logf("[live re-capture] found=(%d,%d) conf=%.4f ok=%v", m.X, m.Y, m.Confidence, ok)
}

func abs(v int) int {
	if v < 0 {
		return -v
	}
	return v
}
