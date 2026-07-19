package main

import (
	"fmt"
	"image/png"
	"os"
	"path/filepath"
	"time"

	"github.com/go-vgo/robotgo"
)

// Recorder ghi lại quá trình auto: screenshot theo mốc + log text, tất cả
// nằm trong 1 thư mục run đặt theo timestamp để mỗi lần chạy tách riêng.
type Recorder struct {
	dir string
	log *os.File
	n   int
}

// NewRecorder tạo thư mục run "<base>/<timestamp>" và mở file run.log trong đó.
func NewRecorder(base string) (*Recorder, error) {
	stamp := time.Now().Format("20060102-150405")
	dir := filepath.Join(base, stamp)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("tạo thư mục record: %w", err)
	}
	lf, err := os.Create(filepath.Join(dir, "run.log"))
	if err != nil {
		return nil, fmt.Errorf("tạo log: %w", err)
	}
	r := &Recorder{dir: dir, log: lf}
	r.Logf("=== bắt đầu run, lưu tại %s ===", dir)
	return r, nil
}

// Close đóng file log (an toàn gọi kể cả khi log nil).
func (r *Recorder) Close() {
	if r.log != nil {
		r.Logf("=== kết thúc run ===")
		r.log.Close()
	}
}

// Logf ghi 1 dòng log kèm timestamp ra cả stdout lẫn file run.log.
func (r *Recorder) Logf(format string, a ...any) {
	line := fmt.Sprintf("[%s] %s", time.Now().Format("15:04:05.000"), fmt.Sprintf(format, a...))
	fmt.Println(line)
	if r.log != nil {
		fmt.Fprintln(r.log, line)
	}
}

// Shot chụp full màn hình lưu thành PNG đánh số tăng dần kèm nhãn action,
// để xem lại đúng trình tự các bước auto đã làm.
func (r *Recorder) Shot(label string) {
	img, err := robotgo.CaptureImg()
	if err != nil {
		r.Logf("shot lỗi (%s): %v", label, err)
		return
	}
	r.n++
	name := fmt.Sprintf("%03d_%s.png", r.n, sanitize(label))
	f, err := os.Create(filepath.Join(r.dir, name))
	if err != nil {
		r.Logf("tạo file shot lỗi: %v", err)
		return
	}
	defer f.Close()
	if err := png.Encode(f, img); err != nil {
		r.Logf("encode shot lỗi: %v", err)
		return
	}
	r.Logf("shot -> %s (%s)", name, label)
}

// sanitize đổi ký tự lạ trong nhãn thành '_' để đặt tên file an toàn.
func sanitize(s string) string {
	out := make([]rune, 0, len(s))
	for _, c := range s {
		switch {
		case c >= 'a' && c <= 'z', c >= 'A' && c <= 'Z', c >= '0' && c <= '9', c == '-', c == '_':
			out = append(out, c)
		default:
			out = append(out, '_')
		}
	}
	return string(out)
}
