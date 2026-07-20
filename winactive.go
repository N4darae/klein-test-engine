package main

import (
	"fmt"
	"syscall"
	"unsafe"

	"github.com/go-vgo/robotgo"
)

// Bơm foreground native qua user32. robotgo.ActivePid không đáng tin với launcher
// CEF (nó luôn return nil, chọn sai handle trong đám cửa sổ ẩn/child, và dính
// foreground-lock của Windows) — nên tự làm: tìm đúng cửa sổ chính rồi ép foreground
// bằng mẹo AttachThreadInput.
var (
	user32 = syscall.NewLazyDLL("user32.dll")

	pEnumWindows              = user32.NewProc("EnumWindows")
	pGetWindowThreadProcessId = user32.NewProc("GetWindowThreadProcessId")
	pIsWindowVisible          = user32.NewProc("IsWindowVisible")
	pGetWindowTextLengthW     = user32.NewProc("GetWindowTextLengthW")
	pGetWindowRect            = user32.NewProc("GetWindowRect")
	pIsIconic                 = user32.NewProc("IsIconic")
	pShowWindow               = user32.NewProc("ShowWindow")
	pSetForegroundWindow      = user32.NewProc("SetForegroundWindow")
	pBringWindowToTop         = user32.NewProc("BringWindowToTop")
	pGetForegroundWindow      = user32.NewProc("GetForegroundWindow")
	pAttachThreadInput        = user32.NewProc("AttachThreadInput")

	kernel32            = syscall.NewLazyDLL("kernel32.dll")
	pGetCurrentThreadID = kernel32.NewProc("GetCurrentThreadId")
)

const (
	swShow    = 5
	swRestore = 9
)

type winRect struct{ left, top, right, bottom int32 }

// findMainWindow tìm cửa sổ TOP-LEVEL của pid: visible + có tiêu đề + diện tích lớn
// nhất (CEF sinh nhiều cửa sổ ẩn/child; cửa sổ chính là cái to nhất có title).
func findMainWindow(pid uint32) syscall.Handle {
	var best syscall.Handle
	var bestArea int32
	cb := syscall.NewCallback(func(hwnd syscall.Handle, _ uintptr) uintptr {
		var wpid uint32
		pGetWindowThreadProcessId.Call(uintptr(hwnd), uintptr(unsafe.Pointer(&wpid)))
		if wpid != pid {
			return 1
		}
		if vis, _, _ := pIsWindowVisible.Call(uintptr(hwnd)); vis == 0 {
			return 1
		}
		if n, _, _ := pGetWindowTextLengthW.Call(uintptr(hwnd)); n == 0 {
			return 1
		}
		var r winRect
		pGetWindowRect.Call(uintptr(hwnd), uintptr(unsafe.Pointer(&r)))
		if area := (r.right - r.left) * (r.bottom - r.top); area > bestArea {
			bestArea = area
			best = hwnd
		}
		return 1
	})
	pEnumWindows.Call(cb, 0)
	return best
}

// activateWindow đưa cửa sổ chính của pid lên foreground (restore nếu minimize).
// Dùng AttachThreadInput để vượt hạn chế SetForegroundWindow (foreground-lock).
func activateWindow(pid int) error {
	hwnd := findMainWindow(uint32(pid))
	if hwnd == 0 {
		return fmt.Errorf("không tìm thấy cửa sổ chính của pid %d", pid)
	}
	if ic, _, _ := pIsIconic.Call(uintptr(hwnd)); ic != 0 {
		pShowWindow.Call(uintptr(hwnd), swRestore)
	} else {
		pShowWindow.Call(uintptr(hwnd), swShow)
	}

	fg, _, _ := pGetForegroundWindow.Call()
	curTid, _, _ := pGetCurrentThreadID.Call()
	var fgTid uintptr
	if fg != 0 {
		fgTid, _, _ = pGetWindowThreadProcessId.Call(fg, 0)
	}
	if fgTid != 0 && fgTid != curTid {
		pAttachThreadInput.Call(curTid, fgTid, 1)
		pBringWindowToTop.Call(uintptr(hwnd))
		pSetForegroundWindow.Call(uintptr(hwnd))
		pAttachThreadInput.Call(curTid, fgTid, 0)
	} else {
		pBringWindowToTop.Call(uintptr(hwnd))
		pSetForegroundWindow.Call(uintptr(hwnd))
	}
	robotgo.MilliSleep(400)
	return nil
}
