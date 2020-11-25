APPNAME = presence
VERSION = $(shell git describe --long --always --dirty 2>/dev/null || echo -n 'v0.1')
GO = go
GOMODULE = main
GOPATH ?= $(shell mktemp -d)
CWD = $(shell pwd)
PREFIX ?= /usr
BINDIR ?= ${PREFIX}/bin
BUILDDIR = .

all: build

build:
	${GO} build -v \
		-o ${BUILDDIR}/${APPNAME} \
		-ldflags "-X '${GOMODULE}.Version=${VERSION}'" \
		"${CWD}/src" \
		&& echo "-> ${BUILDDIR}/${APPNAME}" \
		|| echo "*** Build failed ***" 1>&2; \

install: ${APPNAME}
	@install -v -D -t ${DESTDIR}${BINDIR} ${APPNAME}

clean: ${APPNAME}
	@rm -rf ${APPNAME}

.PHONY: build