package storage

import (
	"bytes"
	"errors"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync/atomic"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
)

const (
	savePath = "notes"
)

var (
	currentStorageSize int64
	maxStorageSize     int64 = 10 * 1024 * 1024 // Default 10MB
	historyResetHours  int   = 72               // Default 72 hours
	ErrStorageTooLarge       = errors.New("storage size exceeds the limit")
)

func InitStorage(maxSize int64, resetHours int) {
	maxStorageSize = maxSize
	historyResetHours = resetHours
	if err := os.MkdirAll(savePath, 0755); err != nil {
		log.Fatalf("failed to create storage directory: %v", err)
	}

	// Initial check and potential init
	PruneHistory()

	// --- Calculate Storage Size ---
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

func PruneHistory() {
	r, err := git.PlainOpen(savePath)
	if err == nil { // Repository exists
		head, err := r.Head()
		if err == nil { // Head exists (i.e., there are commits)
			lastCommit, err := r.CommitObject(head.Hash())
			if err == nil {
				if time.Since(lastCommit.Author.When) > time.Duration(historyResetHours)*time.Hour {
					log.Println("No activity, resetting git history.")
					if err := os.RemoveAll(filepath.Join(savePath, ".git")); err != nil {
						log.Printf("Failed to remove old .git directory: %v", err)
						return
					}
					// Re-initialize
					if _, err := git.PlainInit(savePath, false); err != nil {
						log.Printf("Failed to re-initialize git repository: %v", err)
					}
				}
			}
		}
	} else if err == git.ErrRepositoryNotExists {
		// Initialize if it doesn't exist
		if _, err := git.PlainInit(savePath, false); err != nil {
			log.Printf("Failed to initialize git repository: %v", err)
		}
	} else {
		log.Printf("Failed to open git repository: %v", err)
	}
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

	newContentBytes, err := io.ReadAll(contentReader)
	if err != nil {
		return err
	}

	// Normalize content: trim trailing newlines to a maximum of 3
	re := regexp.MustCompile(`(\s*\n){4,}$`)
	normalizedContentBytes := re.ReplaceAll(newContentBytes, []byte("\n\n\n"))
	contentLength = int64(len(normalizedContentBytes))

	filePath := filepath.Join(savePath, path)

	// Check if there's a meaningful change
	oldContent, err := os.ReadFile(filePath)
	if err == nil { // File exists
		if bytes.Equal(bytes.TrimSpace(oldContent), bytes.TrimSpace(normalizedContentBytes)) {
			return nil // No meaningful change, do nothing
		}
	}

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

	if err := os.WriteFile(filePath, normalizedContentBytes, 0644); err != nil {
		return err
	}

	atomic.AddInt64(&currentStorageSize, sizeDelta)

	// Git versioning
	r, err := git.PlainOpen(savePath)
	if err != nil {
		log.Printf("failed to open git repository: %v", err)
		return nil // Continue even if git fails
	}

	w, err := r.Worktree()
	if err != nil {
		log.Printf("failed to get worktree: %v", err)
		return nil
	}

	_, err = w.Add(path)
	if err != nil {
		log.Printf("failed to git add %s: %v", path, err)
		return nil
	}

	_, err = w.Commit("Update "+path, &git.CommitOptions{
		Author: &object.Signature{
			Name:  "Webnote",
			Email: "webnote@example.com",
			When:  time.Now(),
		},
	})
	if err != nil {
		log.Printf("failed to git commit %s: %v", path, err)
	}
	return nil
}

func IsStorageOverloaded() bool {
	return atomic.LoadInt64(&currentStorageSize) > maxStorageSize
}

func GetNoteVersion(path string, version int) ([]byte, int, error) {
	if !IsValidPath(path) {
		return nil, 0, errors.New("invalid path")
	}

	r, err := git.PlainOpen(savePath)
	if err != nil {
		return nil, 0, err
	}

	logOptions := &git.LogOptions{FileName: &path}
	cIter, err := r.Log(logOptions)
	if err != nil {
		return nil, 0, err
	}
	defer cIter.Close()

	var commits []*object.Commit
	err = cIter.ForEach(func(c *object.Commit) error {
		commits = append(commits, c)
		return nil
	})
	if err != nil {
		return nil, 0, err
	}

	totalVersions := len(commits)

	// The first commit (index 0) is the current state.
	// History starts from version 1, which corresponds to index 1 in the commits slice.
	if version <= 0 || version >= totalVersions {
		return nil, 0, errors.New("version not found")
	}

	commit := commits[version]
	file, err := commit.File(path)
	if err != nil {
		return nil, 0, err
	}

	content, err := file.Contents()
	if err != nil {
		return nil, 0, err
	}

	return []byte(content), totalVersions, nil
}
