.PHONY: generate test

generate:
	protoc -I./testdata --gofast_out=paths=source_relative:./testdata/pb message.proto

test:
	go test -v ./...
	go test -v ./testdata/generated/...