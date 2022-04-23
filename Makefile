#
# Makefile for cryptojack
# version 1.0, Joff Thyer
#
export GO111MODULE=on

all: encrypt decrypt fakedata rbot
	@echo [*] Building encrypt, decrypt, fakedata, and rbot
	@go build -v -trimpath -ldflags="-w -s" ./encrypt
	@go build -v -trimpath -ldflags="-w -s" ./decrypt
	@go build -v -trimpath -ldflags="-w -s" ./fakedata
	@go build -v -trimpath -ldflags="-w -s" ./rbot
	@echo =================================================
	@echo   Successfully compiled the CryptoJack Project! 
	@echo =================================================

7zip:
	7z a CryptoJack-V1.0.1.7z *.exe yaml/*.enc

clean:
	@del *.exe
