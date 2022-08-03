
build:
	go build -o bin/service service/cmd/main.go

build-key-generator:
	go build -o bin/key-generator key_generator/key_generator.go
