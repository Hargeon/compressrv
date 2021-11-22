# compressrv - compress service

## Testing

```bash
go test -v -count=1 ./...
```

## ENV Variables
- ROOT - absolute path the project
- FFMPEG_PATH - absolute path to ffmpeg
- FFPROBE_PATH - absolute path to ffprobe
- RABBIT_USER
- RABBIT_PASSWORD
- RABBIT_HOST
- RABBIT_PORT
- AWS_BUCKET_NAME
- AWS_ACCESS_KEY
- AWS_SECRET_KEY
- AWS_REGION