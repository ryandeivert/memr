package main

import (
	"flag"
	"io"
	"log"
	"os"

	"github.com/ryandeivert/memr"
)

var outputFileFlag = flag.String("output", "output.memr", "file to which memory should be copied")
var withProgressFlag = flag.Bool("progress", true, "whether or not progress should be output during read")

func main() {

	flag.Parse()

	outputFile := *outputFileFlag

	options := func(m *memr.Reader) {
		m.WithProgress = *withProgressFlag
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

	// Use io.Copy to copy the memory to a file
	_, err = io.Copy(writer, reader)
	if err != nil {
		log.Fatalf("failed to copy memory to local file %q: %s", outputFile, err)
	}
}
