package initializers

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func InitS3Conn() {
	endpoint := os.Getenv("ENDPOINT")
	accessKey := os.Getenv("ACCESS_KEY")
	secretKey := os.Getenv("SECRET_ACCESS_KEY")
	region := os.Getenv("REGION")

	// Ensure endpoint has proper format
	if !strings.HasSuffix(endpoint, "/storage/v1") {
		endpoint = strings.TrimSuffix(endpoint, "/") + "/storage/v1"
	}

	s3Client, err := NewS3Client(endpoint, accessKey, secretKey, region)
	if err != nil {
		log.Fatalf("Failed to create S3 client: %v", err)
	}

	// List buckets
	result, err := s3Client.ListBuckets(context.TODO(), &s3.ListBucketsInput{})
	if err != nil {
		log.Fatalf("Failed to list buckets: %v", err)
	}

	fmt.Println("Buckets:")
	for _, bucket := range result.Buckets {
		fmt.Println(*bucket.Name)
	}
}

func NewS3Client(endpoint, accessKey, secretKey, region string) (*s3.Client, error) {
	// Validate and parse the endpoint URL
	parsedURL, err := url.Parse(endpoint)
	if err != nil {
		return nil, fmt.Errorf("invalid endpoint URL: %w", err)
	}

	// Create the AWS configuration
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			accessKey,
			secretKey,
			"",
		)),
		config.WithEndpointResolverWithOptions(
			aws.EndpointResolverWithOptionsFunc(
				func(service, region string, options ...interface{}) (aws.Endpoint, error) {
					return aws.Endpoint{
						URL:               parsedURL.String(),
						SigningRegion:     region,
						HostnameImmutable: true,
						Source:            aws.EndpointSourceCustom,
					}, nil
				},
			),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	// Create and return the S3 client with Supabase-specific options
	s3Client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.UsePathStyle = true
		// No EndpointOptions available in the current SDK version
	})

	return s3Client, nil
}