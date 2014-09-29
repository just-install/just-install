@echo OFF

mkdir C:\gopath
set GOPATH=C:\gopath

go get .\...
go build just-install.go || exit /b 1
.\just-install.exe checklinks || exit /b 1
