SYSTEM := ""

ifeq ($(OS),Windows_NT)
    SYSTEM = "WIN32"
else
    UNAME_S := $(shell uname -s)
    ifeq ($(UNAME_S),Linux)
        SYSTEM = "LINUX"
    endif
    ifeq ($(UNAME_S),Darwin)
        SYSTEM = "OSX"
    endif
endif