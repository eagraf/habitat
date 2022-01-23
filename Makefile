include ./common.mk

all : build

build : clean
	mkdir -p $(BIN_DIR) $(BIN_DIR)/amd64-linux
	go build -o $(BIN_DIR) github.com/eagraf/habitat/cmd/habitat github.com/eagraf/habitat/cmd/habitatctl
	GOOS=linux GOARCH=amd64 go build -o $(BIN_DIR)/amd64-linux github.com/eagraf/habitat/cmd/habitat github.com/eagraf/habitat/cmd/habitatctl

clean :
	rm -rf $(BIN_DIR)

run-dev : build
	$(BIN_DIR)/habitat --procdir $(DEV_PROC_DIR) --communitydir $(DEV_COMMUNITY_DIR) --hostname localhost --root $(DEV_DATA_DIR)

install-setup :
	rm -rf $(DEV_PROC_DIR)/bin/*
	rm -rf $(DEV_PROC_DIR)/web/*
	mkdir -p $(DEV_PROC_DIR)/bin/amd64-linux $(DEV_PROC_DIR)/bin/amd64-darwin
	mkdir -p $(DEV_PROC_DIR)/web
	mkdir -p $(DEV_PROC_DIR)/data
	mkdir -p $(DEV_COMMUNITY_DIR)

install-ipfs :
	mkdir -p $(DEV_PROC_DIR)/bin/amd64-linux $(DEV_PROC_DIR)/bin/amd64-darwin
	mkdir -p $(DEV_PROC_DIR)/web
	mkdir -p $(DEV_PROC_DIR)/data
	cp $(APPS_DIR)/ipfs/ipfs $(DEV_PROC_DIR)/bin/amd64-linux
	cp $(APPS_DIR)/ipfs/ipfs $(DEV_PROC_DIR)/bin/amd64-darwin

install-notes : 
	$(MAKE) -C apps/notes all
	cp $(APPS_DIR)/notes/out/backend/amd64-linux/notes_backend $(DEV_PROC_DIR)/bin/amd64-linux/notes_backend
	cp $(APPS_DIR)/notes/out/backend/amd64-darwin/notes_backend $(DEV_PROC_DIR)/bin/amd64-darwin/notes_backend

	mkdir -p $(DEV_PROC_DIR)/web/notes
	cp -R $(APPS_DIR)/notes/out/frontend/* $(DEV_PROC_DIR)/web/notes

install: install-setup install-notes install-ipfs
