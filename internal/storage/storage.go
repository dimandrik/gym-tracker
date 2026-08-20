package storage

import (
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
)

var ErrInvalidFileType = errors.New("invalid file type")

// только типы изображений — загрузка произвольных файлов (html, скрипты и т.п.)
// под видом фото тренажера не должна проходить, даже если раздаётся как статика
var allowedContentTypes = map[string]bool{
	"image/jpeg": true,
	"image/png":  true,
	"image/gif":  true,
	"image/webp": true,
}

// имя файла — случайный UUID с исходным расширением, оригинальное имя отбрасывается
// во избежание коллизий и path traversal
func SaveFile(file multipart.File, header *multipart.FileHeader, uploadDir string) (string, error) {
	// тип определяется по содержимому файла, а не по расширению или Content-Type из формы —
	// оба этих значения задаёт клиент и легко подделать
	contentType, err := detectContentType(file)
	if err != nil {
		return "", fmt.Errorf("failed to detect content type: %w", err)
	}
	if !allowedContentTypes[contentType] {
		return "", ErrInvalidFileType
	}

	// расширение берётся из имени файла, а не из проверенного contentType — это только
	// влияет на итоговое имя файла, само содержимое уже провалидировано выше
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

// http.DetectContentType смотрит только на первые 512 байт, поэтому дальше
// курсор возвращается в начало файла — иначе io.Copy в SaveFile обрежет начало
func detectContentType(file multipart.File) (string, error) {
	buffer := make([]byte, 512)
	_, err := file.Read(buffer)
	if err != nil {
		return "", err
	}

	if _, err := file.Seek(0, 0); err != nil {
		return "", err
	}

	return http.DetectContentType(buffer), nil
}
