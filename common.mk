MKFILE_PATH := $(dir $(abspath $(lastword $(MAKEFILE_LIST))))
NPROCS = $(shell sysctl hw.ncpu  | grep -o '[0-9]\+')
MAKEFLAGS += -j$(NPROCS)

export BIN_DIR := $(MKFILE_PATH)bin
export DEV_DATA_DIR := $(MKFILE_PATH)data
export DEV_PROC_DIR := $(DEV_DATA_DIR)/procs
export DEV_COMMUNITY_DIR := $(DEV_DATA_DIR)/communities
export APPS_DIR := $(MKFILE_PATH)apps