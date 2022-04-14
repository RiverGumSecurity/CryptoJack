#
# Makefile for golang project
# Author: Joff Thyer
#
TARGET=rbot
LIB=../cjlib
SRC=$(TARGET).go
LDFLAGS="-s -w"
ifeq ($(OS),Windows_NT)
	RM=del /f
	BIN=$(TARGET).exe
else
	RM=rm -f
	BIN=$(TARGET)
endif

$(BIN): $(LIB) $(SRC)
	go build -ldflags=$(LDFLAGS) -trimpath -o $(BIN) $(SRC)

clean:
	$(RM) $(BIN)
