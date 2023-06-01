#
# Windows Makefile for cryptojack
# version 1.0, Joff Thyer
#
export GO111MODULE=on

all:
	@echo [*] Building encrypt, decrypt, fakedata, and rbot
	@if not exist "bin" mkdir bin
	@cd bin && go build -v -trimpath -ldflags="-w -s" ../encrypt
	@cd bin && go build -v -trimpath -ldflags="-w -s" ../decrypt
	@cd bin && go build -v -trimpath -ldflags="-w -s" ../fakedata
	@cd bin && go build -v -trimpath -ldflags="-w -s" ../rbot
	@echo =================================================
	@echo   Successfully compiled the CryptoJack Project! 
	@echo =================================================

release: all
	@python scripts\EncryptYAML.py yaml
	@if not exist "Release\bin" mkdir Release\bin
	@if not exist "Release\yaml" mkdir Release\yaml
	@copy bin\*.exe Release\bin
	@copy yaml\*.enc Release\yaml
	@cd Release && zip CryptoJack.zip bin\*.exe yaml\*.enc
	@rd /q /s Release\bin
	@rd /q /s Release\yaml

clean:
	@del bin\*.exe
	@del Release\CryptoJack.zip
