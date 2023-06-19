package memr

import (
	"debug/elf"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/cheggaaa/pb/v3"
	"github.com/ryandeivert/memr/internal/iomem"
)

// Reader implements the io.ReadCloser interface
// Its values can also used as "options" for the resulting Reader
// returned by the Probe() and NewReader() functions.
type Reader struct {
	// PageHeaderProvider is a function that returns a custom header structure to be applied
	// prepended before each memory "page" that is read
	PageHeaderProvider PageHeaderProviderFunc

	// PageHandler allows for page-level compression, or other custom handling of pages
	// Note: this should really not be used, and exists only for compatibility with other
	// tools. In most cases, the returned io.Reader should be used for compression instead
	PageHandler PageWriterFunc

	// WithProgress should be set to false if progress should not be reported during
	// the reading of memory. The default when calling NewReader is true.
	WithProgress bool

	// ByteOrder should be either binary.LittleEndian or binary.BigEndian.
	// This determines the byte order applied to pager headers. The default
	// when calling NewReader is binary.LittleEndian.
	ByteOrder binary.ByteOrder

	// unexported items
	source    MemSource
	memRanges iomem.MemRanges
	input     io.Closer
	reader    io.Reader
	size      uint64
	bar       *pb.ProgressBar
}

// Source returns the MemSource for this reader (one of: /proc/kcore, /dev/crash, /dev/mem)
func (r *Reader) Source() MemSource {
	return r.source
}

// Close satisfies the io.Closer interface.
// This should be called after all reading is complete to close the underlying
// input file. It also signals the progress bar to flush its output. Without calling
// this, you will likely see incorrect progress output upon completion.
func (r *Reader) Close() error {
	r.bar.Finish()
	return r.input.Close()
}

// Read satisfies the io.Reader interface.
func (r *Reader) Read(p []byte) (int, error) {
	if r.WithProgress && !r.bar.IsStarted() {
		r.bar.Start()
	}
	return r.reader.Read(p)
}

// PageWriterFunc can be used to add special handling of page contents,
// such as page-level compression. This applies to a Reader in-line as
// data is read. Note: use of this should nearly always be avoided.
type PageWriterFunc func(io.Writer) io.WriteCloser

// Probe enumerates all available memory sources and returns the first valid
// reader. Current sources are: /proc/kcore, /dev/crash, /dev/mem.
// Optional options can be provided for the resulting Reader.
// See NewReader for usage of custom options.
func Probe(options ...func(*Reader)) (*Reader, error) {

	memRanges, err := iomem.ReadRanges()
	if err != nil {
		return nil, err
	}
	options = append(options, func(r *Reader) {
		r.memRanges = memRanges
	})
	for _, source := range allMemSources() {
		reader, err := NewReader(source, options...)
		if err != nil {
			log.Printf("[DEBUG] failed to open reader for %s: %v", source, err)
			continue
		}
		return reader, nil
	}

	return nil, fmt.Errorf("failed to open reader for any memory device")
}

// NewReader tries to open the specified memory source for reading.
// Source should be one of: SourceKcore (/proc/kcore), SourceCrash (/dev/crash),
// or SourceMem (/dev/mem). Optional options can be provided for the resulting Reader.
//
// Example:
//
//	// Omit headers from the resulting image (raw) and suppress progress
//	reader, err := memr.NewReader(memr.SourceKcore, func(m *memr.Reader) {
//		m.WithProgress = false
//		m.PageHeaderProvider = nil
//	})
func NewReader(source MemSource, options ...func(*Reader)) (reader *Reader, err error) {

	reader = &Reader{
		PageHeaderProvider: HeaderLime,
		WithProgress:       true,
		ByteOrder:          binary.LittleEndian,
		source:             source,
	}

	for _, option := range options {
		option(reader)
	}

	err = reader.Reset()

	log.Printf("[DEBUG] loaded reader (valid=%t): %+v", err == nil, *reader)

	return
}

func (r *Reader) Reset() (err error) {

	r.input = nil
	r.reader = nil
	r.size = 0
	r.bar = new(pb.ProgressBar)

	// Retain any cached memRanges, these are unlikely to have changed
	if r.memRanges == nil {
		r.memRanges, err = iomem.ReadRanges()
		if err != nil {
			return
		}
	}

	log.Printf("[DEBUG] initializing reader for %s", r.source)

	var blks blocks
	if r.source.isPhysical() {
		file, err := os.Open(string(r.source))
		if err != nil {
			return err
		}
		r.input = file
		blks = physicalBlocks(file, r.memRanges, r.source.forcePageReads())
	} else {
		if err := verifySource(string(r.source)); err != nil {
			return err
		}
		file, err := elf.Open(string(r.source))
		if err != nil {
			return err
		}
		r.input = file
		blks = kcoreBlocks(file, r.memRanges)
	}

	log.Printf("[DEBUG] loaded blocks:\n%s", blks)

	if len(blks) != len(r.memRanges) {
		return fmt.Errorf("unable to load necessary reader(s) for %s", r.source)
	}

	r.reader, r.size = r.initBlockReaders(blks)

	// We now know the expected total size to be read, so set it
	r.bar.SetTotal(int64(r.size))

	return
}

// Size returns the expected size of the memory to be read by the reader
func (r *Reader) Size() uint64 {
	return r.size
}

func applyPageWriter(r io.Reader, handlerFunc PageWriterFunc) io.Reader {
	if handlerFunc == nil {
		log.Print("[DEBUG] no in-line writer function specified, skipping")
		return r
	}

	rPipe, wPipe := io.Pipe()
	writer := handlerFunc(wPipe)

	go func() {
		defer func() {
			cErr := writer.Close()
			if cErr != nil {
				log.Fatalf("failed to close writer: %T; %s", writer, cErr)
			}
			cErr = wPipe.Close()
			if cErr != nil {
				log.Fatalf("failed to close pipe writer: %s", cErr)
			}
		}()

		if _, err := io.Copy(writer, r); err != nil {
			log.Fatalf("page writer failed: %s", err)
		}
	}()

	return rPipe
}
