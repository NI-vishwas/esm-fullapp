package storage

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/aws-sdk-go-v2/aws"
)

type S3Storage struct {
	client     *s3.Client
	bucketName string
	region     string
}

func NewS3Storage() (*S3Storage, error) {
	bucket := os.Getenv("AWS_S3_BUCKET")
	if bucket == "" {
		bucket = "ems-tickets-bucket" // default fallback
	}

	region := os.Getenv("AWS_REGION")
	if region == "" {
		region = "us-east-1"
	}

	// Read optional custom endpoint (highly useful for MinIO / LocalStack during local development)
	customEndpoint := os.Getenv("AWS_S3_ENDPOINT")

	var cfg aws.Config
	var err error

	if customEndpoint != "" {
		// Custom endpoint configuration for local development (LocalStack/MinIO)
		cfg, err = config.LoadDefaultConfig(context.TODO(),
			config.WithRegion(region),
			config.WithEndpointResolverWithOptions(aws.EndpointResolverWithOptionsFunc(
				func(service, reg string, options ...interface{}) (aws.Endpoint, error) {
					return aws.Endpoint{
						URL:               customEndpoint,
						SigningRegion:     region,
						HostnameImmutable: true, // Crucial for MinIO and local DNS resolution
					}, nil
				},
			)),
		)
	} else {
		// Standard production AWS config
		cfg, err = config.LoadDefaultConfig(context.TODO(), config.WithRegion(region))
	}

	if err != nil {
		return nil, fmt.Errorf("unable to load SDK config: %w", err)
	}

	client := s3.NewFromConfig(cfg)

	return &S3Storage{
		client:     client,
		bucketName: bucket,
		region:     region,
	}, nil
}

// UploadTicketPDF uploads the raw bytes of a generated PDF ticket to S3 and returns its public URL
func (s *S3Storage) UploadTicketPDF(ctx context.Context, bookingID string, pdfBytes []byte) (string, error) {
	fileKey := fmt.Sprintf("tickets/booking-%s-%d.pdf", bookingID, time.Now().Unix())

	// Detect content type (should be application/pdf, but fallback dynamically)
	contentType := http.DetectContentType(pdfBytes)
	if contentType == "text/plain; charset=utf-8" {
		contentType = "application/pdf"
	}

	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucketName),
		Key:         aws.String(fileKey),
		Body:        bytes.NewReader(pdfBytes),
		ContentType: aws.String(contentType),
		// Allows public read access so clients can download it via URL
		ACL:         types.ObjectCannedACLPublicRead, 
	})

	if err != nil {
		return "", fmt.Errorf("failed to upload object to S3: %w", err)
	}

	// Build public resource access URL
	var fileURL string
	customEndpoint := os.Getenv("AWS_S3_ENDPOINT")
	if customEndpoint != "" {
		// e.g. http://localhost:4566/ems-tickets-bucket/tickets/booking-123.pdf
		fileURL = fmt.Sprintf("%s/%s/%s", customEndpoint, s.bucketName, fileKey)
	} else {
		// Standard S3 public URL
		fileURL = fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", s.bucketName, s.region, fileKey)
	}

	return fileURL, nil
}