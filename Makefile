IMAGE_TAG ?= patrickdappollonio/docker-http-server
GOOS ?= linux
BIN_NAME = http-server

default: release

build:
	GOOS=$(GOOS) go build -a -tags netgo -ldflags '-s -w' -o $$(pwd)/$(BIN_NAME)

generate:
	go generate

remove-gen:
	rm -rf $$(pwd)/*_gen.go

clean:
	rm -rf $$(pwd)/$(BIN_NAME)

docker:
	docker build --pull=true --rm=true -t $(IMAGE_TAG) .

release: generate build remove-gen

ci: generate build remove-gen docker clean

.NOTPARALLEL:
