PIP=pip3
PYTHON=python3

ibrew:
	/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/master/install.sh)"

iprotoc:
	brew install protobuf

iprotoc-gen-go:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

ipyproto:
	$(PIP) install grpcio
	$(PIP) install grpcio-tools

.PHONY: all protos run

# Regenerate *.pb.go files for both client and server when vehicle.proto changes
protos:
	$(MAKE) -C client protos
	$(MAKE) -C server protos

# Run the simulation (assuming the main function is under client/client.go)
run:
	go run ./client

# Default target: regenerate proto files only
all: protos
