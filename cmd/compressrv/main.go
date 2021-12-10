package main

import (
	"context"
	"encoding/json"
	"log"
	"os"

	"github.com/Hargeon/compressrv/pkg/handler"
	"github.com/Hargeon/compressrv/pkg/service"
	"github.com/Hargeon/compressrv/pkg/service/broker"
	"github.com/Hargeon/compressrv/pkg/service/compressor"
	"github.com/Hargeon/compressrv/pkg/service/storage"

	"github.com/joho/godotenv"
	"github.com/streadway/amqp"
	"go.uber.org/zap"
)

var logger *zap.Logger

func main() {
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalln(err)
	}
	defer logger.Sync()

	err = godotenv.Load()
	if err != nil {
		logger.Fatal("godotenv Load", zap.String("Error", err.Error()))
	}

	// init Rabbit publisher
	publisher := broker.NewRabbit(os.Getenv("RABBIT_URL"))

	publisherConn, err := publisher.Connect("video_update_test")
	if err != nil {
		logger.Fatal("rabbit connect publisher", zap.String("Error", err.Error()))
	}
	defer publisherConn.Close()

	// init Rabbit consumer
	consumer := broker.NewRabbit(os.Getenv("RABBIT_URL"))

	consumerConn, err := consumer.Connect("video_convert_test")
	if err != nil {
		logger.Fatal("rabbit connect consumer", zap.String("Error", err.Error()))
	}
	defer consumerConn.Close()

	msgs, err := consumer.Consume()
	if err != nil {
		logger.Fatal("rabbit connect consumer", zap.String("Error", err.Error()))
	}

	st := storage.NewAWSS3(logger, os.Getenv("AWS_BUCKET_NAME"),
		os.Getenv("AWS_REGION"), os.Getenv("AWS_ACCESS_KEY"),
		os.Getenv("AWS_SECRET_KEY"))

	srv := service.NewService(st, os.Getenv("FFMPEG_PATH"), os.Getenv("FFPROBE_PATH"))
	h := handler.NewHandler(srv, logger)
	forever := make(chan bool)

	go func() {
		for d := range msgs {
			logger.Info("Received", zap.String("Message", string(d.Body)))

			req := new(compressor.Request)

			err = json.Unmarshal(d.Body, req)
			if err != nil {
				logger.Error("json Unmarshal", zap.String("Error", err.Error()))
				finishMsg(d)

				continue
			}

			resp := h.Compress(context.Background(), req)

			body, err := json.Marshal(resp)
			if err != nil {
				logger.Error("marshal", zap.String("Error", err.Error()))
				finishMsg(d)

				continue
			}

			err = publisher.Publish(body)
			if err != nil {
				logger.Error("publish response", zap.String("Error", err.Error()))
			}

			logger.Info("Worker finish", zap.String("Body", string(body)))

			finishMsg(d)
		}
	}()

	logger.Info(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever
}

func finishMsg(d amqp.Delivery) {
	if err := d.Ack(false); err != nil { // needs to mark a message was processed
		logger.Error("Ack", zap.String("Error", err.Error()))
	}
}
