MKFILE_PATH := $(dir $(abspath $(lastword $(MAKEFILE_LIST))))

export OUT_DIR := $(MKFILE_PATH)out

all : build-backend

clean :
	rm -rf $(OUT_DIR)

build-backend :
	mkdir -p $(OUT_DIR)/backend
	go build -o $(OUT_DIR)/backend/notes-api github.com/eagraf/habitat/apps/notes/backend/cmd/api

