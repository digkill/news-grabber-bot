package helpers

import (
	"encoding/base64"
	"fmt"
	"mime"
	"strings"
)

func EncodeImageToBase64(imageBytes []byte, fileMimeType string) (string, error) {

	// Кодируем в base64
	base64Str := base64.StdEncoding.EncodeToString(imageBytes)

	// Определяем MIME-тип по расширению
	mimeType := mime.TypeByExtension(fileMimeType)
	if mimeType == "" {
		mimeType = "application/octet-stream" // По умолчанию, если неизвестный тип
	}

	// Формируем data URL
	dataURL := fmt.Sprintf("data:%s;base64,%s", mimeType, base64Str)

	return dataURL, nil
}

func IsImageOrVideo(ext string) bool {
	switch strings.ToLower(ext) {
	case ".jpg", ".jpeg", ".png", ".gif", ".mp4", ".mov", ".avi":
		return true
	default:
		return false
	}
}
