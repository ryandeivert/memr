package memr

import (
	"log"
	"os"

	"github.com/hashicorp/logutils"
)

type LogLvl int

const (
	LogUnknown LogLvl = iota
	LogWarn
	LogInfo
	LogDebug
)

func (s LogLvl) String() string {
	switch s {
	case LogDebug:
		return "DEBUG"
	case LogInfo:
		return "INFO"
	}
	return "WARN" // unset or warn
}

func init() {
	SetLogLevel(LogWarn) // default to warn and above only
}

// SetupLogging sets logging level on the logger using the specified LogLvl value
func SetLogLevel(lvl LogLvl) {

	filter := &logutils.LevelFilter{
		Levels:   []logutils.LogLevel{"DEBUG", "INFO", "WARN", "ERROR"},
		MinLevel: logutils.LogLevel(lvl.String()),
		Writer:   os.Stderr, // os.Stderr matches what the built-in log package uses
	}

	log.SetOutput(filter)
}
