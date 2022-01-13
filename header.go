package memr

import (
	"encoding/binary"
	"io"
	"log"
)

/*
LiME header format:

typedef struct {
    unsigned int magic;           // Always 0x4C694D45 (LiME)
    unsigned int version;         // Header version number
    unsigned long long s_addr;    // Starting address of physical RAM range
    unsigned long long e_addr;    // Ending address of physical RAM range
    unsigned char reserved[8];    // Currently all zeros
}
*/
const (
	limeMagic uint32 = 0x4C694D45
)

// PageHeaderProviderFunc should be used to provide a page header
type PageHeaderProviderFunc func(start, end uint64) interface{}

func newHeaderReader(h interface{}, ord binary.ByteOrder) io.Reader {
	rPipe, wPipe := io.Pipe()
	go func() {
		defer wPipe.Close()
		err := binary.Write(wPipe, ord, h)
		if err != nil {
			log.Fatalf("failed to serialize page header: %s", err)
		}
	}()

	return rPipe
}

// HeaderLime is a PageHeaderProviderFunc that
// returns a DefaultHeader struct
func HeaderLime(start, end uint64) interface{} {
	return &DefaultHeader{
		Magic:     limeMagic,
		Version:   1,
		StartAddr: start,
		EndAddr:   end - 1,
	}
}

// DefaultHeader is the default header struct value.
// This header can be used in conjunction with a PageHeaderProviderFunc
// to provide a custom header format to be used by the reader
type DefaultHeader struct {
	Magic     uint32  // magic (0x4C694D45 for LiME)
	Version   uint32  // version (always 1 for LiME)
	StartAddr uint64  // start address
	EndAddr   uint64  // end address
	reserved  [8]byte //nolint:unused,structcheck
}
