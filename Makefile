run: build
	@./bin/goredis --listenAddr :5000

build: 
	@go build -o bin/goredis . 
