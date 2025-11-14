run: build
	@./bin/memstore -node-id 1

build:
	@go build -o bin/memstore ./cmd/main.go
