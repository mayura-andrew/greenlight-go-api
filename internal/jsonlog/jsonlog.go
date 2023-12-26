package jsonlog

import (
	"encoding/json"
	"io"
	"os"
	"runtime/debug"
	"sync"
	"time"
)

// define a level type to represent the severity level for a log entry.
type Level int8

// initialize constants which represent a specific severity level. we use the iota
// keyword as a shortcut to assign successive integer values to the constants.

const (
	LevelInfo  Level = iota // has the value 0
	LevelError              // has the value 1.
	LevelFatal              // has the value 2
	LevelOff                // has the value 3
)

// return a human-friendly string for the severity level.

func (l Level) String() string {
	switch l {
	case LevelInfo:
		return "INFO"
	case LevelError:
		return "ERROR"
	case LevelFatal:
		return "FATAL"
	default:
		return ""
	}
}

// define a custom Logger type. This holds the output destination that the log entries
// will be written to, the minimum severity level that log entries will be written for,
// plus a mutex for coordinating the writes.

type Logger struct {
	out      io.Writer
	minLevel Level
	mu       sync.Mutex
}

// return a new Logger instance which writes log entries at or above a minimum
// severity level to a specific output destination.
func New(out io.Writer, minLevel Level) *Logger {
	return &Logger{
		out:      out,
		minLevel: minLevel,
	}
}

// declare some helper methods for writing log entries at the different levels.
// notice that these all accept a map as the second parameter which can contain any arbitrary
// "properties" that you want to appear in the log entry.

func (l *Logger) PrintInfo(message string, properties map[string]string) {
	l.print(LevelInfo, message, properties)
}

func (l *Logger) PrintError(err error, properties map[string]string) {
	l.print(LevelError, err.Error(), properties)
}

func (l *Logger) PrintFatal(err error, properties map[string]string) {
	l.print(LevelFatal, err.Error(), properties)
	os.Exit(1) // for entries at the FATAL level, we also terminate the application
}

func (l *Logger) print(level Level, message string, properties map[string]string) (int, error) {
	// if the severity level of the log entry is below the minimum severity for the
	// logger, then return with no further action.
	if level < l.minLevel {
		return 0, nil
	}

	// Declare an anonymous struct holding the data for the log entry.

	aux := struct {
		Level      string            `json:"level"`
		Time       string            `json:"time"`
		Message    string            `json:"message"`
		Properties map[string]string `json:"properties,omitempty"`
		Trace      string            `json:"trace,omitempty"`
	}{
		Level:      level.String(),
		Time:       time.Now().UTC().Format(time.RFC3339),
		Message:    message,
		Properties: properties,
	}

	// include a stack trace for entries at the ERROR and FATAL levels.

	if level >= LevelError {
		aux.Trace = string(debug.Stack())
	}

	// declare a line variable for holding the actual log entry text.
	var line []byte

	// Marshal the anonymous struct to JSON and store it in the line variable.
	// if there was a problem creating the JSON, set the contents of the log entry to be that
	// plain-text error message instead.

	line, err := json.Marshal(aux)
	if err != nil {
		line = []byte(LevelError.String() + ": unable to marshal log message:" + err.Error())
	}

	// Lock the mutex so that no two writes to the output destination
	// cannot happen concurrently. if we don't do this, it's possible that the text for two or more log entries will be
	// intermingled in the output.
	l.mu.Lock()
	defer l.mu.Unlock()

	// write the log entry followed by a newline.
	return l.out.Write(append(line, '\n'))

}
