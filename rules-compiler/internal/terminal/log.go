package terminal

import (
	"fmt"
	"os"
)

const (
	blue   = "\033[34m"
	cyan   = "\033[36m"
	green  = "\033[32m"
	red    = "\033[31m"
	yellow = "\033[33m"
	reset  = "\033[0m"
)

func Info(format string, args ...any) {
	write(cyan, format, args...)
}

func Step(format string, args ...any) {
	write(blue, format, args...)
}

func Success(format string, args ...any) {
	write(green, format, args...)
}

func Warn(format string, args ...any) {
	write(yellow, format, args...)
}

func Error(format string, args ...any) {
	write(red, format, args...)
}

func ColorizeLogLine(line string) string {
	if !colorEnabled() || line == "" {
		return line
	}
	return cyan + line + reset
}

func write(color string, format string, args ...any) {
	message := fmt.Sprintf(format, args...)
	if colorEnabled() {
		message = color + message + reset
	}
	fmt.Fprintln(os.Stderr, message)
}

func colorEnabled() bool {
	if os.Getenv("NO_COLOR") != "" {
		return false
	}
	return os.Getenv("TERM") != "dumb"
}
