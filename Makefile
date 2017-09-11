IMAGE_TAG ?= patrickdappollonio/docker-http-server
BIN_NAME = http-server

default: release

build:
	go build -a -tags netgo -ldflags '-s -w' -o $$(pwd)/$(BIN_NAME)

generate:
	go generate

remove-gen:
	rm -rf $$(pwd)/*_gen.go

clean:
	rm -rf $$(pwd)/$(BIN_NAME)

docker:
	docker build --pull=true --rm=true -t $(IMAGE_TAG) .

release:
	@$(MAKE) generate
	@$(MAKE) build
	@$(MAKE) remove-gen

ci:
	@$(MAKE) generate
	@$(MAKE) build
	@$(MAKE) remove-gen
	@$(MAKE) docker
	@$(MAKE) clean

.NOTPARALLEL:
