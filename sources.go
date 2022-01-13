package memr

// MemSource type used for memory sources
type MemSource string

const (
	// SourceKcore disignates /proc/kcore as the memory source
	SourceKcore MemSource = "/proc/kcore"
	// SourceCrash disignates /dev/crash as the memory source
	SourceCrash MemSource = "/dev/crash"
	// SourceMem disignates /dev/mem as the memory source
	SourceMem MemSource = "/dev/mem"
)

func allMemSources() []MemSource {
	return []MemSource{
		SourceKcore,
		SourceCrash,
		SourceMem,
	}
}

func (d MemSource) isPhysical() bool {
	return d != SourceKcore
}

func (d MemSource) forcePageReads() bool {
	return d == SourceCrash
}
