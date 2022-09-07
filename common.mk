MKFILE_PATH := $(dir $(abspath $(lastword $(MAKEFILE_LIST))))

# TODO reenable this without system specific tools
# NPROCS = $(shell sysctl hw.ncpu  | grep -o '[0-9]\+')
NPROCS = 8
MAKEFLAGS += -j$(NPROCS)

export BIN_DIR := $(MKFILE_PATH)bin
export DEV_DATA_DIR := $(MKFILE_PATH)data
export DEV_PROC_DIR := $(DEV_DATA_DIR)/procs
export DEV_COMMUNITY_DIR := $(DEV_DATA_DIR)/communities
export APPS_DIR := $(MKFILE_PATH)apps

GO					=	go
GOFMT				=	gofmt
GOPATH				:=	$(shell $(GO) env GOPATH)
GO_ENV				?=
GO_ENV_NATIVE		?=
GO_ENV_CROSS		?=
GO_BUILD_FLAGS		?=
GO_BUILD			=	$(GO_ENV) $(GO) build $(GO_BUILD_FLAGS)
GO_TEST				=	$(GO_ENV) $(GO) test $(GO_TEST_FLAGS)
GO_TARGETS			=	$(shell $(GO) list ./... | grep -v /vendor/)
GO_RUN				=	$(GO_ENV) $(GO) run $(GO_RUN_FLAGS)

ALL_PLATFORMS		=	amd64-linux amd64-windows amd64-darwin
bin_all_platforms	=	$(foreach platform, $(ALL_PLATFORMS), bin/$(platform)/$(strip $(1)))

GO_BUILD_CMD		=	$(GO_BUILD) -o $@ ./$<
GO_DEFAULT_DEPS		=	.FORCE
GO_BUILD_PACKAGE	=	./$(@F)

GO_RUN_FLAGS			?=
GO_TEST_FLAGS			?=
GO_BUILD_FLAGS_NATIVE	?=

TARGETS			?=
SUBDIRS			?=

all:: $(TARGETS)

clean::
	$(RM) -r $(TARGETS)

all install clean test lint::
	@set -e; for dir in $(SUBDIRS); do $(MAKE) -C $$dir $@; done

$(SUBDIRS):
	$(MAKE) -C $@

bin/amd64-linux/%:
	GOARCH=amd64 GOOS=linux $(GO_BUILD) -o $@ $(GO_BUILD_PACKAGE)

bin/amd64-darwin/%:
	GOARCH=amd64 GOOS=darwin $(GO_BUILD) -o $@ $(GO_BUILD_PACKAGE)

bin/amd64-windows/%:
	GOARCH=amd64 GOOS=windows $(GO_BUILD) -o $@ $(GO_BUILD_PACKAGE)
