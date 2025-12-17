// Package logger provides colorful logging for CLI output.
package logger

import (
	"fmt"
	"io"

	"github.com/fatih/color"
)

// Logger provides colorful console output.
type Logger struct {
	out     io.Writer
	info    *color.Color
	success *color.Color
	warn    *color.Color
	err     *color.Color
	bold    *color.Color
	dim     *color.Color
}

// New creates a new Logger writing to the given output.
func New(out io.Writer) *Logger {
	return &Logger{
		out:     out,
		info:    color.New(color.FgCyan),
		success: color.New(color.FgGreen),
		warn:    color.New(color.FgYellow),
		err:     color.New(color.FgRed),
		bold:    color.New(color.Bold),
		dim:     color.New(color.Faint),
	}
}

// Info prints an informational message in cyan.
func (l *Logger) Info(format string, args ...interface{}) {
	l.info.Fprintf(l.out, "ℹ "+format+"\n", args...)
}

// Success prints a success message in green.
func (l *Logger) Success(format string, args ...interface{}) {
	l.success.Fprintf(l.out, "✓ "+format+"\n", args...)
}

// Warn prints a warning message in yellow.
func (l *Logger) Warn(format string, args ...interface{}) {
	l.warn.Fprintf(l.out, "⚠ "+format+"\n", args...)
}

// Error prints an error message in red.
func (l *Logger) Error(format string, args ...interface{}) {
	l.err.Fprintf(l.out, "✗ "+format+"\n", args...)
}

// Step prints a processing step with an arrow.
func (l *Logger) Step(format string, args ...interface{}) {
	l.info.Fprint(l.out, "→ ")
	fmt.Fprintf(l.out, format+"\n", args...)
}

// Bold prints bold text.
func (l *Logger) Bold(format string, args ...interface{}) {
	l.bold.Fprintf(l.out, format, args...)
}

// Dim prints dimmed text.
func (l *Logger) Dim(format string, args ...interface{}) {
	l.dim.Fprintf(l.out, format, args...)
}

// Print prints plain text.
func (l *Logger) Print(format string, args ...interface{}) {
	fmt.Fprintf(l.out, format, args...)
}

// Println prints plain text with newline.
func (l *Logger) Println(format string, args ...interface{}) {
	fmt.Fprintf(l.out, format+"\n", args...)
}

// Header prints a bold header with separator.
func (l *Logger) Header(title string) {
	l.bold.Fprintf(l.out, "\n%s\n", title)
	l.dim.Fprintf(l.out, "─────────────────────────────────\n")
}

// KeyValue prints a key-value pair with the key dimmed.
func (l *Logger) KeyValue(key, value string) {
	l.dim.Fprintf(l.out, "  %s: ", key)
	fmt.Fprintf(l.out, "%s\n", value)
}

// Color helper functions for templates.
var (
	Cyan    = color.New(color.FgCyan).SprintFunc()
	Green   = color.New(color.FgGreen).SprintFunc()
	Yellow  = color.New(color.FgYellow).SprintFunc()
	Bold    = color.New(color.Bold).SprintFunc()
	Dim     = color.New(color.Faint).SprintFunc()
	Magenta = color.New(color.FgMagenta).SprintFunc()
)
