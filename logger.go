package main

import (
	"log"
)

const (
	LOG_ERROR = iota + 1
	LOG_WARNING
	LOG_INFO
	LOG_DEBUG
)

func levelLogf(level int, fmt string, v ...interface{}) {
	if len(options.Verbosity) >= level {
		log.Printf(fmt, v...)
	}
}
