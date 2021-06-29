.PHONY: all help init build run test clean dist

# VERSION  := $(shell git describe --tags)
VERSION   = 0.2.1
REVISION := $(shell git rev-parse --short HEAD)
NAME     := $(shell basename "$(PWD)")
SRCS     := $(shell find . -type f -name '*.go')

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
LDFLAGS  := -ldflags="-s -w -X \"main.Version=$(VERSION)\" -X \"main.Revision=$(REVISION)\" -extldflags \"-static\""

GOXOS     = linux
GOXARCH   = amd64
OUTBIN    = $(BINDIR)/$(NAME)-$(VERSION)-bin
DISTBIN   = $(PKGDIR)/$(GOXOS)_$(GOXARCH)/$(NAME)_$(GOXOS)_$(GOXARCH)-$(VERSION)-bin

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
	$(GOBUILD) $(LDFLAGS) -o $(OUTBIN) $(SRCS)

test: build
	@echo build and run test

dist: init
	@echo build $(GOXOS)_$(GOXARCH) binary
	rm -rf $(PKGDIR)/*
	GOOS=$(GOXOS) GOARCH=$(GOXARCH) $(GOBUILD) $(LDFLAGS) -o $(DISTBIN) $(SRCS)

