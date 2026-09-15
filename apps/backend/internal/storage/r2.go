package storage

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/smithy-go"
	"github.com/omar-shahieen/moneyflow/internal/config"
	"github.com/omar-shahieen/moneyflow/internal/ports"
	"github.com/rs/zerolog"
)

type R2Storage struct {
	client     *s3.Client
	presign    *s3.PresignClient
	bucketName string
	logger     *zerolog.Logger
}

func NewR2Storage(cfg config.R2Config, logger *zerolog.Logger) (*R2Storage, error) {
	endpoint := fmt.Sprintf("https://%s.r2.cloudflarestorage.com", cfg.AccountID)

	awsCfg, err := awsconfig.LoadDefaultConfig(context.TODO(),
		awsconfig.WithRegion("auto"),
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			cfg.AccessKeyID,
			cfg.SecretAccessKey,
			"",
		)),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config for R2: %w", err)
	}

	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(endpoint)
		o.UsePathStyle = true
	})

	presignClient := s3.NewPresignClient(client)

	return &R2Storage{
		client:     client,
		presign:    presignClient,
		bucketName: cfg.BucketName,
		logger:     logger,
	}, nil
}

func (s *R2Storage) GenerateUploadURL(key string, contentType string, expiry time.Duration) (string, error) {
	input := &s3.PutObjectInput{
		Bucket:      aws.String(s.bucketName),
		Key:         aws.String(key),
		ContentType: aws.String(contentType),
	}

	result, err := s.presign.PresignPutObject(context.TODO(), input, func(o *s3.PresignOptions) {
		o.Expires = expiry
	})
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned upload URL: %w", err)
	}

	return result.URL, nil
}

func (s *R2Storage) GenerateDownloadURL(key string, expiry time.Duration) (string, error) {
	input := &s3.GetObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(key),
	}

	result, err := s.presign.PresignGetObject(context.TODO(), input, func(o *s3.PresignOptions) {
		o.Expires = expiry
	})
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned download URL: %w", err)
	}

	return result.URL, nil
}

func (s *R2Storage) Upload(ctx context.Context, key string, data []byte, contentType string, metadata map[string]string) error {
	checksum := computeSHA256(data)

	input := &s3.PutObjectInput{
		Bucket:            aws.String(s.bucketName),
		Key:               aws.String(key),
		Body:              bytes.NewReader(data),
		ContentType:       aws.String(contentType),
		ChecksumAlgorithm: types.ChecksumAlgorithmSha256,
		Metadata:          metadata,
	}

	_, err := s.client.PutObject(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to upload object to R2: %w", err)
	}

	s.logger.Info().
		Str("key", key).
		Str("checksum", checksum).
		Int("size", len(data)).
		Msg("uploaded object to R2")

	return nil
}

func (s *R2Storage) Download(ctx context.Context, key string) ([]byte, error) {
	input := &s3.GetObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(key),
	}

	result, err := s.client.GetObject(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to download object from R2: %w", err)
	}
	defer result.Body.Close()

	data, err := io.ReadAll(result.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read object body: %w", err)
	}

	storedChecksum := ""
	if result.ChecksumSHA256 != nil {
		storedChecksum = *result.ChecksumSHA256
	}

	if storedChecksum != "" {
		actualChecksum := computeSHA256(data)
		if actualChecksum != storedChecksum {
			return nil, fmt.Errorf("checksum mismatch: expected %s, got %s", storedChecksum, actualChecksum)
		}
	}

	return data, nil
}

func (s *R2Storage) Head(ctx context.Context, key string) (*ports.ObjectInfo, error) {
	input := &s3.HeadObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(key),
	}

	result, err := s.client.HeadObject(ctx, input)
	if err != nil {
		if isNotFound(err) {
			return nil, fmt.Errorf("object not found: %s", key)
		}
		return nil, fmt.Errorf("failed to head object in R2: %w", err)
	}

	info := &ports.ObjectInfo{
		Metadata: make(map[string]string),
	}

	if result.ETag != nil {
		info.ETag = *result.ETag
	}
	if result.ChecksumSHA256 != nil {
		info.ChecksumSHA256 = *result.ChecksumSHA256
	}
	if result.ContentLength != nil {
		info.Size = *result.ContentLength
	}
	for k, v := range result.Metadata {
		info.Metadata[k] = v
	}

	return info, nil
}

func (s *R2Storage) Delete(key string) error {
	input := &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(key),
	}

	_, err := s.client.DeleteObject(context.TODO(), input)
	if err != nil {
		return fmt.Errorf("failed to delete object from R2: %w", err)
	}

	return nil
}

func (s *R2Storage) Exists(key string) (bool, error) {
	input := &s3.HeadObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(key),
	}

	_, err := s.client.HeadObject(context.TODO(), input)
	if err != nil {
		if isNotFound(err) {
			return false, nil
		}
		return false, fmt.Errorf("failed to check object existence in R2: %w", err)
	}

	return true, nil
}

func computeSHA256(data []byte) string {
	h := sha256.Sum256(data)
	return hex.EncodeToString(h[:])
}

func isNotFound(err error) bool {
	var apiErr *smithy.GenericAPIError
	if errors.As(err, &apiErr) && apiErr.ErrorCode() == "NotFound" {
		return true
	}
	return false
}
