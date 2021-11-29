# syntax=docker/dockerfile:1

FROM golang:1.16

COPY . /go/src/app

WORKDIR /go/src/app

ENV ROOT "/go/src/app"
ENV FFMPEG_PATH "/usr/bin/ffmpeg"
ENV FFPROBE_PATH "/usr/bin/ffprobe"
ENV RABBIT_USER guest
ENV RABBIT_PASSWORD guest
ENV RABBIT_HOST localhost
ENV RABBIT_PORT 5672
ENV AWS_BUCKET_NAME name
ENV AWS_ACCESS_KEY access_key
ENV AWS_SECRET_KEY secret_key
ENV AWS_REGION region

RUN apt update
RUN apt install -y ffmpeg

RUN go build -o compressrv cmd/compressrv/main.go

CMD ["./compressrv"]