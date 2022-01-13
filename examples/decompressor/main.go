package main

import (
	"compress/gzip"
	"compress/zlib"
	"flag"
	"io"
	"log"
	"os"

	"github.com/cheggaaa/pb/v3"
	"github.com/golang/snappy"
	"github.com/pierrec/lz4/v4"
)

var inputFileFlag = flag.String("input", "", "file to be decompressed")
var outputFileFlag = flag.String("output", "output.memr", "file to which decompressed image should be written")
var compressionFlag = flag.String("compression", "snappy", "type of compression to use for output (snappy, lz4, zlib, gzip)")

func main() {

	flag.Parse()

	inputFile := *inputFileFlag
	outputFile := *outputFileFlag
	compression := *compressionFlag
	if inputFile == "" {
		log.Fatalf("input file must be specified")
	}

	var decompressor func(io.Reader) (io.Reader, error)
	switch compression {
	case "snappy":
		decompressor = func(r io.Reader) (io.Reader, error) {
			return snappy.NewReader(r), nil
		}
	case "lz4":
		decompressor = func(r io.Reader) (io.Reader, error) {
			return lz4.NewReader(r), nil
		}
	case "zlib", "gzip": // slow compressors, do not actually use these
		log.Printf("[WARN] %s compression is very slow and for demo purposes only. it's advisable to use snappy compression instead", compression)
		switch compression {
		case "zlib":
			decompressor = func(r io.Reader) (io.Reader, error) {
				return zlib.NewReader(r)
			}
		case "gzip":
			decompressor = func(r io.Reader) (io.Reader, error) {
				return gzip.NewReader(r)
			}
		}
	default:
		log.Fatalf("invalid compression type specified (%s); must be one of: snappy, lz4, zlib, gzip", compression)
	}

	input, err := os.Open(inputFile)
	if err != nil {
		log.Fatalf("failed to open input: %s", err)
	}
	defer input.Close()

	// Open the outpu file for writing
	output, err := os.Create(outputFile)
	if err != nil {
		log.Fatalf("failed to open output file (%s) for writing %s", outputFile, err)
	}
	defer output.Close()

	fi, err := input.Stat()
	if err != nil {
		log.Fatalf("could not get file info: %s", err)
	}

	bar := pb.Start64(fi.Size())

	var reader io.Reader = bar.NewProxyReader(input)

	reader, err = decompressor(reader)
	if err != nil {
		log.Fatalf("failed to get reader for decompressor (%s): %s", compression, err)
	}

	// Use io.Copy to copy the memory to a file with the desired decompressor
	_, err = io.Copy(output, reader)
	if err != nil {
		log.Fatalf("failed to copy memory to local file %q: %s", outputFile, err)
	}
}
