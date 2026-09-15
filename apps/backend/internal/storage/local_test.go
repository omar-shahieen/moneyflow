package storage

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLocalStorage_Upload_Head_Download_Checksum(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewLocalStorage(tmpDir)
	ctx := context.Background()

	key := "test/user1/import1.csv"
	data := []byte("CategoryName,Amount,Note,Date\nFood,1000,lunch,2024-01-15\n")

	err := store.Upload(ctx, key, data, "text/csv", map[string]string{"status": "pending"})
	require.NoError(t, err)

	info, err := store.Head(ctx, key)
	require.NoError(t, err)
	assert.Equal(t, int64(len(data)), info.Size)
	assert.NotEmpty(t, info.ChecksumSHA256)

	downloaded, err := store.Download(ctx, key)
	require.NoError(t, err)
	assert.Equal(t, data, downloaded)

	fullPath := filepath.Join(tmpDir, key)
	_, err = os.Stat(fullPath + ".sha256")
	assert.NoError(t, err, "sidecar checksum file should exist")
}

func TestLocalStorage_Download_ChecksumMismatch(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewLocalStorage(tmpDir)
	ctx := context.Background()

	key := "test/user1/import1.csv"
	data := []byte("CategoryName,Amount,Note,Date\nFood,1000,lunch,2024-01-15\n")

	err := store.Upload(ctx, key, data, "text/csv", nil)
	require.NoError(t, err)

	fullPath := filepath.Join(tmpDir, key)
	err = os.WriteFile(fullPath+".sha256", []byte("0000000000000000000000000000000000000000000000000000000000000000"), 0644)
	require.NoError(t, err)

	_, err = store.Download(ctx, key)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "checksum mismatch")
}

func TestLocalStorage_Head_NotFound(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewLocalStorage(tmpDir)
	ctx := context.Background()

	_, err := store.Head(ctx, "nonexistent/file.csv")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "object not found")
}

func TestLocalStorage_Download_NotFound(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewLocalStorage(tmpDir)
	ctx := context.Background()

	_, err := store.Download(ctx, "nonexistent/file.csv")
	assert.Error(t, err)
}

func TestLocalStorage_Exists(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewLocalStorage(tmpDir)
	ctx := context.Background()

	key := "test/file.csv"
	exists, err := store.Exists(key)
	require.NoError(t, err)
	assert.False(t, exists)

	err = store.Upload(ctx, key, []byte("data"), "text/csv", nil)
	require.NoError(t, err)

	exists, err = store.Exists(key)
	require.NoError(t, err)
	assert.True(t, exists)
}
