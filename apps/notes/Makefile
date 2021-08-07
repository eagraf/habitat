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
	GOARCH=amd64 GOOS=linux go build -o $(OUT_DIR)/backend/notes-api github.com/eagraf/habitat/apps/notes/backend/cmd/api

.PHONY: out
out : build-frontend build-backend
