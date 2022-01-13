package main

import (
	"encoding/binary"
	"flag"
	"io"
	"log"
	"os"

	"github.com/golang/snappy"
	"github.com/ryandeivert/memr"
)

var outputFileFlag = flag.String("output", "output.memr", "file to which memory should be copied")

func main() {

	flag.Parse()

	outputFile := *outputFileFlag

	options := func(m *memr.Reader) {
		// A custom header should be used when page-level compression is performed
		// This matches what AVML defines as their header for "version 2" using snappy
		m.PageHeaderProvider = func(start, end uint64) interface{} {
			return &memr.DefaultHeader{
				Magic:     binary.LittleEndian.Uint32([]byte("AVML")), // custom "AVML" magic
				Version:   2,
				StartAddr: start,
				EndAddr:   end - 1,
			}
		}
		m.PageHandler = func(w io.Writer) io.WriteCloser {
			return snappy.NewBufferedWriter(w)
		}
	}

	// Use memr.Probe() to enumerate all available types,
	// passing in options for the resulting reader
	reader, err := memr.Probe(options)
	if err != nil {
		log.Fatalf("failed to load memory reader: %s", err)
	}
	defer reader.Close()

	// Open the local file for writing
	writer, err := os.Create(outputFile)
	if err != nil {
		log.Fatalf("failed to open local file %q for writing: %s", outputFile, err)
	}
	defer writer.Close()

	// Use io.Copy to copy the memory to a file with the desired compressor (if not disabled)
	_, err = io.Copy(writer, reader)
	if err != nil {
		log.Fatalf("failed to copy memory to local file %q: %s", outputFile, err)
	}
}
