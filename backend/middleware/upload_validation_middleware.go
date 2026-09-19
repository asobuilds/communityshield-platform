package middleware

import (
	"bytes"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	MaxUploadSize   = 10 << 20 // 10 MB
	MaxImageSize    = 5 << 20  // 5 MB
	MaxDocumentSize = 10 << 20 // 10 MB
)

var allowedImageMIME = map[string]bool{
	"image/jpeg": true,
	"image/png":  true,
	"image/webp": true,
	"image/gif":  true,
}

var allowedDocMIME = map[string]bool{
	"application/pdf": true,
	"image/jpeg":      true,
	"image/png":       true,
}

var magicSignatures = map[string][]byte{
	"image/jpeg":      {0xFF, 0xD8, 0xFF},
	"image/png":       {0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A},
	"image/gif":       {0x47, 0x49, 0x46, 0x38},
	"image/webp":      {0x52, 0x49, 0x46, 0x46},
	"application/pdf": {0x25, 0x50, 0x44, 0x46},
}

func sniffMIME(header []byte) string {
	for mime, sig := range magicSignatures {
		if len(header) >= len(sig) && bytes.HasPrefix(header, sig) {
			return mime
		}
	}
	return ""
}

func isValidExtension(filename string, mime string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	switch mime {
	case "image/jpeg":
		return ext == ".jpg" || ext == ".jpeg"
	case "image/png":
		return ext == ".png"
	case "image/gif":
		return ext == ".gif"
	case "image/webp":
		return ext == ".webp"
	case "application/pdf":
		return ext == ".pdf"
	}
	return false
}

func ValidateUpload(file *multipart.FileHeader, category string) error {
	if file == nil {
		return errors.New("no file provided")
	}

	maxSize := int64(MaxUploadSize)
	if category == "image" {
		maxSize = MaxImageSize
	} else if category == "document" {
		maxSize = MaxDocumentSize
	}
	if category == "evidence" {
		maxSize = 100 << 20 // 100 MB for evidence (video, audio)
	}

	if file.Size > maxSize {
		return errors.New("file exceeds maximum allowed size")
	}

	src, err := file.Open()
	if err != nil {
		return errors.New("cannot open uploaded file")
	}
	defer src.Close()

	header := make([]byte, 512)
	n, _ := io.ReadFull(src, header)
	header = header[:n]

	detected := sniffMIME(header)
	if detected == "" && category != "evidence" {
		return errors.New("file type not recognised")
	}

	if category == "image" {
		if !allowedImageMIME[detected] {
			return errors.New("only JPEG, PNG, WebP or GIF images are allowed")
		}
	} else if category == "document" {
		if !allowedDocMIME[detected] {
			return errors.New("only PDF, JPEG or PNG documents are allowed")
		}
	}
	// category == "evidence": accept any detectable type; virus scan in a later wave

	if detected != "" && !isValidExtension(file.Filename, detected) {
		return errors.New("file extension does not match file contents")
	}

	return nil
}

func UploadValidationMiddleware(category string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !strings.HasPrefix(c.GetHeader("Content-Type"), "multipart/form-data") {
			c.Next()
			return
		}

		form, err := c.MultipartForm()
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid multipart form"})
			c.Abort()
			return
		}

		for _, files := range form.File {
			for _, f := range files {
				if err := ValidateUpload(f, category); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
					c.Abort()
					return
				}
			}
		}

		c.Next()
	}
}
