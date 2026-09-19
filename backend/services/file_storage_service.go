package services

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type FileStorageService struct {
	Root string
}

func NewFileStorageService() *FileStorageService {
	root := os.Getenv("STORAGE_ROOT")
	if root == "" {
		root = "./storage"
	}
	return &FileStorageService{Root: root}
}

// StoredFile describes what was written to disk.
type StoredFile struct {
	RelativePath string
	Hash         string
	Size         int64
	ContentType  string
}

// allowedCategoryDirs maps a logical category to a disk subfolder.
var allowedCategoryDirs = map[string]string{
	"avatar":     "avatars",
	"cover":      "covers",
	"evidence":   "evidence",
	"gov_id":     "gov_ids",
	"voice_note": "voice_notes",
}

// Save persists the uploaded file under Root/<category-dir>/<yyyy>/<mm>/<hash>.<ext>
// Content-addressed: identical files deduplicate on disk.
func (s *FileStorageService) Save(file *multipart.FileHeader, category string) (*StoredFile, error) {
	if file == nil {
		return nil, errors.New("no file provided")
	}

	dirName, ok := allowedCategoryDirs[category]
	if !ok {
		return nil, errors.New("invalid storage category")
	}

	src, err := file.Open()
	if err != nil {
		return nil, errors.New("cannot open file")
	}
	defer src.Close()

	hasher := sha256.New()
	tmpDir := filepath.Join(s.Root, dirName, time.Now().UTC().Format("2006"), time.Now().UTC().Format("01"))
	if err := os.MkdirAll(tmpDir, 0o755); err != nil {
		return nil, err
	}

	tmpFile, err := os.CreateTemp(tmpDir, "upload-*.tmp")
	if err != nil {
		return nil, err
	}
	tmpName := tmpFile.Name()
	defer os.Remove(tmpName)

	size, err := io.Copy(io.MultiWriter(tmpFile, hasher), src)
	if err != nil {
		tmpFile.Close()
		return nil, err
	}
	if err := tmpFile.Close(); err != nil {
		return nil, err
	}

	hash := hex.EncodeToString(hasher.Sum(nil))
	ext := strings.ToLower(filepath.Ext(file.Filename))
	if ext == "" {
		ext = ".bin"
	}

	finalName := hash + ext
	finalPath := filepath.Join(tmpDir, finalName)

	// Deduplicate: if the same hash already exists, discard the temp and reuse.
	if _, err := os.Stat(finalPath); err == nil {
		_ = os.Remove(tmpName)
	} else {
		if err := os.Rename(tmpName, finalPath); err != nil {
			return nil, err
		}
	}

	rel, err := filepath.Rel(s.Root, finalPath)
	if err != nil {
		rel = finalPath
	}

	return &StoredFile{
		RelativePath: filepath.ToSlash(rel),
		Hash:         hash,
		Size:         size,
		ContentType:  file.Header.Get("Content-Type"),
	}, nil
}

// Open returns a reader for a stored relative path. Caller must close.
func (s *FileStorageService) Open(relativePath string) (io.ReadCloser, error) {
	clean := filepath.Clean(relativePath)
	if strings.Contains(clean, "..") {
		return nil, errors.New("invalid path")
	}
	full := filepath.Join(s.Root, clean)
	return os.Open(full)
}

// Delete removes a stored file. Missing files are not an error.
func (s *FileStorageService) Delete(relativePath string) error {
	clean := filepath.Clean(relativePath)
	if strings.Contains(clean, "..") {
		return errors.New("invalid path")
	}
	full := filepath.Join(s.Root, clean)
	err := os.Remove(full)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

// Exists reports whether a stored file is present.
func (s *FileStorageService) Exists(relativePath string) bool {
	clean := filepath.Clean(relativePath)
	if strings.Contains(clean, "..") {
		return false
	}
	full := filepath.Join(s.Root, clean)
	_, err := os.Stat(full)
	return err == nil
}
