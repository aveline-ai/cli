.PHONY: build test vet lint tidy clean install

BIN := aveline
PKG := ./...
OUT := bin/$(BIN)

build:
	@mkdir -p bin
	go build -o $(OUT) ./cmd/aveline

test:
	go test $(PKG)

vet:
	go vet $(PKG)

lint: vet

tidy:
	go mod tidy

install:
	go install ./cmd/aveline

clean:
	rm -rf bin
