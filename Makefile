.PHONY: generate

generate:
	protoc -I./testdata --gofast_out=paths=source_relative:./testdata/pb message.proto