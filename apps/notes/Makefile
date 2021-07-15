MKFILE_PATH := $(dir $(abspath $(lastword $(MAKEFILE_LIST))))

export OUT_DIR := $(MKFILE_PATH)out

all : build-frontend build-backend

clean :
	rm -rf $(OUT_DIR)

build-frontend : clean
	rm -rf $(MKFILE_PATH)frontend/build/*
	mkdir -p $(OUT_DIR)/frontend
	$(MAKE) -C frontend build-frontend
	cp -r $(MKFILE_PATH)frontend/build/* $(OUT_DIR)/frontend

build-backend : clean
	mkdir -p $(OUT_DIR)/backend
	go build -o $(OUT_DIR)/backend/notes-api github.com/eagraf/habitat/apps/notes/backend/cmd/api

