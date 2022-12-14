TARGETS = bin/amd64-linux/habitat \
		bin/amd64-darwin/habitat \
		bin/amd64-linux/habitatctl \
		bin/amd64-darwin/habitatctl

include ../common.mk

all:: $(TARGETS)

install::
	mkdir -p $(BINDIR)
	$(CP) -r ./bin/$(CURARCH)/* $(BINDIR)

clean::
	rm -rf bin

build: clean $(TARGETS) install

.PHONY: build
