package mocks

import (
	"context"
	"mime/multipart"

	"github.com/Ruclo/MyMeals/internal/storage"
	"github.com/stretchr/testify/mock"
)

// MockImageStorage is a mock implementation of storage.ImageStorage
type MockImageStorage struct {
	mock.Mock
}

// Upload mocks the Upload method
func (m *MockImageStorage) Upload(ctx context.Context, file *multipart.FileHeader) (*storage.ImageResult, error) {
	args := m.Called(ctx, file)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*storage.ImageResult), args.Error(1)
}

// UploadCropped mocks the UploadCropped method
func (m *MockImageStorage) UploadCropped(ctx context.Context, file *multipart.FileHeader, width, height int) (*storage.ImageResult, error) {
	args := m.Called(ctx, file, width, height)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*storage.ImageResult), args.Error(1)
}

// Delete mocks the Delete method
func (m *MockImageStorage) Delete(ctx context.Context, publicID string) error {
	args := m.Called(ctx, publicID)
	return args.Error(0)
}
