package initializers

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

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

    if endpoint == "" || accessKey == "" || secretKey == "" || region == "" {
        log.Fatal("Missing required S3 configuration. Please check ENDPOINT, ACCESS_KEY, SECRET_ACCESS_KEY, and REGION environment variables")
    }

    // Debug logging
    log.Printf("Using endpoint: %s", endpoint)
    log.Printf("Using region: %s", region)
    log.Printf("Access key length: %d", len(accessKey))
    log.Printf("Secret key length: %d", len(secretKey))

    endpoint = os.Getenv("ENDPOINT")
    log.Printf("Final formatted endpoint: %s", endpoint)

    s3Client, err := NewS3Client(endpoint, accessKey, secretKey, region)
    if err != nil {
        log.Fatalf("Failed to create S3 client: %v", err)
    }

    // Test connection
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    result, err := s3Client.ListBuckets(ctx, &s3.ListBucketsInput{})
    if err != nil {
        if strings.Contains(err.Error(), "SignatureDoesNotMatch") {
            log.Fatalf("Authentication failed. Please verify your credentials and region settings: %v", err)
        }
        log.Fatalf("Failed to list buckets: %v", err)
    }

    log.Println("Successfully connected to S3 storage")
    log.Println("Available buckets:")
    for _, bucket := range result.Buckets {
        log.Printf("- %s (Created: %s)", *bucket.Name, bucket.CreationDate.Format(time.RFC3339))
    }
}


func NewS3Client(endpoint, accessKey, secretKey, region string) (*s3.Client, error) {
    log.Printf("DEBUG: Received region: %s", region)
    log.Printf("DEBUG: Using endpoint: %s", endpoint)

    // Force "us-east-1" for signing
    customResolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
        return aws.Endpoint{
            URL:           endpoint,
            SigningRegion: "us-east-1", // Force correct signing region
            Source:        aws.EndpointSourceCustom,
        }, nil
    })

    cfg, err := config.LoadDefaultConfig(context.TODO(),
        config.WithRegion("us-east-1"), // Ensure region is set for signing
        config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
            accessKey,
            secretKey,
            "",
        )),
        config.WithEndpointResolverWithOptions(customResolver),
        config.WithDefaultsMode(aws.DefaultsModeInRegion),
    )

    if err != nil {
        return nil, fmt.Errorf("failed to load AWS config: %w", err)
    }

    // Force path-style addressing (important for Supabase S3 compatibility)
    s3Client := s3.NewFromConfig(cfg, func(o *s3.Options) {
        o.UsePathStyle = true
    })

    return s3Client, nil
}

