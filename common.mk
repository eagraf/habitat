MKFILE_PATH := $(dir $(abspath $(lastword $(MAKEFILE_LIST))))

export BIN_DIR := $(MKFILE_PATH)bin
export DEV_DATA_DIR := $(MKFILE_PATH)data
export DEV_PROC_DIR := $(DEV_DATA_DIR)/procs
export DEV_COMMUNITY_DIR := $(DEV_DATA_DIR)/communities
export APPS_DIR := $(MKFILE_PATH)apps