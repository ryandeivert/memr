package memr

import (
	"debug/elf"
	"fmt"
	"log"
	"os"
	"sort"

	"github.com/ryandeivert/memr/internal/iomem"
)

const minSourceSize = 4096

// kcoreBlocks reads logical blocks from the *elf.File as elf.Progs
func kcoreBlocks(file *elf.File, memRanges iomem.MemRanges) blocks {

	rangeMap := memRanges.ToMap()
	sort.SliceStable(file.Progs, func(i, j int) bool { return file.Progs[i].Vaddr < file.Progs[j].Vaddr })

	var blks blocks
	var firstVaddr uint64
	for _, progHeader := range file.Progs {
		// Only care about PT_LOAD program header types
		if progHeader.Type != elf.PT_LOAD {
			continue
		}

		if firstVaddr == 0 {
			firstVaddr = progHeader.Vaddr - memRanges[0].Start
			log.Printf("[DEBUG] first kcore vaddr: %d", firstVaddr)
		}

		startValue := progHeader.Paddr
		if startValue == 0 {
			log.Printf("[DEBUG] kcore physical address unavailable, resorting to virtual address: %d", progHeader.Vaddr)
			startValue = progHeader.Vaddr - firstVaddr
		}

		// Skip if address is not in the iomem start values
		_, ok := rangeMap[startValue]
		if !ok {
			log.Printf("[DEBUG] kcore address not found in memory ranges: %d", startValue)
			continue
		}

		blks = append(blks, &block{
			Reader: progHeader.Open(),
			start:  progHeader.Paddr,
			end:    (progHeader.Paddr + progHeader.Filesz),
		})
	}

	return blks
}

// verifySource checks if a source can be read
// This currently only applies to /proc/kcore
func verifySource(path string) error {
	fi, err := os.Stat(path)
	if err != nil {
		return err
	}

	if fi.Size() <= minSourceSize {
		return fmt.Errorf("%s unavailable", path)
	}

	return nil
}
