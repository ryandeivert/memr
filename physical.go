package memr

import (
	"bufio"
	"io"
	"log"
	"os"

	"github.com/ryandeivert/memr/internal/iomem"
)

// physicalBlocks reads sections directly from an *os.File
// Certain character devices (eg: /dev/crash) require reads
// to be block-aligned, so allow reading on exact OS page size
func physicalBlocks(file io.ReaderAt, memRanges iomem.MemRanges, strictPages bool) (blks blocks) {
	pgsz := os.Getpagesize()
	for _, rng := range memRanges {
		end := rng.End
		if strictPages {
			end = end - (end % uint64(pgsz))
		}

		var blkRdr io.Reader = io.NewSectionReader(file, int64(rng.Start), int64(end-rng.Start))
		if strictPages {
			blkRdr = blockReader(blkRdr, pgsz)
		}
		blks = append(blks, &block{Reader: blkRdr, start: rng.Start, end: end})
	}

	return blks
}

// blockReader forces buffered reads on exact page sizes, typically 4096
// Note: wrapping the io.Reader with bufio.Reader directly (without the Pipe)
// does not work because io.Copy will always default to 32 KiB page sizes
func blockReader(r io.Reader, pgsz int) io.Reader {
	rPipe, wPipe := io.Pipe()
	buffer := bufio.NewReaderSize(r, pgsz)
	go func() {
		defer func() {
			pErr := wPipe.Close()
			if pErr != nil {
				log.Fatalf("failed to close block writer: %s", pErr)
			}
		}()
		c, err := io.Copy(wPipe, buffer)
		if err != nil {
			log.Fatalf("block read failed: %s (read=%d)", err, c)
		}
	}()

	return rPipe
}
