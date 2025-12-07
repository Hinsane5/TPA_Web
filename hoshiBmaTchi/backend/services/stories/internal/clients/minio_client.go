package clients

import (
	"context"
	"fmt"
	"log"
	"mime/multipart"

	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type MinioClient struct {
	client     *minio.Client

	bucketName string
	endpoint   string

}

func NewMinioClient(endpoint, accessKey, secretKey, bucketName string, useSSL bool) (*MinioClient, error) {
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create minio client: %w", err)
	}

	ctx := context.Background()
	exists, err := client.BucketExists(ctx, bucketName)
	if err != nil {
		return nil, fmt.Errorf("failed to check bucket: %w", err)
	}

	if !exists {
		err = client.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{})
		if err != nil {
			return nil, fmt.Errorf("failed to create bucket: %w", err)
		}
		log.Printf("Created bucket: %s", bucketName)

		policy := fmt.Sprintf(`{
			"Version": "2012-10-17",
			"Statement": [
				{
					"Effect": "Allow",
					"Principal": {"AWS": ["*"]},
					"Action": ["s3:GetObject"],
					"Resource": ["arn:aws:s3:::%s/*"]
				}
			]
		}`, bucketName)

		if err := client.SetBucketPolicy(ctx, bucketName, policy); err != nil {
			log.Printf("Warning: failed to set bucket policy: %v", err)
		}











	}

	log.Printf("Connected to MinIO at %s, bucket: %s", endpoint, bucketName)

	return &MinioClient{
		client:     client,
		bucketName: bucketName,
		endpoint:   endpoint,


	}, nil
}

func (m *MinioClient) UploadStory(ctx context.Context, file multipart.File, header *multipart.FileHeader, userID string) (string, error) {
	ext := filepath.Ext(header.Filename)
	filename := fmt.Sprintf("stories/%s/%s%s", userID, uuid.New().String(), ext)


	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	_, err := m.client.PutObject(ctx, m.bucketName, filename, file, header.Size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload file: %w", err)
	}

	url := fmt.Sprintf("http://%s/%s/%s", m.endpoint, m.bucketName, filename)
	log.Printf("Uploaded story to: %s", url)

	return url, nil
}

func (m *MinioClient) DeleteStory(ctx context.Context, mediaURL string) error {
	objectName := extractObjectName(mediaURL, m.endpoint, m.bucketName)
	if objectName == "" {
		return fmt.Errorf("invalid media URL")

	}

	err := m.client.RemoveObject(ctx, m.bucketName, objectName, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}

	log.Printf("Deleted story: %s", objectName)
	return nil
}

func (m *MinioClient) GetPresignedURL(ctx context.Context, mediaURL string, expires time.Duration) (string, error) {
	objectName := extractObjectName(mediaURL, m.endpoint, m.bucketName)



	if objectName == "" {
		return "", fmt.Errorf("invalid media URL")
	}

	presignedURL, err := m.client.PresignedGetObject(ctx, m.bucketName, objectName, expires, nil)






	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	return presignedURL.String(), nil
}

func extractObjectName(url, endpoint, bucket string) string {
	prefix1 := fmt.Sprintf("http://%s/%s/", endpoint, bucket)
	prefix2 := fmt.Sprintf("https://%s/%s/", endpoint, bucket)

	if strings.HasPrefix(url, prefix1) {
		return strings.TrimPrefix(url, prefix1)
	}
	if strings.HasPrefix(url, prefix2) {
		return strings.TrimPrefix(url, prefix2)



	}

	return ""
}
