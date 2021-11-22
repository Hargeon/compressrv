package storage

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

const originalFileFolder = "/tmp/original_video/"

type AWSS3 struct {
	logger     *zap.Logger
	bucketName string
	accessKey  string
	secretKey  string
	region     string
}

// NewAWSS3 initialize AWSS3
func NewAWSS3(logger *zap.Logger, bucketName, region, accessKey, secretKey string) *AWSS3 {
	return &AWSS3{
		logger:     logger,
		bucketName: bucketName,
		accessKey:  accessKey,
		secretKey:  secretKey,
		region:     region,
	}
}

func (s *AWSS3) Download(ctx context.Context, id string) (string, error) {
	fileName := fmt.Sprintf("%s%s%s", os.Getenv("ROOT"), originalFileFolder, id)
	file, err := os.Create(fileName)

	if err != nil {
		return "", err
	}

	sess, err := s.session()

	if err != nil {
		return "", err
	}

	downloader := s3manager.NewDownloader(sess)
	_, err = downloader.DownloadWithContext(ctx, file,
		&s3.GetObjectInput{
			Bucket: aws.String(s.bucketName),
			Key:    aws.String(id),
		})

	if err != nil {
		return "", err
	}

	return fileName, nil
}

// Upload file to aws s3
func (s *AWSS3) Upload(ctx context.Context, fileName string, file io.Reader) (string, error) {
	sess, err := s.session()

	if err != nil {
		return "nil", err
	}

	uploader := s3manager.NewUploader(sess)

	newFileName := fmt.Sprintf("converted_%s_%s", uuid.New().String(), fileName)
	_, err = uploader.UploadWithContext(ctx, &s3manager.UploadInput{
		Body:   file,
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(newFileName),
	})

	if err != nil {
		return "", err
	}

	return newFileName, nil
}

func (s *AWSS3) session() (*session.Session, error) {
	sess, err := session.NewSession(
		&aws.Config{
			Region: aws.String(s.region),
			Credentials: credentials.NewStaticCredentials(
				s.accessKey,
				s.secretKey,
				""), // a token will be created when the session it's used.
		})

	return sess, err
}
