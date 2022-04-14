#
# Makefile for cryptojack
# version 1.0, Joff Thyer
#
export GO111MODULE=on

all: encrypt decrypt fakedata rbot
	@echo [*] Building encrypt, decrypt, fakedata, and rbot
	@go build ./encrypt
	@go build ./decrypt
	@go build ./fakedata
	@go build ./rbot
	@echo =================================================
	@echo   Successfully compiled the CryptoJack Project! 
	@echo =================================================

clean:
	@del *.exe
