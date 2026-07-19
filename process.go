package main

import (
	"fmt"
	"os/exec"

	"github.com/go-vgo/robotgo"
)

// procRunning kiểm tra có tiến trình nào tên khớp (subset, không phân biệt
// hoa thường) đang chạy không. Trả về danh sách PID tìm được.
func procRunning(name string) (bool, []int, error) {
	ids, err := robotgo.FindIds(name)
	if err != nil {
		return false, nil, fmt.Errorf("tìm pid %q: %w", name, err)
	}
	return len(ids) > 0, ids, nil
}

// startLauncher mở launcher qua PowerShell Start-Process (KHÔNG -Verb RunAs).
// Process con chạy độc lập, kế thừa mức quyền của tiến trình auto: nếu auto
// đang chạy as admin thì launcher cũng elevated -> click/phím vào được cửa sổ
// của nó (Windows UIPI).
//
// LƯU Ý: KHÔNG dùng -Verb RunAs khi bản thân auto đã elevated — double-elevation
// làm launcher CEF này sinh process lặp và không hiện cửa sổ. Muốn game elevated,
// hãy chạy klein.exe as admin (PowerShell as admin) rồi để nó mở launcher plain.
func startLauncher(path string) error {
	cmd := exec.Command("powershell", "-NoProfile", "-NonInteractive", "-Command",
		fmt.Sprintf("Start-Process -FilePath %q", path))
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("Start-Process: %w: %s", err, out)
	}
	return nil
}
