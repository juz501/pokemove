BINARY=pokemove
MAIN=src/pokemove/pokemove.go
PACKAGES=github.com/chromatixau/negroni github.com/chromatixau/gomiddleware

all: build run

build: clean
	GOPATH=`pwd -P` go build -o bin/$(BINARY) $(MAIN)

clean:
	rm -f bin/$(BINARY)

run: build
	bin/$(BINARY)

install: 
	GOPATH=`pwd -P` go get $(PACKAGES)