package shared

import (
	"fmt"
	"log"
	"time"
)

// Logger interface untuk logging konsisten
type Logger interface {
	Info(ctx, message string, args ...interface{})
	Error(ctx, message string, err error, args ...interface{})
	Debug(ctx, message string, args ...interface{})
}

type defaultLogger struct{}

func (l *defaultLogger) Info(ctx, message string, args ...interface{}) {
	prefix := fmt.Sprintf("[INFO] [%s] %s", time.Now().Format("15:04:05"), message)
	log.Printf("%s %v", prefix, args)
}

func (l *defaultLogger) Error(ctx, message string, err error, args ...interface{}) {
	prefix := fmt.Sprintf("[ERROR] [%s] %s", time.Now().Format("15:04:05"), message)
	log.Printf("%s: %v %v", prefix, err, args)
}

func (l *defaultLogger) Debug(ctx, message string, args ...interface{}) {
	prefix := fmt.Sprintf("[DEBUG] [%s] %s", time.Now().Format("15:04:05"), message)
	log.Printf("%s %v", prefix, args)
}

func NewLogger() Logger {
	return &defaultLogger{}
}
