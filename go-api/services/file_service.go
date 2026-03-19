package services

import (
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

var (
	ErrFileTooLarge       = errors.New("file size exceeds maximum allowed")
	ErrInvalidFileFormat  = errors.New("invalid file format")
	ErrFileUploadFailed   = errors.New("file upload failed")
)

type FileServiceInterface interface {
	SaveImage(file *multipart.FileHeader) (string, error)
	DeleteImage(filename string) error
}

type FileService struct {
	uploadDir    string
	maxFileSize  int64
	allowedExts  []string
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
		return "", fmt.Errorf("%w: %v", ErrFileUploadFailed, err)
	}
	defer src.Close()

	dst, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrFileUploadFailed, err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		os.Remove(filePath)
		return "", fmt.Errorf("%w: %v", ErrFileUploadFailed, err)
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
