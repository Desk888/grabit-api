package initializers

import (
	"context"
	"fmt"
	"log"
	"net/url"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// Initialize S3 connection - function called by main.go
func InitS3Conn() {
	endpoint := "https://xyz.supabase.co/storage/v1" // Replace with your S3-compatible storage URL
	accessKey := "your-access-key"
	secretKey := "your-secret-key"
	region := "us-east-1"

	// Call NewS3Client to initialize a new S3 client
	s3Client, err := NewS3Client(endpoint, accessKey, secretKey, region)
	if err != nil {
		log.Fatalf("Failed to create S3 client: %v", err)
	}

	// Error loading buckets
	result, err := s3Client.ListBuckets(context.TODO(), &s3.ListBucketsInput{})
	if err != nil {
		log.Fatalf("Failed to list buckets: %v", err)
	}

	// List all available buckets
	fmt.Println("Buckets:")
	for _, bucket := range result.Buckets {
		fmt.Println(*bucket.Name)
	}
}	

// Initialize a new S3 client with any S3 compatible storage
func NewS3Client(endpoint, accessKey, secretKey, region string) (*s3.Client, error) {
	// Parse endpoint URL
	parsedURL, err := url.Parse(endpoint)
	if err != nil {
		return nil, fmt.Errorf("invalid endpoint URL: %w", err)
	}

	// Load AWS config manually since we are using a custom S3-compatible endpoint
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKey, secretKey, "")),
		config.WithEndpointResolverWithOptions(aws.EndpointResolverWithOptionsFunc(
			func(service, region string, options ...interface{}) (aws.Endpoint, error) {
				return aws.Endpoint{
					URL:           parsedURL.String(),
					SigningRegion: region,
				}, nil
			},
		)),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	return s3.NewFromConfig(cfg), nil
}