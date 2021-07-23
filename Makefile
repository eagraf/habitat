include ./common.mk

all : build

build : clean
	mkdir -p $(BIN_DIR)
	go build -o $(BIN_DIR) github.com/eagraf/habitat/cmd/habitat github.com/eagraf/habitat/cmd/habitatctl

clean :
	rm -rf $(BIN_DIR)

run-dev : build
	$(BIN_DIR)/habitat --procdir $(DEV_PROC_DIR)

install-notes : 
	$(MAKE) -C apps/notes all
	mkdir -p $(DEV_PROC_DIR)/notes_backend/bin
	cp -r $(APPS_DIR)/notes/out/backend/* $(DEV_PROC_DIR)/notes_backend/bin
	cp -r $(APPS_DIR)/notes/out/frontend/* $(DEV_PROC_DIR)/nginx/content
