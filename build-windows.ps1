$ErrorActionPreference = "Stop"
$env:CGO_ENABLED = "0"
$env:GOOS = "windows"
$env:GOARCH = "amd64"
New-Item -ItemType Directory -Force -Path dist | Out-Null

go test ./...
go vet ./...
go build -trimpath -buildvcs=false -ldflags="-s -w -buildid= -H=windowsgui" -o dist/AuronQ-Desktop.exe ./cmd/auronq-desktop
go build -trimpath -buildvcs=false -ldflags="-s -w -buildid=" -o dist/auronq-cli.exe ./cmd/auronq
Get-FileHash dist/AuronQ-Desktop.exe -Algorithm SHA256
Get-FileHash dist/auronq-cli.exe -Algorithm SHA256
