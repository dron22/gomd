SOURCEDIR=.
SOURCES := $(shell find $(SOURCEDIR) -name '*.go')

BINARY=${GOPATH}/bin/gomd

VERSION=1.0.0
BUILD_TIME=`date +%FT%T%Z`

LDFLAGS=-ldflags "-X github.com/dron22/gomd/core.Version=${VERSION} -X github.com/dron22/gomd/core.BuildTime=${BUILD_TIME}"

.DEFAULT_GOAL: $(BINARY)

$(BINARY): $(SOURCES)
	go build ${LDFLAGS} -o ${BINARY} gomd.go

.PHONY: install
install:
	go install ${LDFLAGS} ./...

.PHONY: clean
clean:
	if [ -f ${BINARY} ]; then rm ${BINARY}; fi
