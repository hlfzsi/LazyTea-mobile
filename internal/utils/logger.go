package utils
import (
	"fmt"
	"log"
	"os"
	"time"
)
type LogLevel int
const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
)
type Logger struct {
	level LogLevel
}
func NewLogger() *Logger {
	return &Logger{
		level: DEBUG,
	}
}
func (l *Logger) SetLevel(level LogLevel) {
	l.level = level
}
func (l *Logger) Debug(format string, args ...interface{}) {
	if l.level <= DEBUG {
		l.log("DEBUG", format, args...)
	}
}
func (l *Logger) Info(format string, args ...interface{}) {
	if l.level <= INFO {
		l.log("INFO", format, args...)
	}
}
func (l *Logger) Warn(format string, args ...interface{}) {
	if l.level <= WARN {
		l.log("WARN", format, args...)
	}
}
func (l *Logger) Error(format string, args ...interface{}) {
	if l.level <= ERROR {
		l.log("ERROR", format, args...)
	}
}
func (l *Logger) log(level string, format string, args ...interface{}) {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	message := fmt.Sprintf(format, args...)
	logMessage := fmt.Sprintf("[%s] %s: %s", timestamp, level, message)
	log.Println(logMessage)
	if level == "ERROR" {
		fmt.Fprintf(os.Stderr, "%s\n", logMessage)
	}
}
func GenerateID() string {
	return fmt.Sprintf("id-%d", time.Now().UnixNano())
}