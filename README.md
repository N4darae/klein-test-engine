# klein-network-test-engine

Image-find + click trên màn hình bằng Go, dùng **robotgo** (chụp màn hình + điều khiển
chuột/bàn phím) và **gocv/OpenCV** (`MatchTemplate`) để tìm ảnh mẫu.

## Môi trường (đã verify trên máy này)

| Thành phần | Phiên bản / vị trí |
|---|---|
| Go | 1.26.2 windows/amd64, CGO bật |
| gcc/g++ | 14.2.0 — `C:\mingw64-x64\w64devkit\bin` |
| OpenCV | **4.10.0** (MinGW build) — `C:\opencv\build\install` |
| Tesseract | 5.5.0 — `C:\Program Files\Tesseract-OCR` (chỉ cần nếu làm OCR) |
| robotgo | v1.0.2 |
| gocv | v0.41.0 |

## Build & chạy

```powershell
. .\env.ps1        # set PATH + CGO flags cho OpenCV
.\build.ps1        # -> klein.exe
.\klein.exe target.png        # tìm target.png trên màn hình rồi click vào tâm
```

Test:
```powershell
. .\env.ps1
go test -tags customenv -v .
```

## Lưu ý quan trọng

- **Phải build với `-tags customenv`.** gocv v0.41 hardcode tên lib OpenCV `4110`
  (4.11.0), còn máy cài `4100` (4.10.0) → link lỗi nếu dùng flag mặc định. `env.ps1`
  set `CGO_CPPFLAGS`/`CGO_LDFLAGS` tay trỏ đúng lib 4100.
- **Lúc CHẠY cần `libopencv_*4100.dll`** trên PATH (`env.ps1` đã thêm thư mục bin).
- **KHÔNG dùng `github.com/vcaesar/gcv`** — nó phụ thuộc `github.com/go-vgo/gt` đã bị
  gỡ khỏi GitHub nên không build được. Ta gọi thẳng `gocv.MatchTemplate`.
- `robotgo.CaptureImg(x, y, w, h)` chụp 1 vùng; không tham số = full màn hình.
- Ngưỡng (`threshold`) gợi ý: **0.8** cho UI thật có nét; nền trơn/ít texture sẽ cho
  confidence thấp và dễ khớp nhầm.

## OCR (nếu cần sau này)

Tesseract.exe đã cài nhưng **thiếu dev headers** → binding Go `gosseract` chưa build
được. Nếu cần đọc text: hoặc cài thêm header leptonica/tesseract, hoặc gọi
`tesseract.exe` qua `os/exec`.

## File

- `finder.go` — `FindInMat` (lõi matching), `FindOnScreen`, `FindAndClick`.
- `main.go` — CLI demo.
- `finder_test.go` — test tách lõi matching (same-frame, conf≈1.0) khỏi biến thiên chụp lại.
- `env.ps1` / `build.ps1` — cấu hình môi trường + build.
