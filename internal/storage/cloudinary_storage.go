package storage

import (
	"context"
	"fmt"
	"mime/multipart"

	"github.com/Ruclo/MyMeals/internal/errors"
	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
)

// CloudinaryStorage implements ImageStorage using Cloudinary
type CloudinaryStorage struct {
	client *cloudinary.Cloudinary
}

// NewCloudinaryStorage creates a new Cloudinary storage instance
func NewCloudinaryStorage(client *cloudinary.Cloudinary) ImageStorage {
	return &CloudinaryStorage{client: client}
}

// Upload uploads an image with default parameters
func (c *CloudinaryStorage) Upload(ctx context.Context, file *multipart.FileHeader) (*ImageResult, error) {
	result, err := c.client.Upload.Upload(ctx, file, uploader.UploadParams{
		Transformation: "c_limit,h_1920,w_1920",
	})

	if err != nil {
		return nil, errors.NewInternalServerErr("Failed to upload image", err)
	}

	return &ImageResult{
		URL:      result.SecureURL,
		PublicID: result.PublicID,
	}, nil
}

// UploadCropped uploads an image with cropping parameters
func (c *CloudinaryStorage) UploadCropped(ctx context.Context, file *multipart.FileHeader, width, height int) (*ImageResult, error) {
	transformation := fmt.Sprintf("c_crop,h_%d,w_%d", height, width)
	result, err := c.client.Upload.Upload(ctx, file, uploader.UploadParams{
		Transformation: transformation,
	})

	if err != nil {
		return nil, errors.NewInternalServerErr("Failed to upload cropped image", err)
	}

	return &ImageResult{
		URL:      result.SecureURL,
		PublicID: result.PublicID,
	}, nil
}

// Delete removes an image from storage
func (c *CloudinaryStorage) Delete(ctx context.Context, publicID string) error {
	_, err := c.client.Upload.Destroy(ctx, uploader.DestroyParams{PublicID: publicID})
	if err != nil {
		return errors.NewInternalServerErr("Failed to delete image", err)
	}
	return nil
}
