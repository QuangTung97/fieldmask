protoc -I./testdata \
  --gofast_out=\
Mgoogle/protobuf/timestamp.proto=github.com/gogo/protobuf/types,\
Mgoogle/protobuf/wrappers.proto=github.com/gogo/protobuf/types,\
paths=source_relative:./testdata/pb \
  message.proto
