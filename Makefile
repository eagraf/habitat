SUBDIRS		=	cmd apps

include ./common.mk

clean::
	$(RM) -r $(PREFIX)

install::
	mkdir -p $(DEV_HABITAT_PATH)
	mkdir -p $(DEV_HABITAT_PATH)/communities
	mkdir -p $(DEV_HABITAT_PATH)/procs
	cp ./apps.yml $(DEV_HABITAT_PATH)/procs

test::
	go test ./...

test::
	prove -v -r

lint::
	CGO_ENABLED=0 golangci-lint run --skip-dirs '(^|/)virtctl($$|/)' -D errcheck ./...

include ./devtargets.mk
