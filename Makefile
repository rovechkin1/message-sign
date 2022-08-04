
build-service:
	go build -o bin/service service/cmd/main.go

build-key-gen:
	go build -o bin/key-generator key_generator/key_generator.go

build-record-gen:
	go build -o bin/record-generator mongo/record_generator.go