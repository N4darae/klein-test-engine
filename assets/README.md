# assets — ảnh mẫu cho auto-login

Đặt các file PNG crop từ chính màn hình máy bạn vào đây. Auto (`klein auto`)
sẽ tìm các ảnh này để biết click chỗ nào.

| File | Là gì | Dùng ở bước |
|---|---|---|
| `check_update.png` | Link **Check for Updates** trên launcher | click để lấy update |
| `update_ok.png`    | Nút **OK** dialog kết quả update         | đóng dialog sau khi check |
| `start.png`     | Nút **START!!** trên launcher KleinNetwork | click khởi động game |
| `world_bera.png`| Nút world **Bera** (màn chọn World)        | click chọn world |
| `go_channel.png`| Nút **Go to Selected Channel**             | click vào channel |
| `play_card.png` | **Card nhân vật** (Lv/tên/stats, BỎ nút Play) | mốc "char load xong"; click Play theo offset bên dưới |

> Nút **Play** có hiệu ứng shine động nên không match template trực tiếp được.
> Ta match phần tĩnh của card (`play_card.png`) rồi click xuống nút Play
> (offset `playOffsetX/Y` trong `automation.go`).

## Cách tạo ảnh mẫu

1. Mở launcher / vào tới màn có nút cần crop.
2. Chụp full màn hình:
   ```powershell
   . .\env.ps1
   .\klein.exe capture shot.png
   ```
3. Mở `shot.png`, crop sát vùng nút (chỉ lấy phần nét, đặc trưng — tránh lấy
   nhiều nền trơn), lưu thành `assets/start.png`, `assets/play.png`.

Crop càng sát và nhiều texture thì match càng chắc. Nếu match nhầm/không thấy,
chỉnh ngưỡng bằng `-th` (vd `-th 0.7`).
