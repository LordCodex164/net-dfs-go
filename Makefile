build:
	@go build -o tmp/main .

run: build
	@./tmp/main

test:
	@go test ./... -v
