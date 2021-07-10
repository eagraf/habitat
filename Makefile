include ./common.mk

all : build

build : clean
	mkdir -p $(BIN_DIR)
	go build -o $(BIN_DIR) github.com/eagraf/habitat/cmd/habitat github.com/eagraf/habitat/cmd/habitatctl

clean :
	rm -rf $(BIN_DIR)

run-dev : build
	$(BIN_DIR)/habitat
