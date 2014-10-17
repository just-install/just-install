AI = AdvancedInstaller.com
MT = "/cygdrive/c/Program Files (x86)/Microsoft SDKs/Windows/v7.1A/Bin/mt.exe"

TIMESTAMP = $(shell date +%Y%m%d%H%S)
VERSION = 2.2.0
BUILD = $(VERSION).$(TIMESTAMP)


.PHONY: all bootstrap check checklinks clean


all: just-install.msi


bootstrap:
	-go get .\...


clean:
	rm -f just-install.exe
	rm -f just-install.msi
	rm -rf deploy/just-install-cache


check: checklinks


checklinks: just-install.exe
	./just-install.exe checklinks


just-install.msi: just-install.exe
	$(MT) -manifest deploy/just-install.exe.manifest -outputresource:"just-install.exe;1"
	deploy/upx --best just-install.exe
	cd deploy && $(AI) /edit just-install.aip /SetVersion "$(BUILD)"
	cd deploy && $(AI) /rebuild just-install.aip


just-install.exe: just-install.go
	go build just-install.go
