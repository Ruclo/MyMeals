package storage

import (
	"context"
	"mime/multipart"
)

// ImageResult represents the result of an image upload.
type ImageResult struct {
	URL      string
	PublicID string
}

// ImageStorage defines an interface for image storage operations.
type ImageStorage interface {
	// Upload uploads an image with default parameters.
	Upload(ctx context.Context, file *multipart.FileHeader) (*ImageResult, error)

	// UploadCropped uploads an image with cropping parameters.
	UploadCropped(ctx context.Context, file *multipart.FileHeader, width, height int) (*ImageResult, error)

	// Delete removes an image from storage.
	Delete(ctx context.Context, publicID string) error
}
