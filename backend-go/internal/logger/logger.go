package logger

import (
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"sync"

	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	crawlerLogger *slog.Logger
	once          sync.Once
)

// Init initializes the logging system.
func Init() {
	once.Do(func() {
		// Ensure logs directory exists
		if err := os.MkdirAll("logs", 0755); err != nil {
			panic(err)
		}

		// Configure Lumberjack for log rotation
		lj := &lumberjack.Logger{
			Filename:   filepath.Join("logs", "crawler.log"),
			MaxSize:    10, // megabytes per log file
			MaxBackups: 5,  // retain 5 old files
			MaxAge:     7,  // retain for 7 days
			Compress:   true, // compress old files
		}

		// Write to both console and log file
		multiWriter := io.MultiWriter(os.Stdout, lj)

		// We use TextHandler for human readability, easily parsed line by line in web viewer
		handler := slog.NewTextHandler(multiWriter, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		})

		crawlerLogger = slog.New(handler)
	})
}

// Crawler returns the dedicated crawler logger
func Crawler() *slog.Logger {
	if crawlerLogger == nil {
		Init()
	}
	return crawlerLogger
}
