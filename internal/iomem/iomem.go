package iomem

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
)

type MemRanges []*MemRange

type MemRange struct {
	Start, End uint64
}

func (m MemRange) String() string {
	return fmt.Sprintf("start=%d; end=%d", m.Start, m.End)
}

// ReadRanges reads from /proc/iomem
func ReadRanges() (MemRanges, error) {
	file, err := os.Open("/proc/iomem")
	if err != nil {
		return nil, err
	}

	return ranges(file)
}

func ranges(file io.Reader) (MemRanges, error) {
	// Valid lines look like:
	// 00100000-07ffffff : System RAM
	scanner := bufio.NewScanner(file)
	var ranges MemRanges
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, " ") {
			continue
		}
		parts := strings.Split(line, " : ")
		if len(parts) != 2 || parts[1] != "System RAM" {
			continue
		}

		startAndEnd := strings.Split(parts[0], "-")
		if len(startAndEnd) != 2 {
			continue
		}

		var err error
		curRange := new(MemRange)
		if curRange.Start, err = strconv.ParseUint(startAndEnd[0], 16, 64); err != nil {
			log.Printf("[WARN] invalid start of range, skipping: %s (%s)", startAndEnd[0], err)
			continue
		}
		if curRange.End, err = strconv.ParseUint(startAndEnd[1], 16, 64); err != nil {
			log.Printf("[WARN] invalid end of range, skipping: %s (%s)", startAndEnd[1], err)
			continue
		}

		ranges = append(ranges, curRange)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	if len(ranges) == 0 {
		return nil, fmt.Errorf("no valid ranges in /proc/iomem")
	}

	log.Printf("[DEBUG] loaded ranges:\n%+v", ranges)
	return ranges, nil
}

func (m MemRanges) ToMap() map[uint64]interface{} {
	rangeMap := make(map[uint64]interface{})
	for _, rng := range m {
		rangeMap[rng.Start] = nil // ignore end, unused
	}
	return rangeMap
}

func (m MemRanges) String() string {
	var str []string
	for i, rng := range m {
		str = append(str, fmt.Sprintf("[%02d] %s", i, rng))
	}
	return strings.Join(str, "\n")
}
