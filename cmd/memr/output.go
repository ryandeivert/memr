package main

import (
	"fmt"
	"io"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/golang/snappy"
)

func S3Writer(reader io.ReadCloser, compress bool, region, bucket, key string, concurrency int) (*s3manager.UploadOutput, error) {

	sess := session.Must(session.NewSessionWithOptions(session.Options{
		Config: aws.Config{Region: aws.String(region)},
	}))

	// Create an uploader with the session and custom options
	uploader := s3manager.NewUploader(sess, func(u *s3manager.Uploader) {
		u.PartSize = s3manager.MinUploadPartSize // 5MB (smallest allowed)
		// 5 goroutines is the default. A lower value here will have a lesser
		// impact on memory pressure, and should be considered
		// Increasing this value will directly impact how much memory we use
		// set to s3manager.DefaultUploadConcurrency by default
		u.Concurrency = concurrency

		// A buffer provider could be used to allow for larger buffers (64 KiB?) in memory
		// u.BufferProvider = s3manager.NewBufferedReadSeekerWriteToPool(64 * 1024)
	})

	var s3Reader io.Reader = reader
	if compress {
		var wPipe *io.PipeWriter
		s3Reader, wPipe = io.Pipe()

		go func() {
			writer := snappy.NewBufferedWriter(wPipe)
			defer func() {
				cErr := reader.Close()
				if cErr != nil {
					log.Fatalf("failed to close reader: %s", cErr)
				}
				cErr = writer.Close()
				if cErr != nil {
					log.Fatalf("failed to close writer: %T; %s", writer, cErr)
				}
				cErr = wPipe.Close()
				if cErr != nil {
					log.Fatalf("failed to close pipe writer: %s", cErr)
				}
			}()

			if _, err := io.Copy(writer, reader); err != nil {
				log.Fatalf("compressor failed: %s", err)
			}
		}()
	} else {
		defer reader.Close()
	}

	// Upload the file to S3
	result, err := uploader.Upload(&s3manager.UploadInput{
		ACL:    aws.String(s3.ObjectCannedACLBucketOwnerFullControl),
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		Body:   s3Reader,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to upload to s3: %s", err)
	}

	return result, nil
}
