package memr

import (
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"strings"
)

type blocks []*block

func (b blocks) String() string {
	var str []string
	for i, v := range b {
		str = append(str, fmt.Sprintf("[%02d] %s", i, v))
	}
	return strings.Join(str, "\n")
}

// block embeds the io.Reader used for reading actual pages of
// memory. This will either be an io.SectionReader that reads
// over raw memory sources (/dev/crash or /dev/mem) or extracts the
// pages (programs) from the /proc/kcore ELF file using *elf.Prog.Open()
type block struct {
	io.Reader
	start, end uint64
}

func (b *block) String() string {
	return fmt.Sprintf("start=%d; end=%d (raw reader: %T)", b.start, b.end, b.Reader)
}

func (b *block) size() uint64 {
	return b.end - b.start
}

func (r *Reader) initBlockReaders(blks blocks) (io.Reader, uint64) {

	var total uint64
	var readers []io.Reader
	for _, blk := range blks {
		if r.PageHeaderProvider != nil {
			header := r.PageHeaderProvider(blk.start, blk.end)
			total += uint64(binary.Size(header))
			readers = append(readers, r.bar.NewProxyReader(newHeaderReader(header, r.ByteOrder)))
		}

		total += blk.size()
		readers = append(readers, applyPageWriter(r.bar.NewProxyReader(blk), r.PageHandler))
	}

	log.Printf("[DEBUG] total size to be read: %d", total)

	return io.MultiReader(readers...), total
}
