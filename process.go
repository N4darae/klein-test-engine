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

// startAsAdmin mở exe với quyền admin qua PowerShell Start-Process -Verb RunAs
// (sẽ bật hộp thoại UAC). Trả về khi lệnh PowerShell kết thúc; process con
// (launcher) chạy độc lập sau đó.
//
// LƯU Ý: launcher/game chạy elevated thì tiến trình auto (klein.exe) cũng phải
// chạy as admin mới gửi được click/phím vào cửa sổ của chúng (Windows UIPI).
func startAsAdmin(path string) error {
	cmd := exec.Command("powershell", "-NoProfile", "-NonInteractive", "-Command",
		fmt.Sprintf("Start-Process -FilePath %q -Verb RunAs", path))
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("Start-Process RunAs: %w: %s", err, out)
	}
	return nil
}
