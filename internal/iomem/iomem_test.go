package iomem

import (
	"bytes"
	"testing"
)

func TestReadRanges(t *testing.T) {
	var buffer bytes.Buffer
	buffer.WriteString(`00000000-00000fff : reserved
00001000-0009ffff : System RAM
000a0000-000bffff : PCI Bus 0000:00
000c0000-000c7fff : Video ROM
000f0000-000fffff : System ROM
00100000-54d93fff : System RAM
  01000000-01556eb4 : Kernel code
  01556eb5-01c2170f : Kernel data
  01d77000-02045963 : Kernel bss
  03000000-0c8fffff : System RAM
54d94000-54da2fff : reserved
  54d98018-54d98067 : APEI ERST
  54d98070-54d98077 : APEI ERST
54da3000-5a12dfff : System RAM
5a12e000-6a1aefff : reserved
6a1af000-6c65ffff : System RAM
6c660000-6c691fff : reserved
6c692000-6de34fff : System RAM
6de35000-793fefff : reserved
`)
	ranges, err := ranges(&buffer)
	if err != nil {
		t.Error("Failed to read ranges", err)
	}

	expectedRanges := []*MemRange{
		{4096, 655359},
		{1048576, 1423523839},
		{1423585280, 1511186431},
		{1780150272, 1818623999},
		{1818828800, 1843613695},
	}

	for i, rng := range ranges {
		if rng.Start != expectedRanges[i].Start {
			t.Errorf("[%d] invalid start extracted: %d != %d", i, rng.Start, expectedRanges[i].Start)
		}
		if rng.End != expectedRanges[i].End {
			t.Errorf("[%d] invalid end extracted: %d != %d", i, rng.End, expectedRanges[i].End)
		}
	}
}
