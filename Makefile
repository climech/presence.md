APPNAME = presence
VERSION = $(shell git describe --long --always --dirty 2>/dev/null || echo -n 'v0.1')
GO = go
GOMODULE = presence
GOPATH ?= $(shell mktemp -d)
PREFIX ?= /usr
BINDIR ?= ${PREFIX}/bin
CWD = $(shell pwd)

all: build

build:
	@env -C "${CWD}/src" ${GO} build \
		-v \
		-o "${CWD}/${APPNAME}" \
		-ldflags "-X '${GOMODULE}.app.Version=${VERSION}'" \
		"${CWD}/src" \
		&& echo "-> ${APPNAME}" \
		|| echo "*** Build failed ***" 1>&2;

test:
	@env -C "${CWD}/src" ${GO} test -count=1 \
		./config \
		./store

install: ${APPNAME}
	@install -v -D -t "${DESTDIR}${BINDIR}" ${APPNAME}

clean: ${APPNAME}
	@rm -rf ${APPNAME}

.PHONY: build test