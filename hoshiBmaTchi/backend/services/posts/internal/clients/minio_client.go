package clients

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

const (
	maxRetries     = 10
	retryDelay     = 5 * time.Second
	connectTimeout = 10 * time.Second
)

func NewMinIOClient() (*minio.Client, error) {
	endpoint := os.Getenv("MINIO_ENDPOINT")
	accessKeyID := os.Getenv("MINIO_ACCESS_KEY_ID")
	secretAccessKey := os.Getenv("MINIO_SECRET_ACCESS_KEY")
	bucketName := os.Getenv("MINIO_BUCKET_NAME")
	useSSL := os.Getenv("MINIO_USE_SSL") == "true"

	// Validate environment variables
	if endpoint == "" || accessKeyID == "" || secretAccessKey == "" || bucketName == "" {
		return nil, fmt.Errorf("missing required MinIO environment variables")
	}

	log.Printf("Connecting to MinIO at %s (SSL: %v)", endpoint, useSSL)

	var minioClient *minio.Client
	var err error

	// Retry connection with exponential backoff
	for attempt := 1; attempt <= maxRetries; attempt++ {
		minioClient, err = minio.New(endpoint, &minio.Options{
			Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
			Secure: useSSL,
		})

		if err != nil {
			log.Printf("Failed to create MinIO client (attempt %d/%d): %v", attempt, maxRetries, err)
			if attempt < maxRetries {
				time.Sleep(retryDelay)
				continue
			}
			return nil, fmt.Errorf("failed to create MinIO client after %d attempts: %w", maxRetries, err)
		}

		// Test connection
		ctx, cancel := context.WithTimeout(context.Background(), connectTimeout)
		defer cancel()

		exists, errBucketExists := minioClient.BucketExists(ctx, bucketName)
		if errBucketExists != nil {
			log.Printf("Failed to check bucket (attempt %d/%d): %v", attempt, maxRetries, errBucketExists)
			if attempt < maxRetries {
				time.Sleep(retryDelay)
				continue
			}
			return nil, fmt.Errorf("failed to verify MinIO connection: %w", errBucketExists)
		}

		// Create bucket if it doesn't exist
		if !exists {
			ctx, cancel := context.WithTimeout(context.Background(), connectTimeout)
			defer cancel()

			err = minioClient.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{})
			if err != nil {
				log.Printf("Failed to create bucket (attempt %d/%d): %v", attempt, maxRetries, err)
				if attempt < maxRetries {
					time.Sleep(retryDelay)
					continue
				}
				return nil, fmt.Errorf("failed to create bucket: %w", err)
			}
			log.Printf("Successfully created bucket '%s'", bucketName)
		} else {
			log.Printf("Bucket '%s' already exists", bucketName)
		}

		// Set bucket policy for public read access (required for presigned URLs to work)
		if err := setBucketPolicy(minioClient, bucketName); err != nil {
			// Don't fail if policy setting fails - log warning and continue
			log.Printf("Warning: Failed to set bucket policy (this is optional): %v", err)
		}

		log.Println("Successfully connected to MinIO")
		return minioClient, nil
	}

	return nil, fmt.Errorf("failed to initialize MinIO client after %d attempts", maxRetries)
}

// setBucketPolicy sets a policy that allows public read access to objects
// This is necessary for presigned URLs to work properly
func setBucketPolicy(client *minio.Client, bucketName string) error {
	ctx, cancel := context.WithTimeout(context.Background(), connectTimeout)
	defer cancel()

	// Policy that allows public read access
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

	err := client.SetBucketPolicy(ctx, bucketName, policy)
	if err != nil {
		return fmt.Errorf("failed to set bucket policy: %w", err)
	}

	

	log.Printf("Successfully set bucket policy for '%s'", bucketName)
	return nil
}

