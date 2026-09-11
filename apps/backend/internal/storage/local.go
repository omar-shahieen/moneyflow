package storage

import (
	"fmt"
	"time"
)

type LocalStorage struct {
	basePath string
}

func NewLocalStorage(basePath string) *LocalStorage {
	return &LocalStorage{basePath: basePath}
}

func (s *LocalStorage) GenerateUploadURL(key string, contentType string, expiry time.Duration) (string, error) {
	url := fmt.Sprintf("/uploads/%s?expires=%d", key, time.Now().Add(expiry).Unix())
	return url, nil
}

func (s *LocalStorage) GenerateDownloadURL(key string, expiry time.Duration) (string, error) {
	url := fmt.Sprintf("/downloads/%s?expires=%d", key, time.Now().Add(expiry).Unix())
	return url, nil
}

func (s *LocalStorage) Delete(key string) error {
	return nil
}

func (s *LocalStorage) Exists(key string) (bool, error) {
	return false, nil
}
