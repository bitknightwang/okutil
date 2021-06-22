.PHONY: all help init build run test clean dist

GOCMD     = go

GOBUILD   = $(GOCMD) build
GOCLEAN   = $(GOCMD) clean
GOTEST    = $(GOCMD) test
GOGET     = $(GOCMD) get
GORUN     = $(GOCMD) run
GOMOD     = $(GOCMD) mod

BUILDDIR  = ./build
BINDIR    = $(BUILDDIR)/bin
PKGDIR    = $(BUILDDIR)/pkg

NAME      = okutil
VERSION   = 0.0.1
GOFILES  := $(shell find . -type f -name '*.go')
GOXOS     = linux
GOXARCH   = amd64
OUTBIN    = $(BINDIR)/$(NAME)
DISTBIN   = $(PKGDIR)/$(GOXOS)_$(GOXARCH)/$(NAME)


all: help

help:
	@echo "Build and distribute util package"
	@echo "    init                       go mod init dependencies"
	@echo "    clean                      clean build output"
	@echo "    build                      compile binary"
	@echo "    test                       run test"
	@echo "    dist                       compile and generate $(GOXOS)_$(GOXARCH) binary"

init:
	@echo "initialize dependencies"
	mkdir -p $(BINDIR)
	mkdir -p $(PKGDIR)
	$(GOMOD) tidy

clean:
	$(GOCLEAN)
	rm -rf $(BUILDDIR)

build: init
	@echo build binary
	rm -rf $(BINDIR)/*
	$(GOBUILD) -o $(OUTBIN) $(GOFILES)

test: build
	@echo build and run test

dist: init
	@echo build $(GOXOS)_$(GOXARCH) binary
	rm -rf $(PKGDIR)/*
	GOOS=$(GOXOS) GOARCH=$(GOXARCH) $(GOBUILD) -o $(DISTBIN) $(GOFILES)

