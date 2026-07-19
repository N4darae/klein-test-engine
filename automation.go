package main

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/go-vgo/robotgo"
)

// AutoConfig gom các tham số cho luồng auto-login.
type AutoConfig struct {
	LauncherPath string  // đường dẫn KleinNetwork.exe, dùng để mở nếu chưa chạy
	LauncherProc string  // tên tiến trình launcher để check PID
	GameProc     string  // tên tiến trình game (MapleStory)
	AssetsDir    string  // thư mục chứa ảnh mẫu (start.png, play.png, ...)
	OutDir       string  // thư mục lưu record (screenshot + log)
	Threshold    float32 // ngưỡng match ảnh
}

// DefaultAutoConfig trả về cấu hình mặc định hợp lý cho máy này.
func DefaultAutoConfig() AutoConfig {
	return AutoConfig{
		LauncherProc: "KleinNetwork",
		GameProc:     "MapleStory",
		AssetsDir:    "assets",
		OutDir:       "records",
		Threshold:    0.8,
	}
}

// waitForImage poll tìm template trên màn hình tới khi thấy (conf >= threshold)
// hoặc hết timeout. Ghi log tiến trình qua rec.
func waitForImage(tpl string, threshold float32, timeout, interval time.Duration, rec *Recorder) (Match, bool) {
	base := filepath.Base(tpl)
	deadline := time.Now().Add(timeout)
	var last Match
	for {
		m, ok, err := FindOnScreen(tpl, threshold)
		if err != nil {
			rec.Logf("tìm %s lỗi: %v", base, err)
		} else {
			last = m
			if ok {
				rec.Logf("thấy %s tại (%d,%d) conf=%.3f", base, m.X, m.Y, m.Confidence)
				return m, true
			}
		}
		if time.Now().After(deadline) {
			rec.Logf("hết %s chờ %s (conf cuối=%.3f)", timeout, base, last.Confidence)
			return last, false
		}
		robotgo.MilliSleep(int(interval / time.Millisecond))
	}
}

// clickAt di chuột tới tâm match rồi click trái, có log lại.
func clickAt(m Match, rec *Recorder) {
	robotgo.Move(m.X, m.Y)
	robotgo.MilliSleep(80)
	robotgo.Click("left", false)
	rec.Logf("click tại (%d,%d)", m.X, m.Y)
}

// Offset từ tâm card (play_card.png) xuống tâm nút Play.
// Nút Play có hiệu ứng shine động nên không match template được; ta crop card
// PHẦN TĨNH (Lv/tên/stats/INVISIBLE LOGIN, bỏ nút Play) để match ổn định rồi
// tính xuống nút Play. Client MapleStory cố định kích thước nên offset không đổi.
const playOffsetX, playOffsetY = 0, 108

// clickStep chờ template xuất hiện rồi click vào tâm. Chụp screenshot trước/sau.
func clickStep(rec *Recorder, img string, threshold float32, timeout time.Duration) error {
	name := filepath.Base(img)
	m, ok := waitForImage(img, threshold, timeout, 1*time.Second, rec)
	if !ok {
		rec.Shot("khong-thay-" + name)
		return fmt.Errorf("không thấy %s", name)
	}
	clickAt(m, rec)
	rec.Shot("sau-click-" + name)
	return nil
}

// RunPhase1 tự đăng nhập vào game (đúng flow thực tế của KleinNetwork/MapleStory):
//  1. đảm bảo launcher chạy (check PID, chưa có thì mở as admin);
//  2. click START -> chọn world Bera -> vào channel;
//  3. đợi màn chọn nhân vật (mốc INVISIBLE LOGIN) rồi click Play.
//
// Mỗi bước chụp screenshot + ghi log vào records/<timestamp>/.
func RunPhase1(cfg AutoConfig) error {
	rec, err := NewRecorder(cfg.OutDir)
	if err != nil {
		return err
	}
	defer rec.Close()

	a := func(n string) string { return filepath.Join(cfg.AssetsDir, n) }

	// 1. đảm bảo launcher đang chạy
	running, ids, err := procRunning(cfg.LauncherProc)
	if err != nil {
		rec.Logf("check pid lỗi: %v", err)
	}
	switch {
	case running:
		rec.Logf("launcher %q đã chạy (pid=%v)", cfg.LauncherProc, ids)
	case cfg.LauncherPath == "":
		return fmt.Errorf("launcher chưa chạy và chưa cấu hình đường dẫn (dùng -launcher hoặc env KLEIN_LAUNCHER)")
	default:
		rec.Logf("launcher chưa chạy -> mở (plain): %s", cfg.LauncherPath)
		if err := startLauncher(cfg.LauncherPath); err != nil {
			return fmt.Errorf("mở launcher: %w", err)
		}
	}
	rec.Shot("bat-dau")

	// 1b. đợi launcher sẵn sàng = nút START hiện. Cold-start launcher CEF load
	// khá lâu (~30s) nên cho timeout rộng; nếu đã chạy sẵn thì thấy ngay.
	if _, ok := waitForImage(a("start.png"), 0.72, 120*time.Second, 2*time.Second, rec); !ok {
		rec.Shot("launcher-chua-san-sang")
		return fmt.Errorf("launcher không hiện màn START (chưa load xong?)")
	}
	rec.Shot("launcher-san-sang")

	// 2. Check for Updates: click -> chờ dialog kết quả -> bấm OK.
	// Nếu có bản mới launcher sẽ tự tải; dialog cuối cùng luôn có nút OK để đóng.
	// (nút launcher hay bị cánh hoa sakura bay đè -> để ngưỡng thấp 0.72)
	if err := clickStep(rec, a("check_update.png"), 0.72, 30*time.Second); err != nil {
		rec.Logf("cảnh báo: %v (bỏ qua bước update)", err)
	} else if err := clickStep(rec, a("update_ok.png"), 0.72, 180*time.Second); err != nil {
		rec.Logf("cảnh báo: không thấy nút OK dialog update: %v", err)
	}

	// 3. launcher START -> world Bera -> vào channel
	if err := clickStep(rec, a("start.png"), 0.72, 60*time.Second); err != nil {
		return err
	}
	if err := clickStep(rec, a("world_bera.png"), 0.72, 60*time.Second); err != nil {
		return err
	}
	if err := clickStep(rec, a("go_channel.png"), 0.72, 30*time.Second); err != nil {
		return err
	}

	// 4. đợi màn chọn nhân vật (match card tĩnh) rồi click Play (offset xuống dưới)
	rec.Logf("đợi màn chọn nhân vật...")
	m, ok := waitForImage(a("play_card.png"), 0.85, 180*time.Second, 2*time.Second, rec)
	if !ok {
		rec.Shot("khong-thay-char-select")
		return fmt.Errorf("không tới được màn chọn nhân vật")
	}
	rec.Shot("truoc-click-Play")
	robotgo.Move(m.X+playOffsetX, m.Y+playOffsetY)
	robotgo.MilliSleep(80)
	robotgo.Click("left", false)
	rec.Logf("click Play tại (%d,%d)", m.X+playOffsetX, m.Y+playOffsetY)
	rec.Shot("sau-click-Play")

	rec.Logf("phase 1 hoàn tất.")
	return nil
}
