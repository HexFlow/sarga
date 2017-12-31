package slog

import (
	"log"
)

type Level int

const (
	Zero Level = iota
	None
	Error
	Debug
	Verbose
	VVerbose
)

func GetLevelFromString(a string) Level {
	switch a {
	case "none":
		return None
	case "error":
		return Error
	case "debug":
		return Debug
	case "verbose":
		return Verbose
	case "vverbose":
		return VVerbose
	default:
		return Error
	}
}

type SLog struct {
	Level Level
}

func (l *SLog) Println(level Level, a ...interface{}) {
	if l.Level >= level {
		log.Println(a...)
	}
}

func (l *SLog) Printf(level Level, format string, a ...interface{}) {
	if l.Level >= level {
		log.Printf(format, a...)
	}
}
