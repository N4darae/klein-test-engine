package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("usage: klein-network-test-engine <template.png> [threshold]")
		fmt.Println("  tìm ảnh mẫu trên màn hình rồi click vào tâm.")
		os.Exit(1)
	}
	tpl := os.Args[1]
	var threshold float32 = 0.8

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
