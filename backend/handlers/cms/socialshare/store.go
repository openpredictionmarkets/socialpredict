package socialshare

import (
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

const uploadFileBaseName = "social-share-image"

type UploadedImage struct {
	ContentType string
	SizeBytes   int64
	Data        []byte
}

type ImageStore interface {
	Save(data []byte, contentType string) error
	Load() (*UploadedImage, error)
}

type FileImageStore struct {
	dir string
}

func NewDefaultImageStore() *FileImageStore {
	dir := strings.TrimSpace(os.Getenv("SOCIAL_SHARE_UPLOAD_DIR"))
	if dir == "" {
		dir = filepath.Join("data", "uploads", "social-share")
	}
	return NewFileImageStore(dir)
}

func NewFileImageStore(dir string) *FileImageStore {
	return &FileImageStore{dir: dir}
}

func (s *FileImageStore) Save(data []byte, contentType string) error {
	if s == nil || strings.TrimSpace(s.dir) == "" {
		return errors.New("social share image store is unavailable")
	}
	if err := os.MkdirAll(s.dir, 0o755); err != nil {
		return err
	}
	if err := s.removeExisting(); err != nil {
		return err
	}
	path := filepath.Join(s.dir, uploadFileBaseName+extensionForContentType(contentType))
	tmpPath := path + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmpPath, path)
}

func (s *FileImageStore) Load() (*UploadedImage, error) {
	if s == nil || strings.TrimSpace(s.dir) == "" {
		return nil, os.ErrNotExist
	}
	for _, contentType := range supportedImageContentTypes() {
		path := filepath.Join(s.dir, uploadFileBaseName+extensionForContentType(contentType))
		data, err := os.ReadFile(path)
		if err == nil {
			return &UploadedImage{
				ContentType: contentType,
				SizeBytes:   int64(len(data)),
				Data:        data,
			}, nil
		}
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			return nil, err
		}
	}
	return nil, os.ErrNotExist
}

func (s *FileImageStore) removeExisting() error {
	for _, contentType := range supportedImageContentTypes() {
		path := filepath.Join(s.dir, uploadFileBaseName+extensionForContentType(contentType))
		if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
			return err
		}
	}
	return nil
}

func supportedImageContentTypes() []string {
	return []string{"image/png", "image/jpeg", "image/webp"}
}

func extensionForContentType(contentType string) string {
	switch contentType {
	case "image/jpeg":
		return ".jpg"
	case "image/webp":
		return ".webp"
	default:
		return ".png"
	}
}

func DetectUploadedImageContentType(data []byte) string {
	return http.DetectContentType(data)
}
