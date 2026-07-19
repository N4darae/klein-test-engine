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

## Auto-login (phase 1)

Tự đăng nhập vào game: đảm bảo launcher chạy → click **START** → đợi nhân vật
load xong → click **Play**. Mỗi bước chụp screenshot + ghi log vào
`records/<timestamp>/`.

```powershell
. .\env.ps1
# 1) tạo ảnh mẫu (crop nút START, Play từ ảnh này -> assets/start.png, assets/play.png)
.\klein.exe capture shot.png
# 2) chạy auto (mở launcher as admin nếu chưa chạy)
.\klein.exe auto -launcher "C:\path\to\KleinNetwork.exe"
```

- Flow thực tế: **START → chọn world Bera → vào channel → màn chọn nhân vật → Play**.
- Cần 4 ảnh mẫu trong `assets/`: `start.png`, `world_bera.png`, `go_channel.png`,
  `play_card.png` — xem `assets/README.md`.
- Nút **Play** có shine động → match card tĩnh (`play_card.png`) rồi click Play theo offset.
- Nếu launcher đã mở sẵn thì bỏ `-launcher` cũng được (chỉ cần khi phải tự mở).
- Ngưỡng match chỉnh bằng `-th` (mặc định 0.8); `probe` để đo confidence trước.

> **QUAN TRỌNG — quyền admin:** nếu launcher/game chạy **as admin** (elevated),
> thì `klein.exe` cũng phải chạy **as admin** mới gửi được chuột/phím vào cửa sổ
> của chúng (cơ chế UIPI của Windows). Mở PowerShell as admin rồi chạy `klein auto`.

### Lệnh con

| Lệnh | Việc |
|---|---|
| `klein auto [-launcher .. -assets .. -out .. -th ..]` | chạy auto-login phase 1 |
| `klein capture [out.png]` | chụp full màn hình ra file (để crop ảnh mẫu) |
| `klein probe <template.png> [threshold]` | in vị trí + confidence, KHÔNG click (để tune) |
| `klein find <template.png> [threshold]` | tìm ảnh rồi click (hành vi cũ) |

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
- `automation.go` — `RunPhase1` + `AutoConfig`, `waitForImage` (luồng auto-login).
- `process.go` — `procRunning` (check PID), `startAsAdmin` (mở exe as admin).
- `recorder.go` — `Recorder`: screenshot đánh số + log ra `records/<timestamp>/`.
- `main.go` — CLI: `auto`, `capture`, `find`.
- `finder_test.go` — test tách lõi matching (same-frame, conf≈1.0) khỏi biến thiên chụp lại.
- `env.ps1` / `build.ps1` — cấu hình môi trường + build.
