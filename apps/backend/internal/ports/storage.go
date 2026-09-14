package ports

import "time"

type Storage interface {
	GenerateUploadURL(key string, contentType string, expiry time.Duration) (string, error)
	GenerateDownloadURL(key string, expiry time.Duration) (string, error)
	Delete(key string) error
	Exists(key string) (bool, error)
}
