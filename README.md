# compressrv - compress service

## Build protobuff
```bash
protoc -I pkg/proto --go_out=pkg/proto --go_opt=paths=source_relative --go-grpc_out=require_unimplemented_servers=false:pkg/proto --go-grpc_opt=paths=source_relative pkg/proto/compressor.proto
```
