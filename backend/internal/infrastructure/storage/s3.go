package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type S3 struct {
	bucket    string
	client    *s3.Client
	presigner *s3.PresignClient
}

func NewS3(ctx context.Context, endpoint, region, bucket, accessKey, secretKey string, pathStyle bool) (*S3, error) {
	cfg, err := awsconfig.LoadDefaultConfig(ctx, awsconfig.WithRegion(region),
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKey, secretKey, "")))
	if err != nil {
		return nil, fmt.Errorf("load S3 configuration: %w", err)
	}
	client := s3.NewFromConfig(cfg, func(options *s3.Options) {
		options.BaseEndpoint = aws.String(endpoint)
		options.UsePathStyle = pathStyle
	})
	return &S3{bucket: bucket, client: client, presigner: s3.NewPresignClient(client)}, nil
}

func (s *S3) VerifyObject(ctx context.Context, objectKey, mimeType string, size int64) error {
	result, err := s.client.HeadObject(ctx, &s3.HeadObjectInput{Bucket: aws.String(s.bucket), Key: aws.String(objectKey)})
	if err != nil {
		return fmt.Errorf("inspect uploaded object: %w", err)
	}
	if result.ContentLength == nil || *result.ContentLength != size {
		return fmt.Errorf("uploaded object size does not match requested size")
	}
	if result.ContentType == nil || *result.ContentType != mimeType {
		return fmt.Errorf("uploaded object content type does not match requested MIME type")
	}
	return nil
}

func (s *S3) PresignGet(ctx context.Context, objectKey string, ttl time.Duration) (string, error) {
	result, err := s.presigner.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket), Key: aws.String(objectKey),
	}, s3.WithPresignExpires(ttl))
	if err != nil {
		return "", fmt.Errorf("presign S3 download: %w", err)
	}
	return result.URL, nil
}

func (s *S3) PresignPut(ctx context.Context, objectKey, mimeType string, size int64, ttl time.Duration) (string, map[string]string, error) {
	result, err := s.presigner.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(s.bucket), Key: aws.String(objectKey), ContentType: aws.String(mimeType), ContentLength: aws.Int64(size),
	}, s3.WithPresignExpires(ttl))
	if err != nil {
		return "", nil, fmt.Errorf("presign S3 upload: %w", err)
	}
	headers := map[string]string{"Content-Type": mimeType, "Content-Length": fmt.Sprintf("%d", size)}
	return result.URL, headers, nil
}
