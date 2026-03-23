package logger

import (
	"log"
	"os"
	"strings"
	"sync"
)

const (
	levelDebug = 10
	levelInfo  = 20
	levelWarn  = 30
	levelError = 40
)

var (
	once         sync.Once
	currentLevel = levelInfo
)

func Init() {
	once.Do(func() {
		log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds)
		switch strings.ToLower(strings.TrimSpace(os.Getenv("LOG_LEVEL"))) {
		case "debug":
			currentLevel = levelDebug
		case "info", "":
			currentLevel = levelInfo
		case "warn", "warning":
			currentLevel = levelWarn
		case "error":
			currentLevel = levelError
		default:
			currentLevel = levelInfo
			log.Printf("[WARN] unknown LOG_LEVEL=%q, fallback to info", os.Getenv("LOG_LEVEL"))
		}
	})
}

func shouldLog(level int) bool {
	return level >= currentLevel
}

func Debugf(format string, args ...interface{}) {
	if shouldLog(levelDebug) {
		log.Printf("[DEBUG] "+format, args...)
	}
}

func Infof(format string, args ...interface{}) {
	if shouldLog(levelInfo) {
		log.Printf("[INFO] "+format, args...)
	}
}

func Warnf(format string, args ...interface{}) {
	if shouldLog(levelWarn) {
		log.Printf("[WARN] "+format, args...)
	}
}

func Errorf(format string, args ...interface{}) {
	if shouldLog(levelError) {
		log.Printf("[ERROR] "+format, args...)
	}
}

func Fatalf(format string, args ...interface{}) {
	log.Fatalf("[FATAL] "+format, args...)
}
