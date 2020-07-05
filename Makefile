GO ?= go

NOTESCAN_BIN = notescan
NOTESCAN_CMD = cmd/notescan.go

.PHONY: build
build:
	$(GO) build -o $(NOTESCAN_BIN) $(NOTESCAN_CMD)
