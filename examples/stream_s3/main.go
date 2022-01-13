package main

import (
	"compress/gzip"
	"compress/zlib"
	"flag"
	"fmt"
	"io"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/golang/snappy"
	"github.com/pierrec/lz4/v4"
	"github.com/ryandeivert/memr"
)

type compressorFunc func(io.Writer) io.WriteCloser

var bucketFlag = flag.String("bucket", "", "S3 bucket to which memory should be streamed")
var keyFlag = flag.String("key", "", "key for object in S3")
var compressionFlag = flag.String("compression", "snappy", "type of compression to use for output (snappy, lz4, zlib, gzip, none)")
var concurrencyFlag = flag.Int("threads", s3manager.DefaultUploadConcurrency, "number of goroutines to use for upload")

func main() {

	flag.Parse()

	bucket := *bucketFlag
	key := *keyFlag
	compression := *compressionFlag

	if bucket == "" {
		log.Fatal("bucket flag is required")
	} else if key == "" {
		log.Fatal("key flag is required")
	}

	var compressor compressorFunc
	switch compression {
	case "snappy":
		compressor = func(w io.Writer) io.WriteCloser {
			return snappy.NewBufferedWriter(w)
		}
	case "lz4":
		compressor = func(w io.Writer) io.WriteCloser {
			return lz4.NewWriter(w)
		}
	case "zlib", "gzip": // slow compressors, do not actually use these
		log.Printf("[WARN] %s compression is very slow and for demo purposes only. it's advisable to use snappy compression instead", compression)
		switch compression {
		case "zlib":
			compressor = func(w io.Writer) io.WriteCloser {
				return zlib.NewWriter(w)
			}
		case "gzip":
			compressor = func(w io.Writer) io.WriteCloser {
				return gzip.NewWriter(w)
			}
		}
	case "none":
		// explicitly disabling compression
	default:
		log.Fatalf("invalid compression type specified (%s); must be one of: snappy, lz4, zlib, gzip", compression)
	}

	// Use memr.Probe() to enumerate all available types
	reader, err := memr.Probe()
	if err != nil {
		log.Fatalf("failed to load memory reader: %s", err)
	}
	defer reader.Close()

	res, err := copyToS3(reader, compressor, bucket, key)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Uploaded object to S3: %s\n", res.Location)
}

// Copy the data to S3 through the open reader, using optional compression
func copyToS3(reader io.ReadCloser, compressor compressorFunc, bucket, key string) (*s3manager.UploadOutput, error) {
	sess := session.Must(session.NewSession())
	// Create an uploader with the session and custom options
	uploader := s3manager.NewUploader(sess, func(u *s3manager.Uploader) {
		// PartSize: 5MB (smallest allowed)
		u.PartSize = s3manager.MinUploadPartSize

		// Concurrency: 5 goroutines is the default. A lower value here will
		// have a lesser impact on memory pressure, and should be considered
		// Increasing this value will directly impact how much memory we use
		// set to s3manager.DefaultUploadConcurrency by default
		u.Concurrency = *concurrencyFlag

		// BufferProvider: this could be used to allow for larger buffers (64 KiB?)
		// in memory but may not be desirable if memory pressure is of concern
		// u.BufferProvider = s3manager.NewBufferedReadSeekerWriteToPool(64 * 1024)
	})

	var s3Reader io.Reader = reader
	if compressor != nil {
		var wPipe *io.PipeWriter
		s3Reader, wPipe = io.Pipe()

		go func() {
			writer := compressor(wPipe)
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
