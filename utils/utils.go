package utils

import (
	"image"
	"io"
	"mime"
)

// Guess image mime types from gif/jpeg/png/webp
func GuessImageFormat(r io.Reader) (format string, err error) {
	_, format, err = image.DecodeConfig(r)
	return
}

// Guess image mime types from gif/jpeg/png/webp
func GuessImageMimeTypes(r io.Reader) string {
	format, _ := GuessImageFormat(r)
	if format == "" {
		return ""
	}
	return mime.TypeByExtension("." + format)
}
