all: build

build: 
	@go build -o telnest

run: 
	@go run .

fmt: 
	@go fmt ./...

clean: 
	@rm -f telnest
