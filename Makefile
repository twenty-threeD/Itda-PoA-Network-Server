#!/usr/bin/make -f

APP_NAME := itdad
GOBIN ?= $(shell go env GOPATH)/bin

# buf resolves codegen plugins off PATH, so GOBIN has to be on it.
export PATH := $(GOBIN):$(PATH)

.PHONY: all build test proto-gen proto-lint tools clean

all: proto-gen build

## tools: install the protobuf codegen toolchain into GOBIN
tools:
	go install github.com/bufbuild/buf/cmd/buf@v1.50.0
	go install github.com/cosmos/gogoproto/protoc-gen-gocosmos@v1.7.2
	go install github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway@v1.16.0

## proto-gen: regenerate *.pb.go from proto/ into x/<module>/types/
# buf writes into the full go_package path, so the tree is flattened back
# into the repo root afterwards and the scratch directory removed.
proto-gen:
	cd proto && $(GOBIN)/buf generate --template buf.gen.gogo.yaml
	cp -r proto/github.com/twenty-threeD/Itda-PoA-Network-Server/* .
	rm -rf proto/github.com

## proto-lint: check proto style against the rules in proto/buf.yaml
proto-lint:
	cd proto && $(GOBIN)/buf lint

build:
	go build -o build/$(APP_NAME) ./cmd/$(APP_NAME)

test:
	go test ./...

clean:
	rm -rf build/
