TARGETS = bin/amd64-linux/ipfs-driver \
		bin/amd64-darwin/ipfs-driver \
		bin/amd64-linux/schema-registry \
		bin/amd64-linux/schema-registry

include ../common.mk

IPFS_DRIVER_DIR = $(DEV_APPDIR)/ipfs-driver
SCHEMA_REGISTRY_DIR = $(DEV_APPDIR)/schema-registry


install:: install-ipfs-driver install-schema-registry

clean::
	rm -rf bin

install-ipfs-driver:
	mkdir -p $(IPFS_DRIVER_DIR)/bin	
	$(CP) ./bin/$(CURARCH)/ipfs-driver $(IPFS_DRIVER_DIR)/bin
	$(CP) ./ipfs-driver/habitat.yaml $(IPFS_DRIVER_DIR)


install-schema-registry:
	mkdir -p $(SCHEMA_REGISTRY_DIR)/bin	
	$(CP) ./bin/$(CURARCH)/schema-registry $(SCHEMA_REGISTRY_DIR)/bin
	$(CP) ./schema-registry/habitat.yaml $(SCHEMA_REGISTRY_DIR)
