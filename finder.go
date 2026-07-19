package main

import (
	"fmt"

	"github.com/go-vgo/robotgo"
	"gocv.io/x/gocv"
)

// Match là kết quả tìm được của một template trên màn hình.
type Match struct {
	X, Y       int     // tâm vùng khớp (toạ độ màn hình)
	W, H       int     // kích thước template
	Confidence float32 // độ tin cậy 0..1 (TM_CCOEFF_NORMED)
}

// FindInMat tìm template trong một ảnh screen (Mat RGB) đã có sẵn.
// Đây là phần lõi thuần OpenCV, tách riêng để test được không phụ thuộc
// vào việc chụp màn hình. tpl phải cùng thứ tự kênh (RGB) với screen.
func FindInMat(screen, tpl gocv.Mat, threshold float32) (Match, bool) {
	res := gocv.NewMat()
	defer res.Close()
	mask := gocv.NewMat()
	defer mask.Close()
	gocv.MatchTemplate(screen, tpl, &res, gocv.TmCcoeffNormed, mask)

	_, maxVal, _, maxLoc := gocv.MinMaxLoc(res)
	w, h := tpl.Cols(), tpl.Rows()
	m := Match{
		X:          maxLoc.X + w/2,
		Y:          maxLoc.Y + h/2,
		W:          w,
		H:          h,
		Confidence: maxVal,
	}
	return m, maxVal >= threshold
}

// FindOnScreen chụp toàn màn hình rồi dùng OpenCV MatchTemplate để tìm
// ảnh mẫu templatePath. Trả về ok=false nếu độ tin cậy dưới threshold.
// threshold gợi ý: 0.8 (khớp gần như tuyệt đối), 0.6 (lỏng hơn).
func FindOnScreen(templatePath string, threshold float32) (Match, bool, error) {
	// 1. chụp màn hình -> image.Image -> Mat (RGB)
	shot, err := robotgo.CaptureImg()
	if err != nil {
		return Match{}, false, fmt.Errorf("capture screen: %w", err)
	}
	screen, err := gocv.ImageToMatRGB(shot)
	if err != nil {
		return Match{}, false, fmt.Errorf("capture->mat: %w", err)
	}
	defer screen.Close()

	// 2. đọc template (IMRead trả về BGR) rồi đổi sang RGB cho khớp screen
	tpl := gocv.IMRead(templatePath, gocv.IMReadColor)
	if tpl.Empty() {
		return Match{}, false, fmt.Errorf("không đọc được template: %s", templatePath)
	}
	defer tpl.Close()
	gocv.CvtColor(tpl, &tpl, gocv.ColorBGRToRGB)

	// 3. matching
	m, ok := FindInMat(screen, tpl, threshold)
	return m, ok, nil
}

// FindAndClick tìm template rồi click chuột trái vào tâm nếu đạt ngưỡng.
func FindAndClick(templatePath string, threshold float32) (Match, bool, error) {
	m, ok, err := FindOnScreen(templatePath, threshold)
	if err != nil || !ok {
		return m, ok, err
	}
	robotgo.Move(m.X, m.Y)
	robotgo.MilliSleep(80)
	robotgo.Click("left", false)
	return m, true, nil
}
