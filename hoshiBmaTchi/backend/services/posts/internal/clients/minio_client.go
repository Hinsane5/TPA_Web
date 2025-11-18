package clients

import (
	"context"
	"log"
	"os"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/minio/minio-go/v7/pkg/cors"
)

func NewMinIOClient() (*minio.Client, error) {
	endpoint := os.Getenv("MINIO_ENDPOINT")
	accessKeyID := os.Getenv("MINIO_ACCESS_KEY_ID")
	secretAccessKey := os.Getenv("MINIO_SECRET_ACCESS_KEY")
	bucketName := os.Getenv("MINIO_BUCKET_NAME")
	useSSL := os.Getenv("MINIO_USE_SSL") == "true"

	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		log.Printf("Gagal terhubung ke MinIO (attempt): %v", err)
		return nil, err
	}
	
	log.Println("Berhasil terhubung ke MinIO")

	ctx := context.Background()
	err = minioClient.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{})
	if err != nil {
		exists, errBucketExists := minioClient.BucketExists(ctx, bucketName)
		if errBucketExists == nil && exists {
			log.Printf("Bucket '%s' sudah ada, tidak perlu dibuat", bucketName)
		} else {
			log.Printf("Gagal membuat/memeriksa bucket: %v", err)
			return nil, err
		}
	} else {
		log.Printf("Berhasil membuat bucket '%s'", bucketName)
	}

	log.Printf("Menetapkan CORS policy untuk bucket: %s", bucketName)

	// Define the CORS rule
	// This allows PUT/GET/POST/DELETE from your frontend's origin
	corsRule := cors.Rule{
		AllowedOrigin: []string{"http://localhost:5173"},
		AllowedMethod: []string{"PUT", "GET", "POST", "DELETE"},
		AllowedHeader: []string{"Content-Type", "Authorization", "Origin"},
		ExposeHeader:  []string{"ETag"},
		MaxAgeSeconds:  3600,
	}

	config := cors.Config{
		CORSRules: []cors.Rule{corsRule}, 
	}

	err = minioClient.SetBucketCors(ctx, bucketName, &config)
	if err != nil {
		log.Printf("Failed to set CORS: %v", err)
		return nil, err
	}

	log.Println("Berhasil menetapkan CORS policy untuk bucket:", bucketName)
	// --- END OF ADDED SECTION ---

	return minioClient, nil
}