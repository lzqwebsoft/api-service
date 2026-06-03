package utils

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// DailyLogWriter is a thread-safe writer that writes to a daily file under a specific directory.
type DailyLogWriter struct {
	mu   sync.Mutex
	dir  string
	file *os.File
	day  string
}

// NewDailyLogWriter creates and initializes a DailyLogWriter.
func NewDailyLogWriter(dir string) (*DailyLogWriter, error) {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}
	w := &DailyLogWriter{dir: dir}
	if err := w.rotate(); err != nil {
		return nil, err
	}
	return w, nil
}

// rotate checks if the date has changed, closes the current log file if so, and opens a new one.
func (w *DailyLogWriter) rotate() error {
	today := time.Now().Format("2006-01-02")
	if w.file != nil && w.day == today {
		return nil
	}
	if w.file != nil {
		_ = w.file.Close()
	}
	filename := filepath.Join(w.dir, today+".log")
	f, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("failed to open log file %s: %w", filename, err)
	}
	w.file = f
	w.day = today
	return nil
}

// Write writes the bytes to the active log file, rotating if needed.
func (w *DailyLogWriter) Write(p []byte) (n int, err error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	if err := w.rotate(); err != nil {
		return 0, err
	}
	return w.file.Write(p)
}

// Close closes the active log file.
func (w *DailyLogWriter) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.file != nil {
		err := w.file.Close()
		w.file = nil
		return err
	}
	return nil
}

// Logger formats messages and forwards them to its standard and file writers.
type Logger struct {
	writer io.Writer
}

var std *Logger

// InitLogger initializes the global logger to log to stdout and to date-split files under the specified directory.
func InitLogger(dir string) error {
	fileWriter, err := NewDailyLogWriter(dir)
	if err != nil {
		return err
	}
	// Output concurrently to stdout and fileWriter
	std = &Logger{
		writer: io.MultiWriter(os.Stdout, fileWriter),
	}
	return nil
}

func (l *Logger) log(level, msg string) {
	now := time.Now().Format("2006/01/02 15:04:05")
	if !strings.HasSuffix(msg, "\n") {
		msg += "\n"
	}
	_, _ = fmt.Fprintf(l.writer, "%s [%s] %s", now, level, msg)
}

// Debug logs debug messages.
func Debug(args ...interface{}) {
	if std != nil {
		std.log("DEBUG", fmt.Sprint(args...))
	} else {
		fmt.Print(time.Now().Format("2006/01/02 15:04:05") + " [DEBUG] " + fmt.Sprintln(args...))
	}
}

// Debugf logs formatted debug messages.
func Debugf(format string, args ...interface{}) {
	if std != nil {
		std.log("DEBUG", fmt.Sprintf(format, args...))
	} else {
		fmt.Printf(time.Now().Format("2006/01/02 15:04:05")+" [DEBUG] "+format+"\n", args...)
	}
}

// Info logs info messages.
func Info(args ...interface{}) {
	if std != nil {
		std.log("INFO", fmt.Sprint(args...))
	} else {
		fmt.Print(time.Now().Format("2006/01/02 15:04:05") + " [INFO] " + fmt.Sprintln(args...))
	}
}

// Infof logs formatted info messages.
func Infof(format string, args ...interface{}) {
	if std != nil {
		std.log("INFO", fmt.Sprintf(format, args...))
	} else {
		fmt.Printf(time.Now().Format("2006/01/02 15:04:05")+" [INFO] "+format+"\n", args...)
	}
}

// Warn logs warn messages.
func Warn(args ...interface{}) {
	if std != nil {
		std.log("WARN", fmt.Sprint(args...))
	} else {
		fmt.Print(time.Now().Format("2006/01/02 15:04:05") + " [WARN] " + fmt.Sprintln(args...))
	}
}

// Warnf logs formatted warn messages.
func Warnf(format string, args ...interface{}) {
	if std != nil {
		std.log("WARN", fmt.Sprintf(format, args...))
	} else {
		fmt.Printf(time.Now().Format("2006/01/02 15:04:05")+" [WARN] "+format+"\n", args...)
	}
}

// Error logs error messages.
func Error(args ...interface{}) {
	if std != nil {
		std.log("ERROR", fmt.Sprint(args...))
	} else {
		fmt.Print(time.Now().Format("2006/01/02 15:04:05") + " [ERROR] " + fmt.Sprintln(args...))
	}
}

// Errorf logs formatted error messages.
func Errorf(format string, args ...interface{}) {
	if std != nil {
		std.log("ERROR", fmt.Sprintf(format, args...))
	} else {
		fmt.Printf(time.Now().Format("2006/01/02 15:04:05")+" [ERROR] "+format+"\n", args...)
	}
}

// Fatal logs fatal messages and exits with status 1.
func Fatal(args ...interface{}) {
	if std != nil {
		std.log("FATAL", fmt.Sprint(args...))
	} else {
		fmt.Print(time.Now().Format("2006/01/02 15:04:05") + " [FATAL] " + fmt.Sprintln(args...))
	}
	os.Exit(1)
}

// Fatalf logs formatted fatal messages and exits with status 1.
func Fatalf(format string, args ...interface{}) {
	if std != nil {
		std.log("FATAL", fmt.Sprintf(format, args...))
	} else {
		fmt.Printf(time.Now().Format("2006/01/02 15:04:05")+" [FATAL] "+format+"\n", args...)
	}
	os.Exit(1)
}
