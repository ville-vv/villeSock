#villeSock makefile
BINARY=./bin/villeSock
VERSION=v1.2.0
BUILD=`data +%FT%T%Z`

LDFLAGS=-ldflags "-X main.Version=${VERSION}"

build:
	go build ${LDFLAGS} -o ${BINARY}

install:
	go install ${LDFLAGS}

clean:
	if [ -f ${BINARY} ] ; then rm ${BINARY} ; fi

.PHONY: clean install