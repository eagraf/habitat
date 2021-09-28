MKFILE_PATH := $(dir $(abspath $(lastword $(MAKEFILE_LIST))))

export OUT_DIR := $(MKFILE_PATH)out

all : clean build-backend

clean :
	rm -rf $(OUT_DIR)

build-frontend : 
	rm -rf $(MKFILE_PATH)frontend/build/*
	mkdir -p $(OUT_DIR)/frontend
	$(MAKE) -C frontend build-frontend
	cp -R $(MKFILE_PATH)frontend/build/* $(OUT_DIR)/frontend

build-backend : 
	mkdir -p $(OUT_DIR)/backend
	GOOS=linux GOARCH=amd64 go build -o $(OUT_DIR)/backend/amd64-linux/communities_backend github.com/eagraf/habitat/apps/communities/backend/cmd/api
	GOOS=darwin GOARCH=amd64 go build -o $(OUT_DIR)/backend/amd64-darwin/communities_backend github.com/eagraf/habitat/apps/notes/communities/cmd/api

.PHONY: out
out : build-backend
