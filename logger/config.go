package logger

import "strings"

// LoggerLevelConf points on a logger level in config
type LoggerLevelConf string

const (
	// LoggerLevelDEBUG -
	LoggerLevelDEBUG LoggerLevelConf = "DEBUG"
	// LoggerLevelINFO -
	LoggerLevelINFO LoggerLevelConf = "INFO"
	// LoggerLevelWARN -
	LoggerLevelWARN LoggerLevelConf = "WARN"
	// LoggerLevelERROR -
	LoggerLevelERROR LoggerLevelConf = "ERROR"
	// LoggerLevelFATAL -
	LoggerLevelFATAL LoggerLevelConf = "FATAL"
)

// Variants -
func (LoggerLevelConf) Variants() []LoggerLevelConf {
	return []LoggerLevelConf{
		LoggerLevelDEBUG, LoggerLevelINFO, LoggerLevelWARN, LoggerLevelERROR, LoggerLevelFATAL,
	}
}

// Eq -
func (ll LoggerLevelConf) Eq(o LoggerLevelConf) bool {
	return strings.EqualFold(string(ll), string(o))
}
