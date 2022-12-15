TARGETS = bin/amd64-linux/community \
		bin/amd64-darwin/community \
		web

include ../../common.mk

all:: $(TARGETS)

clean::
	rm -rf web
	rm -rf bin

bin/amd64-linux/%:
	$(GO_BUILD_AMD64_LINUX) -o $@ ./backend/cmd/api

bin/amd64-darwin/%:
	$(GO_BUILD_AMD64_DARWIN) -o $@ ./backend/cmd/api

bin/amd64-windows/%:
	$(GO_BUILD_AMD64_WINDOWS) -o $@ ./backend/cmd/api

web:
	mkdir -p ./web
	(cd frontend && npm install && npm run build)
	$(CP) -r frontend/build/* ./web
