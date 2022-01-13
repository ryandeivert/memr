package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/ryandeivert/memr"
)

var outputFileFlag = flag.String("output", "output.memr", "file to which memory should be copied")
var sourceFlag = flag.String("source", "", fmt.Sprintf("source from which memory should be read. should be one of: %s, %s, %s", memr.SourceKcore, memr.SourceCrash, memr.SourceMem))

func main() {

	flag.Parse()

	outputFile := *outputFileFlag
	source := memr.MemSource(*sourceFlag)

	switch source {
	case "":
		log.Fatal("source flag must be specified")
	case memr.SourceKcore, memr.SourceCrash, memr.SourceMem:
	default:
		log.Fatalf("unknown source specified: %s", source)
	}

	// Use memr.NewReader(source) to open a reader for this source
	reader, err := memr.NewReader(source)
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
