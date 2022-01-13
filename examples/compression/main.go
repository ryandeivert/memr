package main

import (
	"compress/gzip"
	"compress/zlib"
	"flag"
	"io"
	"log"
	"os"

	"github.com/golang/snappy"
	"github.com/pierrec/lz4/v4"
	"github.com/ryandeivert/memr"
)

var outputFileFlag = flag.String("output", "output.memr", "file to which memory should be copied")
var compressionFlag = flag.String("compression", "snappy", "type of compression to use for output (snappy, lz4, zlib, gzip, none)")

func main() {

	flag.Parse()

	outputFile := *outputFileFlag
	compression := *compressionFlag

	var compressor func(io.Writer) io.Writer
	switch compression {
	case "snappy":
		compressor = func(w io.Writer) io.Writer {
			return snappy.NewBufferedWriter(w)
		}
	case "lz4":
		compressor = func(w io.Writer) io.Writer {
			return lz4.NewWriter(w)
		}
	case "zlib", "gzip": // slow compressors, do not actually use these
		log.Printf("[WARN] %s compression is very slow and for demo purposes only. it's advisable to use snappy compression instead", compression)
		switch compression {
		case "zlib":
			compressor = func(w io.Writer) io.Writer {
				return zlib.NewWriter(w)
			}
		case "gzip":
			compressor = func(w io.Writer) io.Writer {
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

	// Open the local file for writing
	file, err := os.Create(outputFile)
	if err != nil {
		log.Fatalf("failed to open local file %q for writing: %s", outputFile, err)
	}
	defer file.Close()

	var writer io.Writer = file
	if compressor != nil { // could be nil if set to "none"
		writer = compressor(file)
	}

	// Use io.Copy to copy the memory to a file with the desired compressor (if not disabled)
	_, err = io.Copy(writer, reader)
	if err != nil {
		log.Fatalf("failed to copy memory to local file %q: %s", outputFile, err)
	}
}
