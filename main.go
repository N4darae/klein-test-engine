package main

import (
	"flag"
	"fmt"
	"image/png"
	"os"

	"github.com/go-vgo/robotgo"
)

func usage() {
	fmt.Println("usage:")
	fmt.Println("  klein auto   [-launcher <KleinNetwork.exe>] [-assets dir] [-out dir] [-th 0.8]")
	fmt.Println("      chạy auto-login phase 1: đảm bảo launcher chạy -> START -> đợi char -> Play")
	fmt.Println("  klein capture [out.png]")
	fmt.Println("      chụp full màn hình ra file (để crop ảnh mẫu nút START/Play)")
	fmt.Println("  klein probe <template.png> [threshold]")
	fmt.Println("      tìm ảnh mẫu, in ra vị trí + confidence (KHÔNG click) — để tune")
	fmt.Println("  klein find <template.png> [threshold]")
	fmt.Println("      tìm ảnh mẫu trên màn hình rồi click vào tâm")
}

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}
	switch os.Args[1] {
	case "auto":
		cmdAuto(os.Args[2:])
	case "capture":
		cmdCapture(os.Args[2:])
	case "probe":
		cmdProbe(os.Args[2:])
	case "find":
		cmdFind(os.Args[2:])
	case "-h", "--help", "help":
		usage()
	default:
		// tương thích ngược: `klein target.png [threshold]` -> find
		cmdFind(os.Args[1:])
	}
}

// cmdAuto chạy luồng auto-login phase 1.
func cmdAuto(args []string) {
	cfg := DefaultAutoConfig()
	fs := flag.NewFlagSet("auto", flag.ExitOnError)
	fs.StringVar(&cfg.LauncherPath, "launcher", os.Getenv("KLEIN_LAUNCHER"), "đường dẫn KleinNetwork.exe (mở nếu chưa chạy)")
	fs.StringVar(&cfg.LauncherProc, "proc", cfg.LauncherProc, "tên tiến trình launcher để check PID")
	fs.StringVar(&cfg.AssetsDir, "assets", cfg.AssetsDir, "thư mục ảnh mẫu")
	fs.StringVar(&cfg.OutDir, "out", cfg.OutDir, "thư mục lưu record")
	th := float64(cfg.Threshold)
	fs.Float64Var(&th, "th", th, "ngưỡng match ảnh (0..1)")
	fs.Parse(args)
	cfg.Threshold = float32(th)

	if err := RunPhase1(cfg); err != nil {
		fmt.Fprintln(os.Stderr, "auto lỗi:", err)
		os.Exit(2)
	}
}

// cmdCapture chụp full màn hình lưu ra file PNG (mặc định capture.png).
func cmdCapture(args []string) {
	out := "capture.png"
	if len(args) >= 1 {
		out = args[0]
	}
	img, err := robotgo.CaptureImg()
	if err != nil {
		fmt.Fprintln(os.Stderr, "capture lỗi:", err)
		os.Exit(2)
	}
	f, err := os.Create(out)
	if err != nil {
		fmt.Fprintln(os.Stderr, "tạo file lỗi:", err)
		os.Exit(2)
	}
	defer f.Close()
	if err := png.Encode(f, img); err != nil {
		fmt.Fprintln(os.Stderr, "encode lỗi:", err)
		os.Exit(2)
	}
	fmt.Printf("đã lưu %s\n", out)
}

// cmdProbe tìm template và in vị trí + confidence, KHÔNG click. Dùng để
// tune ngưỡng / kiểm tra ảnh mẫu có khớp không.
func cmdProbe(args []string) {
	if len(args) < 1 {
		fmt.Println("usage: klein probe <template.png> [threshold]")
		os.Exit(1)
	}
	tpl := args[0]
	var threshold float32 = 0.8
	if len(args) >= 2 {
		var t float64
		if _, err := fmt.Sscanf(args[1], "%f", &t); err == nil {
			threshold = float32(t)
		}
	}
	m, ok, err := FindOnScreen(tpl, threshold)
	if err != nil {
		fmt.Fprintln(os.Stderr, "lỗi:", err)
		os.Exit(2)
	}
	fmt.Printf("conf=%.3f ok=%v tâm=(%d,%d) size=%dx%d\n", m.Confidence, ok, m.X, m.Y, m.W, m.H)
}

// cmdFind giữ hành vi cũ: tìm template rồi click vào tâm.
func cmdFind(args []string) {
	if len(args) < 1 {
		fmt.Println("usage: klein find <template.png> [threshold]")
		os.Exit(1)
	}
	tpl := args[0]
	var threshold float32 = 0.8
	if len(args) >= 2 {
		var t float64
		if _, err := fmt.Sscanf(args[1], "%f", &t); err == nil {
			threshold = float32(t)
		}
	}

	m, ok, err := FindAndClick(tpl, threshold)
	if err != nil {
		fmt.Fprintln(os.Stderr, "lỗi:", err)
		os.Exit(2)
	}
	if !ok {
		fmt.Printf("KHÔNG tìm thấy (conf=%.3f < %.2f)\n", m.Confidence, threshold)
		os.Exit(3)
	}
	fmt.Printf("Đã click tại (%d,%d), conf=%.3f\n", m.X, m.Y, m.Confidence)
}
