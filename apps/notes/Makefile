MKFILE_PATH := $(dir $(abspath $(lastword $(MAKEFILE_LIST))))

export OUT_DIR := $(MKFILE_PATH)out

all : clean build-frontend build-backend

clean :
	rm -rf $(OUT_DIR)

build-frontend : 
	rm -rf $(MKFILE_PATH)frontend/build/*
	mkdir -p $(OUT_DIR)/frontend
	$(MAKE) -C frontend build-frontend
	cp -R $(MKFILE_PATH)frontend/build/* $(OUT_DIR)/frontend

build-backend : 
	mkdir -p $(OUT_DIR)/backend
	GOOS=linux GOARCH=amd64 go build -o $(OUT_DIR)/backend/amd64-linux/notes_backend github.com/eagraf/habitat/apps/notes/backend/cmd/api
	GOOS=darwin GOARCH=amd64 go build -o $(OUT_DIR)/backend/amd64-darwin/notes_backend github.com/eagraf/habitat/apps/notes/backend/cmd/api

.PHONY: out
out : build-frontend build-backend
