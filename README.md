# compressrv - compress service

## Build protobuff
```bash
protoc -I pkg/proto --go_out=pkg/proto --go_opt=paths=source_relative --go-grpc_out=require_unimplemented_servers=false:pkg/proto --go-grpc_opt=paths=source_relative pkg/proto/compressor.proto
```
## ENV Variables
- ROOT - absolute path the project
- FFMPEG_PATH - absolute path to ffmpeg
- FFPROBE_PATH - absolute path to ffprobe