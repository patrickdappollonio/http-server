.PHONY: build

default: build

build:
	go build -a -tags netgo -ldflags '-s -w' -o docker-http-server
