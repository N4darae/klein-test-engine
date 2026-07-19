package main

import (
	"fmt"
	"os/exec"
	"strings"

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

// startLauncher mở launcher với quyền admin (game cần admin để chạy), nhưng
// TỰ PHÁT HIỆN mức quyền hiện tại để tránh double-elevation:
//   - klein.exe ĐÃ chạy as admin  -> Start-Process plain (con kế thừa elevated).
//     Nếu dùng -Verb RunAs ở đây, launcher CEF sẽ sinh process lặp + không hiện
//     cửa sổ.
//   - klein.exe CHƯA admin        -> Start-Process -Verb RunAs (bật UAC nâng quyền)
//     để launcher (và game) được elevated.
// Nhờ vậy game luôn chạy as admin dù klein.exe chạy ở mức quyền nào.
func startLauncher(path string) error {
	p := strings.ReplaceAll(path, "'", "''") // escape cho chuỗi single-quote PS
	ps := fmt.Sprintf(
		`$p='%s';`+
			`$admin=(New-Object Security.Principal.WindowsPrincipal([Security.Principal.WindowsIdentity]::GetCurrent())).IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator);`+
			`if($admin){Start-Process -FilePath $p}else{Start-Process -FilePath $p -Verb RunAs}`, p)
	cmd := exec.Command("powershell", "-NoProfile", "-NonInteractive", "-Command", ps)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("mở launcher: %w: %s", err, out)
	}
	return nil
}
