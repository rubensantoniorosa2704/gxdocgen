package utils

import (
	"fmt"
	"os"
)

// ANSI color codes
const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorCyan   = "\033[36m"
)

// Info logs an informational message with cyan color
func Info(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	fmt.Fprintf(os.Stdout, "%s[INFO]%s %s\n", colorCyan, colorReset, message)
}

// Success logs a success message with green color
func Success(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	fmt.Fprintf(os.Stdout, "%s[SUCCESS]%s %s\n", colorGreen, colorReset, message)
}

// Warning logs a warning message with yellow color
func Warning(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	fmt.Fprintf(os.Stderr, "%s[WARNING]%s %s\n", colorYellow, colorReset, message)
}

// Error logs an error message with red color
func Error(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	fmt.Fprintf(os.Stderr, "%s[ERROR]%s %s\n", colorRed, colorReset, message)
}

// Fatal logs a fatal error message and exits the program
func Fatal(format string, args ...interface{}) {
	Error(format, args...)
	os.Exit(1)
}
