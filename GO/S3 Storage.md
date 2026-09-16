---
id: s3_storage_go
aliases:
  - AWS S3 in Go
  - S3 Upload and Download Go
tags:
  - go
  - aws
  - s3
  - cloud-storage
dg-publish: true
---

# AWS S3 in Go: Upload, Download & Presigned URLs

Amazon S3 (Simple Storage Service) is the industry standard for object storage. This guide uses **AWS SDK for Go v2** (`github.com/aws/aws-sdk-go-v2`), which is the modern, modular, and performant SDK.

```
                  +-----------------------------------+
                  |         Go Application            |
                  +-----------------------------------+
                       │              │           │
     PutObject/Uploader│      GetObject│          │PresignClient
                       ▼              ▼           ▼
               +-----------------------------------------+
               |                 AWS S3                  |
               |  Bucket: "my-app-assets"                |
               |  - uploads/report.pdf                   |
               |  - images/avatar.png                    |
               +-----------------------------------------+
```

---

## 1. Prerequisites & Installation

```bash
# Core config and S3 client
go get github.com/aws/aws-sdk-go-v2
go get github.com/aws/aws-sdk-go-v2/config
go get github.com/aws/aws-sdk-go-v2/service/s3
# High-level transfer manager for concurrent multipart uploads/downloads
go get github.com/aws/aws-sdk-go-v2/feature/s3/manager
```

---

## 2. Client Initialization

The SDK automatically detects credentials from environment variables (`AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`), shared credentials file (`~/.aws/credentials`), or IAM Roles on EC2/ECS/EKS.

```go
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type S3Service struct {
	client *s3.Client
	bucket string
}

func NewS3Service(ctx context.Context, region string, bucket string) (*S3Service, error) {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		return nil, fmt.Errorf("unable to load AWS SDK config: %w", err)
	}

	client := s3.NewFromConfig(cfg)
	return &S3Service{
		client: client,
		bucket: bucket,
	}, nil
}
```

---

## 3. Uploading Files

### Approach A: Standard `PutObject` (Best for small files < 100MB)
Reads from any `io.Reader`.

```go
import (
	"os"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func (s *S3Service) UploadFile(ctx context.Context, s3Key string, filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open local file: %w", err)
	}
	defer file.Close()

	_, err = s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(s3Key),
		Body:        file,
		ContentType: aws.String("application/octet-stream"),
	})
	if err != nil {
		return fmt.Errorf("failed to put object to S3: %w", err)
	}

	return nil
}
```

### Approach B: Concurrent Multipart Upload (Best for large files > 100MB)
Automatically splits large files into parts and uploads them in parallel.

```go
import "github.com/aws/aws-sdk-go-v2/feature/s3/manager"

func (s *S3Service) UploadLargeFile(ctx context.Context, s3Key string, filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	// Configure uploader with concurrency settings
	uploader := manager.NewUploader(s.client, func(u *manager.Uploader) {
		u.PartSize = 10 * 1024 * 1024 // 10MB per part
		u.Concurrency = 5             // 5 parallel part uploads
	})

	result, err := uploader.Upload(ctx, &s3.PutObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(s3Key),
		Body:   file,
	})
	if err != nil {
		return "", fmt.Errorf("multipart upload failed: %w", err)
	}

	return result.Location, nil
}
```

---

## 4. Downloading Files (Streaming `io.Reader`)

Streaming downloads directly to local files or HTTP response writers avoids loading entire gigabytes into RAM.

```go
import (
	"io"
	"os"
)

func (s *S3Service) DownloadFile(ctx context.Context, s3Key string, destinationPath string) error {
	output, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(s3Key),
	})
	if err != nil {
		return fmt.Errorf("failed to get object from S3: %w", err)
	}
	defer output.Body.Close()

	outFile, err := os.Create(destinationPath)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer outFile.Close()

	// Stream directly to destination file with O(1) memory
	_, err = io.Copy(outFile, output.Body)
	if err != nil {
		return fmt.Errorf("failed during download copy: %w", err)
	}

	return nil
}
```

---

## 5. Generating Presigned URLs

Presigned URLs allow users/browsers to safely upload or download S3 objects directly without routing traffic through your Go backend servers.

```go
import "time"

// PresignedDownloadURL generates a temporary read link valid for duration.
func (s *S3Service) PresignedDownloadURL(ctx context.Context, s3Key string, lifetime time.Duration) (string, error) {
	presignClient := s3.NewPresignClient(s.client)

	req, err := presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(s3Key),
	}, s3.WithPresignExpires(lifetime))
	if err != nil {
		return "", fmt.Errorf("failed to presign get: %w", err)
	}

	return req.URL, nil
}

// PresignedUploadURL generates a temporary link allowing direct client PUT upload.
func (s *S3Service) PresignedUploadURL(ctx context.Context, s3Key string, lifetime time.Duration) (string, error) {
	presignClient := s3.NewPresignClient(s.client)

	req, err := presignClient.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(s3Key),
	}, s3.WithPresignExpires(lifetime))
	if err != nil {
		return "", fmt.Errorf("failed to presign put: %w", err)
	}

	return req.URL, nil
}
```
