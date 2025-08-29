package storage

import (
	"context"
	"fmt"
	"mime/multipart"
	"path/filepath"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/spf13/viper"
)

type StorageService interface {
	UploadFile(ctx context.Context, file multipart.File, header *multipart.FileHeader, folder string, userID uint) (string, error)
	GetFileURL(ctx context.Context, key string) (string, error)
	DeleteFile(ctx context.Context, key string) error
}

type MinIOService struct {
	client *minio.Client
	bucket string
}

func NewS3Service() (StorageService, error) {
	accessKey := viper.GetString("AWS_ACCESS_KEY_ID")
	secretKey := viper.GetString("AWS_SECRET_ACCESS_KEY")
	region := viper.GetString("AWS_REGION")
	bucket := viper.GetString("AWS_S3_BUCKET")
	
	if region == "" {
		region = "us-east-1"
	}
	if bucket == "" {
		bucket = "salesrep-uploads"
	}

	endpoint := fmt.Sprintf("s3.%s.amazonaws.com", region)
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: true,
		Region: region,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create S3 client: %w", err)
	}

	service := &MinIOService{
		client: client,
		bucket: bucket,
	}

	return service, nil
}

func (s *MinIOService) ensureBucket(ctx context.Context) error {
	exists, err := s.client.BucketExists(ctx, s.bucket)
	if err != nil {
		return err
	}

	if !exists {
		err = s.client.MakeBucket(ctx, s.bucket, minio.MakeBucketOptions{})
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *MinIOService) UploadFile(ctx context.Context, file multipart.File, header *multipart.FileHeader, folder string, userID uint) (string, error) {
	timestamp := time.Now().Unix()
	ext := filepath.Ext(header.Filename)
	filename := fmt.Sprintf("%d_%d%s", userID, timestamp, ext)
	objectKey := filepath.Join(folder, filename)

	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	_, err := s.client.PutObject(ctx, s.bucket, objectKey, file, header.Size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload file to MinIO: %w", err)
	}

	return objectKey, nil
}

func (s *MinIOService) GetFileURL(ctx context.Context, key string) (string, error) {
	url, err := s.client.PresignedGetObject(ctx, s.bucket, key, 7*24*time.Hour, nil)
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	return url.String(), nil
}

func (s *MinIOService) DeleteFile(ctx context.Context, key string) error {
	err := s.client.RemoveObject(ctx, s.bucket, key, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete file from MinIO: %w", err)
	}

	return nil
}

type LocalStorageService struct {
	basePath string
}

func NewLocalStorageService() StorageService {
	basePath := viper.GetString("STORAGE_PATH")
	if basePath == "" {
		basePath = "uploads"
	}

	return &LocalStorageService{
		basePath: basePath,
	}
}

func (s *LocalStorageService) UploadFile(ctx context.Context, file multipart.File, header *multipart.FileHeader, folder string, userID uint) (string, error) {
	return "", fmt.Errorf("local storage not implemented for file uploads")
}

func (s *LocalStorageService) GetFileURL(ctx context.Context, key string) (string, error) {
	return fmt.Sprintf("/uploads/%s", key), nil
}

func (s *LocalStorageService) DeleteFile(ctx context.Context, key string) error {
	return fmt.Errorf("local storage delete not implemented")
}