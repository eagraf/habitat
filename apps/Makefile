TARGETS = bin/amd64-linux/ipfs-driver \
		bin/amd64-darwin/ipfs-driver

include ../common.mk

IPFS_DRIVER_DIR = $(DEV_APPDIR)/ipfs-driver

install:: install-ipfs-driver

clean::
	rm -rf bin

install-ipfs-driver:
	mkdir -p $(IPFS_DRIVER_DIR)/bin	
	$(CP) ./bin/$(CURARCH)/ipfs-driver $(IPFS_DRIVER_DIR)/bin
	$(CP) ./ipfs-driver/habitat.yaml $(IPFS_DRIVER_DIR)
