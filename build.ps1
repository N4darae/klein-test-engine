# Build klein.exe với môi trường OpenCV. Dùng: .\build.ps1
. "$PSScriptRoot\env.ps1"
go build -tags customenv -o klein.exe .
if ($LASTEXITCODE -eq 0) { Write-Host "OK -> klein.exe" } else { Write-Host "BUILD FAILED" }
