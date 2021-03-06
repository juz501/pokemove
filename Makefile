BINARY=pokemove
MAIN=pokemove.go
PACKAGES=github.com/urfave/negroni github.com/juz501/go_logger_middleware

all: build run

background:
	bin/$(BINARY) &

build: clean
	GOPATH=`pwd -P` go build -o bin/$(BINARY) $(MAIN)

clean:
	rm -f bin/$(BINARY)

run: build
	bin/$(BINARY)

install: 
	GOPATH=`pwd -P` go get $(PACKAGES)
