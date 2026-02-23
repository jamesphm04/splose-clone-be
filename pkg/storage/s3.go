// Package storage wraps AWS S3
package storage

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"go.uber.org/zap"
)

// Client wraps the AWS S3 Client with bucket-scroped operations
type Client struct {
	s3     *s3.Client
	bucket string
	log    *zap.Logger
}

// NewClient creates an S3 Client
// if endpoint is non-empty the client points at that URL (LocalStack, MinIO)
func NewClient(ctx context.Context, region, accessKey, secretKey, bucket, endpoint string, log *zap.Logger) (*Client, error) {
	opts := []func(*awsconfig.LoadOptions) error{
		awsconfig.WithRegion(region),
		awsconfig.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(accessKey, secretKey, ""),
		),
	}

	cfg, err := awsconfig.LoadDefaultConfig(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("loading aws config: %w", err)
	}

	s3Opts := []func(*s3.Options){}
	if endpoint != "" {
		s3Opts = append(s3Opts, func(o *s3.Options) {
			o.BaseEndpoint = aws.String(endpoint)
			o.UsePathStyle = true
		})
	}

	c := &Client{
		s3:     s3.NewFromConfig(cfg, s3Opts...),
		bucket: bucket,
		log:    log.Named("s3"),
	}

	log.Info("S3 client initialized",
		zap.String("region", region),
		zap.String("bucket", bucket),
		zap.Bool("customEndpoint", endpoint != ""),
	)
	return c, nil
}

// UploadInput carries the data required to store a file
type UploadInput struct {
	Key         string
	Body        io.Reader
	ContentType string
	Size        int64
}

type UploadOutput struct {
	URL string
}

// Upload stores a file in S3 and returns its URL
func (c *Client) Upload(ctx context.Context, in UploadInput) (*UploadOutput, error) {
	o := &s3.PutObjectInput{
		Bucket:        aws.String(c.bucket),
		Key:           aws.String(in.Key),
		Body:          in.Body,
		ContentType:   aws.String(in.ContentType),
		ContentLength: aws.Int64(in.Size),
	}

	_, err := c.s3.PutObject(ctx, o)
	if err != nil {
		c.log.Error("PutObject failed", zap.String("key", in.Key), zap.Error(err))
		return nil, fmt.Errorf("s3 PutObject %q: %w", in.Key, err)
	}

	url := fmt.Sprintf("https://%s.s3.amazonaws.com/%s", c.bucket, in.Key)
	c.log.Info("object uploaded", zap.String("key", in.Key), zap.Int64("size", in.Size))
	return &UploadOutput{URL: url}, nil
}

// Delete removes an object from S3
func (c *Client) Delete(ctx context.Context, key string) error {
	o := &s3.DeleteObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(key),
	}

	_, err := c.s3.DeleteObject(ctx, o)
	if err != nil {
		c.log.Error("DeleteObject failed", zap.String("key", key), zap.Error(err))
		return fmt.Errorf("s3 DeleteObject %q: %w", key, err)
	}

	c.log.Info("object deleted", zap.String("key", key))
	return nil
}

// PresignURL generates a time-limited pre-signed GET URL for private objects. Preview shortly -> Private
func (c *Client) PresignURL(ctx context.Context, key string, ttl time.Duration) (string, error) {
	presignClient := s3.NewPresignClient(c.s3)

	o := &s3.GetObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(key),
	}

	req, err := presignClient.PresignGetObject(ctx, o, s3.WithPresignExpires(ttl))
	if err != nil {
		c.log.Error("PresignGetObject failed", zap.String("key", key), zap.Error(err))
		return "", fmt.Errorf("presigning %q: %w", key, err)
	}
	return req.URL, nil
}
