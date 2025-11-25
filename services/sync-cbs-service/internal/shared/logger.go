package shared

import (
	"encoding/json"
	"log"
	"os"
)

type Logger struct {
	logger *log.Logger
}

func NewLogger() *Logger {
	return &Logger{
		logger: log.New(os.Stdout, "[SYNC-CBS] ", log.LstdFlags|log.Lshortfile),
	}
}

func (l *Logger) Info(message string, fields map[string]interface{}) {
	l.log("INFO", message, fields)
}

func (l *Logger) Warn(message string, fields map[string]interface{}) {
	l.log("WARN", message, fields)
}

func (l *Logger) Error(message string, fields map[string]interface{}) {
	l.log("ERROR", message, fields)
}

func (l *Logger) Debug(message string, fields map[string]interface{}) {
	l.log("DEBUG", message, fields)
}

func (l *Logger) log(level string, message string, fields map[string]interface{}) {
	logEntry := map[string]interface{}{
		"level":   level,
		"message": message,
	}
	if fields != nil {
		for k, v := range fields {
			logEntry[k] = v
		}
	}
	data, _ := json.Marshal(logEntry)
	l.logger.Printf("%s", string(data))
}
