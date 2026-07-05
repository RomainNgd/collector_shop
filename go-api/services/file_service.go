package services

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"poc-gin/config"
	"poc-gin/pkg/logger"
	"strings"

	"github.com/google/uuid"
)

const (
	imageHeaderBytes = 12
	errorWrapFormat  = "%w: %v"
)

var (
	ErrFileTooLarge      = errors.New("file size exceeds maximum allowed")
	ErrInvalidFileFormat = errors.New("invalid file format")
	ErrFileUploadFailed  = errors.New("file upload failed")
)

type FileServiceInterface interface {
	SaveImage(file *multipart.FileHeader) (string, error)
	DeleteImage(filename string) error
}

type FileService struct {
	uploadDir   string
	maxFileSize int64
	allowedExts []string
}

func NewFileService(cfg *config.UploadConfig) (*FileService, error) {
	if err := os.MkdirAll(cfg.Dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create upload directory: %w", err)
	}

	return &FileService{
		uploadDir:   cfg.Dir,
		maxFileSize: cfg.MaxFileSize,
		allowedExts: []string{".jpg", ".jpeg", ".png", ".webp"},
	}, nil
}

func (s *FileService) SaveImage(file *multipart.FileHeader) (string, error) {
	if file.Size > s.maxFileSize {
		return "", ErrFileTooLarge
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	if !s.isAllowedExtension(ext) {
		return "", ErrInvalidFileFormat
	}

	filename := fmt.Sprintf("%s%s", uuid.New().String(), ext)
	filePath := filepath.Join(s.uploadDir, filename)

	src, err := file.Open()
	if err != nil {
		return "", fmt.Errorf(errorWrapFormat, ErrFileUploadFailed, err)
	}
	defer src.Close()

	header := make([]byte, imageHeaderBytes)
	n, err := io.ReadFull(src, header)
	if err != nil && !errors.Is(err, io.ErrUnexpectedEOF) && !errors.Is(err, io.EOF) {
		return "", fmt.Errorf(errorWrapFormat, ErrFileUploadFailed, err)
	}
	header = header[:n]

	if !isAllowedImageContent(ext, header) {
		return "", ErrInvalidFileFormat
	}

	if _, err := src.Seek(0, io.SeekStart); err != nil {
		return "", fmt.Errorf(errorWrapFormat, ErrFileUploadFailed, err)
	}

	dst, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf(errorWrapFormat, ErrFileUploadFailed, err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		os.Remove(filePath)
		return "", fmt.Errorf(errorWrapFormat, ErrFileUploadFailed, err)
	}

	return filename, nil
}

func (s *FileService) DeleteImage(filename string) error {
	if filename == "" {
		return nil
	}

	filename = filepath.Base(filename)
	filePath := filepath.Join(s.uploadDir, filename)

	if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
		logger.Error("Failed to delete file %s: %v", filename, err)
		return fmt.Errorf("failed to delete file: %w", err)
	}

	return nil
}

func (s *FileService) isAllowedExtension(ext string) bool {
	for _, allowed := range s.allowedExts {
		if ext == allowed {
			return true
		}
	}
	return false
}

func isAllowedImageContent(ext string, header []byte) bool {
	switch ext {
	case ".jpg", ".jpeg":
		return len(header) >= 3 && header[0] == 0xff && header[1] == 0xd8 && header[2] == 0xff
	case ".png":
		return bytes.HasPrefix(header, []byte{0x89, 'P', 'N', 'G', 0x0d, 0x0a, 0x1a, 0x0a})
	case ".webp":
		return len(header) >= 12 &&
			bytes.Equal(header[0:4], []byte("RIFF")) &&
			bytes.Equal(header[8:12], []byte("WEBP"))
	default:
		return false
	}
}
