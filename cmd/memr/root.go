package main

/*

This CLI demonstrates using memr both for writing to a file and copying
to S3 using manager.Uploader. However, this could be extended to do
things like copying to os.Stdout, using io.Copy(os.Stdout, reader), or
uploading to other cloud providers.

*/

import (
	"fmt"
	"io"
	"log"
	"os"

	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/golang/snappy"
	"github.com/ryandeivert/memr"
	"github.com/spf13/cobra"
)

var (
	version               = "development"
	concurrency           = manager.DefaultUploadConcurrency
	compress              = true
	progress              = true
	allDevices            = []string{"/proc/kcore", "/dev/crash", "/dev/mem"}
	region                = "us-east-1"
	s3Bucket, s3ObjectKey string
	localFile             string
)

// rootCmd is the entry point command for the CLI
var rootCmd = &cobra.Command{
	Use:          "memr",
	Short:        "Linux memory reader",
	SilenceUsage: true,
	Long:         "A portable user-land volatile memory reader for Linux written in golang",
	Example: `
Writing to local file:
memr --local-file <FILE>

Streaming directly to S3 bucket:
memr --bucket <BUCKET> --key <KEY>

Targeting a specific device:
memr /dev/mem --local-file <FILE>

Skipping compression:
memr --compress=false  --local-file <FILE>`,
	ValidArgs: allDevices,
	Args:      cobra.OnlyValidArgs,
	RunE: func(cmd *cobra.Command, devices []string) (err error) {

		options := func(m *memr.Reader) {
			m.WithProgress = progress
		}

		var reader *memr.Reader
		if len(devices) == 0 {
			reader, err = memr.Probe(options)
		} else {
			for _, t := range devices {
				reader, err = memr.NewReader(memr.MemSource(t), options)
				if err == nil {
					break
				}
			}
		}
		if err != nil {
			return fmt.Errorf("failed to load memory reader: %s", err)
		}
		defer reader.Close()

		// Using local file
		if localFile != "" {
			var writer io.WriteCloser
			writer, err = os.Create(localFile)
			if err != nil {
				return fmt.Errorf("failed to open local file for writing %s", err)
			}
			defer writer.Close()

			if compress {
				writer = snappy.NewBufferedWriter(writer)
			}

			read, err := io.Copy(writer, reader)
			if err != nil {
				return fmt.Errorf("failed to copy memory to local file %s", err)
			}

			reader.Close()

			if reader.Size() != uint64(read) {
				return fmt.Errorf("failed to read all data. expected=%d; read=%d ", reader.Size(), read)
			}

			log.Printf("acquired memory using %q to file: %s", reader.Source(), localFile)
		} else {

			// Not using a local file, so assume s3
			res, err := S3Writer(reader, compress, region, s3Bucket, s3ObjectKey, concurrency)
			if err != nil {
				return err
			}

			log.Printf("acquired memory using %q and uploaded to S3: %s", reader.Source(), res.Location)
		}

		return nil
	},
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		flags := cmd.Flags()
		verbosity, err := flags.GetCount("verbose")
		if err != nil {
			return err
		}
		memr.SetLogLevel(memr.LogLvl(verbosity))

		if localFile+s3Bucket+s3ObjectKey == "" {
			return fmt.Errorf("either \"--local-file\" flag, or \"--bucket\" and \"--key\" flags, must be supplied")
		}
		if localFile != "" {
			return nil
		}
		if s3Bucket != "" && s3ObjectKey == "" {
			return fmt.Errorf("\"--key\" flag must be supplied")
		}
		if s3ObjectKey != "" && s3Bucket == "" {
			return fmt.Errorf("\"--bucket\" flag must be supplied")
		}
		return nil
	},
}

func init() {
	// Global (persistent) flags
	_ = rootCmd.PersistentFlags().CountP("verbose", "v", "enable verbose logging")

	rootCmd.PersistentFlags().BoolVarP(&compress, "compress", "c", true, "compress the output with snappy")
	rootCmd.PersistentFlags().BoolVarP(&progress, "progress", "p", true, "show progress")
	rootCmd.PersistentFlags().IntVarP(&concurrency, "concurrency", "t", concurrency, "number of threads to use for S3 upload")
	rootCmd.PersistentFlags().StringVarP(&region, "region", "r", region, "AWS region to use with S3 client")
	rootCmd.PersistentFlags().StringVarP(&s3Bucket, "bucket", "b", s3Bucket, "S3 bucket to which output should be sent")
	rootCmd.PersistentFlags().StringVarP(&s3ObjectKey, "key", "k", s3ObjectKey, "key to use for uploading to S3 bucket")
	rootCmd.PersistentFlags().StringVarP(&localFile, "local-file", "f", localFile, "local file to write to, instead of S3")
}

func main() {
	rootCmd.Version = version
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
