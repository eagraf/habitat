TARGETS = bin/amd64-linux/ipfs-driver \
		bin/amd64-darwin/ipfs-driver

include ../../common.mk

all:: $(TARGETS)

clean::
	rm -rf bin

bin/amd64-linux/%:
	$(GO_BUILD_AMD64_LINUX) -o $@ .

bin/amd64-darwin/%:
	$(GO_BUILD_AMD64_DARWIN) -o $@ .

bin/amd64-windows/%:
	$(GO_BUILD_AMD64_WINDOWS) -o $@ .
