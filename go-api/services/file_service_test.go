package services

import (
	"bytes"
	"mime/multipart"
	"net/http/httptest"
	"os"
	"path/filepath"
	"poc-gin/config"
	"strings"
	"testing"
)

func buildMultipartHeader(t *testing.T, fieldName, filename string, content []byte) *multipart.FileHeader {
	t.Helper()

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile(fieldName, filename)
	if err != nil {
		t.Fatalf("failed to create form file: %v", err)
	}
	if _, err := part.Write(content); err != nil {
		t.Fatalf("failed to write form file: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("failed to close writer: %v", err)
	}

	req := httptest.NewRequest("POST", "/", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	if err := req.ParseMultipartForm(int64(len(content)) + 1024); err != nil {
		t.Fatalf("failed to parse multipart form: %v", err)
	}
	return req.MultipartForm.File[fieldName][0]
}

func TestNewFileService(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "uploads")
	service, err := NewFileService(&config.UploadConfig{Dir: dir, MaxFileSize: 1024})
	if err != nil {
		t.Fatalf("expected service init, got %v", err)
	}
	if service.uploadDir != dir {
		t.Fatalf("unexpected upload dir: %s", service.uploadDir)
	}
	if _, err := os.Stat(dir); err != nil {
		t.Fatalf("expected upload dir to exist: %v", err)
	}
}

func TestFileServiceSaveImage(t *testing.T) {
	dir := t.TempDir()
	service, err := NewFileService(&config.UploadConfig{Dir: dir, MaxFileSize: 10})
	if err != nil {
		t.Fatalf("expected service init, got %v", err)
	}

	t.Run("rejects oversized files", func(t *testing.T) {
		file := buildMultipartHeader(t, "image", "photo.png", []byte("01234567890"))
		_, err := service.SaveImage(file)
		if err != ErrFileTooLarge {
			t.Fatalf("expected ErrFileTooLarge, got %v", err)
		}
	})

	t.Run("rejects invalid extensions", func(t *testing.T) {
		file := buildMultipartHeader(t, "image", "photo.gif", []byte("1234"))
		_, err := service.SaveImage(file)
		if err != ErrInvalidFileFormat {
			t.Fatalf("expected ErrInvalidFileFormat, got %v", err)
		}
	})

	t.Run("saves allowed images", func(t *testing.T) {
		file := buildMultipartHeader(t, "image", "photo.png", []byte("1234"))
		filename, err := service.SaveImage(file)
		if err != nil {
			t.Fatalf("expected image save, got %v", err)
		}
		if !strings.HasSuffix(filename, ".png") {
			t.Fatalf("expected png filename, got %s", filename)
		}
		if _, err := os.Stat(filepath.Join(dir, filename)); err != nil {
			t.Fatalf("expected file on disk: %v", err)
		}
	})
}

func TestFileServiceDeleteImage(t *testing.T) {
	dir := t.TempDir()
	service, err := NewFileService(&config.UploadConfig{Dir: dir, MaxFileSize: 1024})
	if err != nil {
		t.Fatalf("expected service init, got %v", err)
	}

	t.Run("ignores empty filename", func(t *testing.T) {
		if err := service.DeleteImage(""); err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}
	})

	t.Run("ignores missing files", func(t *testing.T) {
		if err := service.DeleteImage("missing.png"); err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}
	})

	t.Run("deletes only basename target", func(t *testing.T) {
		filename := "kept.png"
		path := filepath.Join(dir, filename)
		if err := os.WriteFile(path, []byte("content"), 0644); err != nil {
			t.Fatalf("failed to seed file: %v", err)
		}

		if err := service.DeleteImage(filepath.Join("..", filename)); err != nil {
			t.Fatalf("expected file deletion, got %v", err)
		}
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Fatalf("expected file to be deleted, got err=%v", err)
		}
	})
}
