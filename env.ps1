# Dot-source file này để set môi trường build/run cho gocv + OpenCV.
#   . .\env.ps1
# OpenCV 4.10.0 (MinGW build) tại C:\opencv\build\install

$OpenCV = "C:\opencv\build\install"
$OpenCVBin = "$OpenCV\x64\mingw\bin"     # chứa libopencv_*4100.dll (cần lúc CHẠY)
$OpenCVLib = "$OpenCV\x64\mingw\lib"
$OpenCVPkg = "$OpenCVLib\pkgconfig"

# Toolchain MinGW (gcc/g++ + pkg-config)
$MinGWBin = "C:\mingw64-x64\w64devkit\bin"

$env:PATH = "$OpenCVBin;$MinGWBin;$env:PATH"
$env:PKG_CONFIG_PATH = $OpenCVPkg

# gocv v0.41 hardcode tên lib 4110; OpenCV ở đây là 4100 -> phải dùng
# -tags customenv và tự set CGO flags (path tuyệt đối, KHÔNG qua pkg-config
# vì pkg-config tự define-prefix sai đường dẫn).
$libs = & pkg-config --libs-only-l opencv4
$env:CGO_CPPFLAGS = "-I$($OpenCV -replace '\\','/')/include"
$env:CGO_LDFLAGS  = "-L$($OpenCVLib -replace '\\','/') $libs"
$env:GOFLAGS = "-tags=customenv"

Write-Host "OpenCV env ready:" (& pkg-config --modversion opencv4)
