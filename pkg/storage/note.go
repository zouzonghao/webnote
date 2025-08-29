package storage

import (
	"errors"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
)

const (
	savePath = "notes"
)

var (
	currentStorageSize int64
	maxStorageSize     int64 = 10 * 1024 * 1024 // Default 10MB
	ErrStorageTooLarge       = errors.New("storage size exceeds the limit")
)

func InitStorage(maxSize int64) {
	maxStorageSize = maxSize
	if err := os.MkdirAll(savePath, 0755); err != nil {
		log.Fatalf("failed to create storage directory: %v", err)
	}

	var totalSize int64
	err := filepath.Walk(savePath, func(_ string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			totalSize += info.Size()
		}
		return nil
	})
	if err != nil {
		log.Fatalf("failed to calculate initial storage size: %v", err)
	}
	atomic.StoreInt64(&currentStorageSize, totalSize)
}

func IsValidPath(path string) bool {
	return !strings.Contains(path, "..") && !strings.Contains(path, "/")
}

func GetNote(path string) (*os.File, int64, error) {
	if !IsValidPath(path) {
		return nil, 0, errors.New("invalid path")
	}
	filePath := filepath.Join(savePath, path)
	info, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, 0, nil // Use nil to indicate not found
		}
		return nil, 0, err
	}

	file, err := os.Open(filePath)
	if err != nil {
		return nil, 0, err
	}
	return file, info.Size(), nil
}

func SaveNote(path string, contentReader io.Reader, contentLength int64) error {
	if !IsValidPath(path) {
		return errors.New("invalid path")
	}

	filePath := filepath.Join(savePath, path)

	var oldSize int64
	info, err := os.Stat(filePath)
	if err == nil {
		oldSize = info.Size()
	} else if !os.IsNotExist(err) {
		return err
	}

	if contentLength == 0 {
		if oldSize > 0 {
			if err := os.Remove(filePath); err != nil {
				return err
			}
			atomic.AddInt64(&currentStorageSize, -oldSize)
		}
		return nil
	}

	newSize := contentLength
	sizeDelta := newSize - oldSize

	if atomic.LoadInt64(&currentStorageSize)+sizeDelta > maxStorageSize {
		return ErrStorageTooLarge
	}

	tempFile, err := os.CreateTemp(savePath, "temp-note-")
	if err != nil {
		return err
	}
	defer os.Remove(tempFile.Name())

	written, err := io.Copy(tempFile, contentReader)
	if err != nil {
		tempFile.Close()
		return err
	}
	tempFile.Close()

	if written != newSize {
		return errors.New("content length mismatch")
	}

	if err := os.Rename(tempFile.Name(), filePath); err != nil {
		return err
	}

	atomic.AddInt64(&currentStorageSize, sizeDelta)
	return nil
}

func IsStorageOverloaded() bool {
	return atomic.LoadInt64(&currentStorageSize) > maxStorageSize
}
