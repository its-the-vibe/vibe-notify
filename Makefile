BINARY   := vibe-notify
GOFLAGS  :=
BUILD_DIR := bin

.PHONY: all build test lint clean

all: build

## build: compile the binary into $(BUILD_DIR)/$(BINARY)
build:
	go build $(GOFLAGS) -o $(BUILD_DIR)/$(BINARY) .

## test: run all unit tests with race detection and coverage
test:
	go test -race -coverprofile=coverage.out ./...

## lint: run go vet
lint:
	go vet ./...

## clean: remove build artefacts
clean:
	rm -rf $(BUILD_DIR) coverage.out
