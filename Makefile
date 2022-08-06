
build-service:
	go build -o bin/service service/cmd/main.go

build-key-gen:
	go build -o bin/key-generator key-generator/key_generator.go

build-record-gen:
	go build -o bin/record-generator record-generator/record_generator.go