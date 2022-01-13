package main

import (
	"encoding/binary"
	"flag"
	"io"
	"log"
	"os"

	"github.com/ryandeivert/memr"
)

var outputFileFlag = flag.String("output", "output.memr", "file to which memory should be copied")
var headerMagicFlag = flag.String("magic", "MEMR", "32-bit magic string (q4 chars in length) be used in headers")
var rawFlag = flag.Bool("raw", false, "set to true if headers should be omitted from output (aka \"raw\" image)")

func main() {

	flag.Parse()

	outputFile := *outputFileFlag
	headerMagic := *headerMagicFlag
	raw := *rawFlag
	if !raw && len(headerMagic) != 4 {
		log.Fatalf("header magic must 4 characters: %s", headerMagic)
	}

	var headerFunc memr.PageHeaderProviderFunc
	if !raw {
		// Utilize a custom header function to control how headers
		// are formatted. The result is an interface type, allowing
		// for arbitrary formats
		headerFunc = func(start, end uint64) interface{} {
			return &memr.DefaultHeader{
				// custom magic defined here will override "EMiL" default
				Magic:     binary.LittleEndian.Uint32([]byte(headerMagic)),
				Version:   1,
				StartAddr: start,
				EndAddr:   end - 1,
			}
		}
	}

	options := func(m *memr.Reader) {
		m.PageHeaderProvider = headerFunc // this can be nil, which omits headers from output (raw image)
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
