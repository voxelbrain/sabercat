package main

import (
	"log"
)

const (
	LOG_ERROR   = iota
	LOG_WARNING = iota
	LOG_INFO    = iota
	LOG_DEBUG   = iota
)

func levelLogf(level int, fmt string, v ...interface{}) {
	if len(options.Verbosity) >= level {
		log.Printf(fmt, v...)
	}
}
