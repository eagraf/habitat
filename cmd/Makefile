TARGETS = bin/amd64-linux/habitat \
		bin/amd64-darwin/habitat \
		bin/amd64-linux/habitatctl \
		bin/amd64-darwin/habitatctl

include ../common.mk

all:: $(TARGETS)

install::
	mkdir -p $(BINDIR)/amd64-linux
	$(CP) -r ./bin/$(CURARCH)/* $(BINDIR)
	$(CP) -r ./bin/amd64-linux/* $(BINDIR)/amd64-linux

clean::
	rm -rf bin

build: clean $(TARGETS) install

bin/amd64-linux/%:
	$(GO_BUILD_AMD64_LINUX) -o $@ $(GO_BUILD_PACKAGE)

bin/amd64-darwin/%:
	$(GO_BUILD_AMD64_DARWIN) -o $@ $(GO_BUILD_PACKAGE)

bin/amd64-windows/%:
	$(GO_BUILD_AMD64_WINDOWS) -o $@ $(GO_BUILD_PACKAGE)

.PHONY: build
