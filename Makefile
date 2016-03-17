all: build

build:
	go get ./...
	go build .

install: build
	cp -f gosass /usr/local/bin/gosass

deps:
	go get -u gopkg.in/fsnotify.v1
	go get -u github.com/dullgiulio/pingo

unittests:
	mkdir -p integration/out
	go test github.com/dailymuse/gosass/compiler
