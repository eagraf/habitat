MKFILE_PATH := $(dir $(abspath $(lastword $(MAKEFILE_LIST))))

export BIN_DIR := $(MKFILE_PATH)bin