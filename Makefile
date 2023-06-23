.PHONY: generate test lint install-tools

generate:
	protoc -I./testdata --gofast_out=paths=source_relative:./testdata/pb message.proto

test:
	go test -v ./...
	go test -v ./testdata/generated/...

test-race:
	go test -v -race -count=1 ./...
	go test -v -race -count=1 ./testdata/generated/...

lint:
	go fmt ./...
	go vet ./...
	revive -config revive.toml -formatter friendly ./...

install-tools:
	go install github.com/mgechev/revive
