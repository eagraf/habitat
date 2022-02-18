include ./common.mk

install-community : 
	$(MAKE) -C apps/community all
	cp $(APPS_DIR)/community/out/backend/amd64-linux/community_backend $(DEV_PROC_DIR)/bin/amd64-linux/community_backend
	cp $(APPS_DIR)/community/out/backend/amd64-darwin/community_backend $(DEV_PROC_DIR)/bin/amd64-darwin/community_backend

	mkdir -p $(DEV_PROC_DIR)/web/community
	cp -R $(APPS_DIR)/community/out/frontend/* $(DEV_PROC_DIR)/web/community

all : build

build : clean install-community
	mkdir -p $(BIN_DIR) $(BIN_DIR)/amd64-linux
	go build -o $(BIN_DIR) github.com/eagraf/habitat/cmd/habitat github.com/eagraf/habitat/pkg/habitatctl
	GOOS=linux GOARCH=amd64 go build -o $(BIN_DIR)/amd64-linux github.com/eagraf/habitat/cmd/habitat github.com/eagraf/habitat/pkg/habitatctl

clean :
	rm -rf $(BIN_DIR)

habitat :
	echo "run habitat"
	HABITAT_PATH=$(DEV_DATA_DIR) $(BIN_DIR)/habitat --hostname localhost

c-frontend :
	echo "run community frontend"
	serve -s $(DEV_PROC_DIR)/web/community
	# npm --prefix $(APPS_DIR)/community/frontend start

c-backend :
	echo "run community backend"
	cd $(DEV_PROC_DIR)/bin/amd64-darwin && HABITAT_PATH=$(DEV_DATA_DIR) ./community_backend


run-communities : c-frontend c-backend

run : habitat c-backend c-frontend

install-setup :
	rm -rf $(DEV_PROC_DIR)/bin/*
	rm -rf $(DEV_PROC_DIR)/web/*
	mkdir -p $(DEV_PROC_DIR)/bin/amd64-linux $(DEV_PROC_DIR)/bin/amd64-darwin
	mkdir -p $(DEV_PROC_DIR)/web
	mkdir -p $(DEV_PROC_DIR)/data
	mkdir -p $(DEV_COMMUNITY_DIR)

install-ipfs :
	cp $(APPS_DIR)/ipfs/start-ipfs $(DEV_PROC_DIR)/bin/amd64-linux
	cp $(APPS_DIR)/ipfs/start-ipfs $(DEV_PROC_DIR)/bin/amd64-darwin
	

install-notes : 
	$(MAKE) -C apps/notes all
	cp $(APPS_DIR)/notes/out/backend/amd64-linux/notes_backend $(DEV_PROC_DIR)/bin/amd64-linux/notes_backend
	cp $(APPS_DIR)/notes/out/backend/amd64-darwin/notes_backend $(DEV_PROC_DIR)/bin/amd64-darwin/notes_backend

	mkdir -p $(DEV_PROC_DIR)/web/notes
	cp -R $(APPS_DIR)/notes/out/frontend/* $(DEV_PROC_DIR)/web/notes

install: install-setup install-notes install-ipfs install-community
