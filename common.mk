MKFILE_PATH := $(dir $(abspath $(lastword $(MAKEFILE_LIST))))

export BIN_DIR := $(MKFILE_PATH)bin
export DEV_PROC_DIR := $(MKFILE_PATH)procs
export APPS_DIR := $(MKFILE_PATH)apps