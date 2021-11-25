MKFILE_PATH := $(dir $(abspath $(lastword $(MAKEFILE_LIST))))

BIN_DIR := $(MKFILE_PATH)bin

all : bin/raft

clean : 
	$(RM) -rf $(BIN_DIR)

bin/raft:
	mkdir -p $(BIN_DIR)
	GOOS=linux GOARCH=amd64 go build -o $(BIN_DIR)/amd64-linux/raft
	GOOS=darwin GOARCH=amd64 go build -o $(BIN_DIR)/amd64-darwin/raft
