package ports

import (
	"context"
	"time"
)

type ObjectInfo struct {
	ETag            string
	ChecksumSHA256  string
	Metadata        map[string]string
	Size            int64
}

type Storage interface {
	GenerateUploadURL(key string, contentType string, expiry time.Duration) (string, error)
	GenerateDownloadURL(key string, expiry time.Duration) (string, error)
	Upload(ctx context.Context, key string, data []byte, contentType string, metadata map[string]string) error
	Download(ctx context.Context, key string) ([]byte, error)
	Head(ctx context.Context, key string) (*ObjectInfo, error)
	Delete(key string) error
	Exists(key string) (bool, error)
}
