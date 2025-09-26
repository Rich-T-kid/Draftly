package services

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// S3Service handles document storage and retrieval from S3
type S3Service struct {
	client *s3.Client
	bucket string
}

// NewS3Service creates a new S3 service instance
func NewS3Service() (*S3Service, error) {
	bucket := os.Getenv("S3_BUCKET_NAME")
	if bucket == "" {
		return nil, fmt.Errorf("S3_BUCKET_NAME environment variable is required")
	}

	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(os.Getenv("AWS_REGION")),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	client := s3.NewFromConfig(cfg)

	return &S3Service{
		client: client,
		bucket: bucket,
	}, nil
}

// UploadDocument uploads document content to S3
func (s *S3Service) UploadDocument(documentID int, content []byte) (string, error) {
	key := s.generateKey(documentID)

	_, err := s.client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(content),
		ContentType: aws.String("text/plain"),
	})

	if err != nil {
		return "", fmt.Errorf("failed to upload document to S3: %w", err)
	}

	return key, nil
}

// DownloadDocument retrieves document content from S3
func (s *S3Service) DownloadDocument(key string) ([]byte, error) {
	result, err := s.client.GetObject(context.TODO(), &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})

	if err != nil {
		return nil, fmt.Errorf("failed to download document from S3: %w", err)
	}
	defer result.Body.Close()

	content, err := io.ReadAll(result.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read document content: %w", err)
	}

	return content, nil
}

// UpdateDocument updates existing document content in S3
func (s *S3Service) UpdateDocument(key string, content []byte) error {
	_, err := s.client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(content),
		ContentType: aws.String("text/plain"),
	})

	if err != nil {
		return fmt.Errorf("failed to update document in S3: %w", err)
	}

	return nil
}

// DeleteDocument removes document from S3
func (s *S3Service) DeleteDocument(key string) error {
	_, err := s.client.DeleteObject(context.TODO(), &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})

	if err != nil {
		return fmt.Errorf("failed to delete document from S3: %w", err)
	}

	return nil
}

// DocumentExists checks if a document exists in S3
func (s *S3Service) DocumentExists(key string) (bool, error) {
	_, err := s.client.HeadObject(context.TODO(), &s3.HeadObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})

	if err != nil {
		if strings.Contains(err.Error(), "NotFound") {
			return false, nil
		}
		return false, fmt.Errorf("failed to check document existence: %w", err)
	}

	return true, nil
}

// GetDocumentContent retrieves document content directly
func (s *S3Service) GetDocumentContent(documentID int) ([]byte, error) {
	key := s.generateKey(documentID)
	return s.DownloadDocument(key)
}

// generateKey creates a consistent S3 key for a document
func (s *S3Service) generateKey(documentID int) string {
	return fmt.Sprintf("documents/%d.txt", documentID)
}
