package storage

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/omar-shahieen/moneyflow/internal/ports"
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

func (s *LocalStorage) Upload(ctx context.Context, key string, data []byte, contentType string, metadata map[string]string) error {
	fullPath := filepath.Join(s.basePath, key)
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	if err := os.WriteFile(fullPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write file %s: %w", fullPath, err)
	}

	checksum := computeChecksum(data)
	checksumPath := fullPath + ".sha256"
	if err := os.WriteFile(checksumPath, []byte(checksum), 0644); err != nil {
		return fmt.Errorf("failed to write checksum file %s: %w", checksumPath, err)
	}

	return nil
}

func (s *LocalStorage) Download(ctx context.Context, key string) ([]byte, error) {
	fullPath := filepath.Join(s.basePath, key)

	data, err := os.ReadFile(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", fullPath, err)
	}

	checksumPath := fullPath + ".sha256"
	expectedChecksum, err := os.ReadFile(checksumPath)
	if err == nil {
		actualChecksum := computeChecksum(data)
		if actualChecksum != string(expectedChecksum) {
			return nil, fmt.Errorf("checksum mismatch: expected %s, got %s", string(expectedChecksum), actualChecksum)
		}
	}

	return data, nil
}

func (s *LocalStorage) Head(ctx context.Context, key string) (*ports.ObjectInfo, error) {
	fullPath := filepath.Join(s.basePath, key)

	stat, err := os.Stat(fullPath)
	if os.IsNotExist(err) {
		return nil, fmt.Errorf("object not found: %s", key)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to stat file %s: %w", fullPath, err)
	}

	info := &ports.ObjectInfo{
		Size:     stat.Size(),
		Metadata: make(map[string]string),
	}

	checksumPath := fullPath + ".sha256"
	checksumBytes, err := os.ReadFile(checksumPath)
	if err == nil {
		info.ChecksumSHA256 = string(checksumBytes)
	}

	return info, nil
}

func (s *LocalStorage) Delete(key string) error {
	fullPath := filepath.Join(s.basePath, key)
	checksumPath := fullPath + ".sha256"

	os.Remove(fullPath)
	os.Remove(checksumPath)
	return nil
}

func (s *LocalStorage) Exists(key string) (bool, error) {
	fullPath := filepath.Join(s.basePath, key)
	_, err := os.Stat(fullPath)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func computeChecksum(data []byte) string {
	h := sha256.Sum256(data)
	return hex.EncodeToString(h[:])
}
