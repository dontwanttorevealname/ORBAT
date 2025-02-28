package storage

import (
	"cloud.google.com/go/storage"
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"time"
)

// Client is the global storage client
var Client *storage.Client
var BucketName string
var environment string

// Initialize sets up the storage client
func Initialize() error {
	var err error
	ctx := context.Background()
	
	Client, err = storage.NewClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to create storage client: %v", err)
	}

	BucketName = os.Getenv("GCS_BUCKET_NAME")
	if BucketName == "" {
		return fmt.Errorf("GCS_BUCKET_NAME environment variable not set")
	}

	// Store the environment setting
	environment = os.Getenv("ENV")
	if environment == "" {
		environment = "development"
	}
	
	return nil
}

// Close closes the storage client
func Close() {
	if Client != nil {
		Client.Close()
	}
}

// UploadImage uploads an image to Google Cloud Storage
func UploadImage(file io.Reader, filename string) (string, error) {
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, time.Minute*10)
	defer cancel()

	bucket := Client.Bucket(BucketName)
	obj := bucket.Object(filename)

	writer := obj.NewWriter(ctx)
	if _, err := io.Copy(writer, file); err != nil {
		return "", err
	}
	if err := writer.Close(); err != nil {
		return "", err
	}

	// Make the object public
	if err := obj.ACL().Set(ctx, storage.AllUsers, storage.RoleReader); err != nil {
		return "", err
	}

	return fmt.Sprintf("https://storage.googleapis.com/%s/%s", BucketName, filename), nil
}

// DeleteImage deletes an image from Google Cloud Storage
func DeleteImage(imageURL string) error {
	// Don't actually delete files when in test environment
	if environment == "test" {
		fmt.Printf("Test environment: Skipping deletion of image %s\n", imageURL)
		return nil
	}

	// Extract bucket name and object path from the URL
	// URL format: https://storage.googleapis.com/BUCKET_NAME/PATH/TO/OBJECT
	urlParts := strings.Split(imageURL, "/")
	if len(urlParts) < 4 {
		return fmt.Errorf("invalid GCS URL format")
	}
	
	objectPath := strings.Join(urlParts[4:], "/")

	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()

	// Delete the object from GCS
	err := Client.Bucket(BucketName).Object(objectPath).Delete(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete image from storage: %v", err)
	}
	
	return nil
} 