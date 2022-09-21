TARGETS = bin/amd64-linux/habitat \
		bin/amd64-darwin/habitat \
		bin/amd64-linux/habitatctl \
		bin/amd64-darwin/habitatctl

include ../common.mk

install::
	mkdir -p $(BINDIR)
	$(CP) -r ./bin/$(CURARCH)/* $(BINDIR)

clean::
	rm -rf bin
