package mlog

import (
	"log"
	"os"
)

type Log struct {
	Trace   *log.Logger
	Info    *log.Logger
	Warning *log.Logger
	Error   *log.Logger
}

func New() Log {

	trace := log.New(os.Stdin,
		"TRACE: ",
		log.Ldate|log.Ltime|log.Lmicroseconds|log.Lshortfile)

	info := log.New(os.Stdin,
		"INFO: ",
		log.Ldate|log.Ltime|log.Lmicroseconds|log.Lshortfile)

	warning := log.New(os.Stdout,
		"WARNING: ",
		log.Ldate|log.Ltime|log.Lmicroseconds|log.Lshortfile)
	error := log.New(os.Stderr,
		"ERROR: ",
		log.Ldate|log.Ltime|log.Lmicroseconds|log.Lshortfile)

	l := Log{Info: info, Error: error, Trace: trace, Warning: warning}

	return l

}
