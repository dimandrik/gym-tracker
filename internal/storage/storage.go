package storage

import (
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
)

// имя файла — случайный UUID с исходным расширением, оригинальное имя отбрасывается
// во избежание коллизий и path traversal
func SaveFile(file multipart.File, header *multipart.FileHeader, uploadDir string) (string, error) {
	ext := filepath.Ext(header.Filename)
	filename := uuid.New().String() + ext
	fullPath := filepath.Join(uploadDir, filename)

	dst, err := os.Create(fullPath)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %w", err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		return "", fmt.Errorf("failed to save file: %w", err)
	}

	return "/uploads/" + filename, nil
}

func DeleteFile(photoURL, uploadDir string) error {
	if photoURL == "" {
		return nil
	}

	filename := strings.TrimPrefix(photoURL, "/uploads/")
	fullPath := filepath.Join(uploadDir, filename)
	if err := os.Remove(fullPath); err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}

	return nil
}
