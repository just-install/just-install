set GOPATH=C:\go\gopath

go build "just-install.go" || exit /b 1
"%ProgramFiles%\Microsoft SDKs\Windows\v6.0A\bin\mt.exe" -manifest "just-install.exe.manifest" -outputresource:"just-install.exe;1" || exit /b 1
misc\tools\upx.exe -9 just-install.exe || exit /b 1
