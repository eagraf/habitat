MIN_MAKE_VERSION	=	4.0.0

ifneq ($(MIN_MAKE_VERSION),$(firstword $(sort $(MAKE_VERSION) $(MIN_MAKE_VERSION))))
$(error you must have a version of GNU make newer than v$(MIN_MAKE_VERSION) installed)
endif

export MAKECMDGOALS

SHELL			:=	bash
.SHELLFLAGS		:=	-eu -o pipefail -c

MAKEFLAGS		+=	--warn-undefined-variables	\
				--no-builtin-rules 		\

CP			=	cp

OS			=	$(shell uname)

ARCH			=	unknown
CURARCH			=	unknown

ifeq ($(OS), Linux)
ARCH			=	amd64 # sry arm
CURARCH			=	amd64-linux
endif

ifeq ($(OS), Darwin)
ARCH			=	amd64
CURARCH			=	amd64-darwin
endif

# If TOPDIR isn't already defined, let's go with a default
ifeq ($(origin TOPDIR), undefined)
TOPDIR			:=	$(realpath $(patsubst %/,%, $(dir $(lastword $(MAKEFILE_LIST)))))
endif

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

GO_BUILD_CMD		=	$(GO_BUILD) -o $@ ./$<
GO_DEFAULT_DEPS		=	.FORCE
GO_BUILD_PACKAGE	=	./$(@F)

GO_RUN_FLAGS			?=
GO_TEST_FLAGS			?=
GO_BUILD_FLAGS_NATIVE	?=

GO_BUILD_AMD64_LINUX		= GOARCH=amd64 GOOS=linux $(GO_BUILD)
GO_BUILD_AMD64_DARWIN		= GOARCH=amd64 GOOS=darwin $(GO_BUILD)
GO_BUILD_AMD64_WINDOWS		= GOARCH=amd64 GOOS=windows$(GO_BUILD)

TARGETS			?=
SUBDIRS			?=

HABITAT_ROOT	=	$(TOPDIR)
PREFIX			=	$(HABITAT_ROOT)/dist
BINDIR			=	$(PREFIX)/bin
APPDIR			=	$(PREFIX)/apps

DEV_HABITAT_PATH 		= $(TOPDIR)/.habitat
DEV_APPDIR				= $(DEV_HABITAT_PATH)/apps
DEV_HABITAT_APP_PATH 	= $(DEV_APPDIR):${HABITAT_APP_PATH}

all:: $(TARGETS)

clean::
	$(RM) -r $(TARGETS)

all install clean test lint::
	@set -e; for dir in $(SUBDIRS); do $(MAKE) -C $$dir $@; done

$(SUBDIRS):
	$(MAKE) -C $@

.PHONY: all install clean distclean test lint $(SUBDIRS)

.FORCE:

.ONESHELL:
