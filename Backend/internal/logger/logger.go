package logger

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/nessieaudio/ecommerce-backend/internal/services/email"
)

// ErrorLevel represents the severity of an error
type ErrorLevel string

const (
	LevelInfo     ErrorLevel = "INFO"
	LevelWarning  ErrorLevel = "WARNING"
	LevelError    ErrorLevel = "ERROR"
	LevelCritical ErrorLevel = "CRITICAL"
)

// Logger handles structured logging and error monitoring
type Logger struct {
	file        *os.File
	emailClient *email.Client
	adminEmail  string
}

// ErrorContext contains contextual information about an error
type ErrorContext struct {
	Level      ErrorLevel
	Message    string
	Error      error
	UserIP     string
	Endpoint   string
	Details    map[string]interface{}
	StackTrace string
}

// New creates a new logger instance
func New(logPath string, emailClient *email.Client, adminEmail string) (*Logger, error) {
	// Ensure logs directory exists
	logDir := filepath.Dir(logPath)
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}

	// Open or create log file with append mode
	file, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}

	return &Logger{
		file:        file,
		emailClient: emailClient,
		adminEmail:  adminEmail,
	}, nil
}

// Log writes a structured log entry
func (l *Logger) Log(ctx ErrorContext) {
	timestamp := time.Now().Format("2006-01-02 15:04:05")

	// Build log entry
	var logEntry strings.Builder
	logEntry.WriteString(fmt.Sprintf("[%s] [%s] %s\n", timestamp, ctx.Level, ctx.Message))

	if ctx.Error != nil {
		logEntry.WriteString(fmt.Sprintf("  Error: %v\n", ctx.Error))
	}

	if ctx.Endpoint != "" {
		logEntry.WriteString(fmt.Sprintf("  Endpoint: %s\n", ctx.Endpoint))
	}

	if ctx.UserIP != "" {
		logEntry.WriteString(fmt.Sprintf("  IP: %s\n", ctx.UserIP))
	}

	for key, value := range ctx.Details {
		logEntry.WriteString(fmt.Sprintf("  %s: %v\n", key, value))
	}

	if ctx.StackTrace != "" {
		logEntry.WriteString(fmt.Sprintf("  Stack Trace:\n%s\n", ctx.StackTrace))
	}

	logEntry.WriteString("---\n")

	// Write to file
	if _, err := l.file.WriteString(logEntry.String()); err != nil {
		log.Printf("Failed to write to log file: %v", err)
	}

	// Also log to console for immediate visibility
	log.Print(strings.TrimSuffix(logEntry.String(), "---\n"))

	// Send email alert for critical errors
	if ctx.Level == LevelCritical && l.emailClient != nil && l.adminEmail != "" {
		go l.sendEmailAlert(ctx, timestamp)
	}
}

// sendEmailAlert sends an email notification for critical errors
func (l *Logger) sendEmailAlert(ctx ErrorContext, timestamp string) {
	subject := fmt.Sprintf("[CRITICAL] Nessie Audio - %s", ctx.Message)

	var body strings.Builder
	body.WriteString(fmt.Sprintf("<h2>Critical Error Alert</h2>\n"))
	body.WriteString(fmt.Sprintf("<p><strong>Time:</strong> %s</p>\n", timestamp))
	body.WriteString(fmt.Sprintf("<p><strong>Message:</strong> %s</p>\n", ctx.Message))

	if ctx.Error != nil {
		body.WriteString(fmt.Sprintf("<p><strong>Error:</strong> %v</p>\n", ctx.Error))
	}

	if ctx.Endpoint != "" {
		body.WriteString(fmt.Sprintf("<p><strong>Endpoint:</strong> %s</p>\n", ctx.Endpoint))
	}

	if ctx.UserIP != "" {
		body.WriteString(fmt.Sprintf("<p><strong>User IP:</strong> %s</p>\n", ctx.UserIP))
	}

	if len(ctx.Details) > 0 {
		body.WriteString("<h3>Details:</h3>\n<ul>\n")
		for key, value := range ctx.Details {
			body.WriteString(fmt.Sprintf("<li><strong>%s:</strong> %v</li>\n", key, value))
		}
		body.WriteString("</ul>\n")
	}

	if ctx.StackTrace != "" {
		body.WriteString(fmt.Sprintf("<h3>Stack Trace:</h3>\n<pre>%s</pre>\n", ctx.StackTrace))
	}

	if err := l.emailClient.SendHTMLEmail(l.adminEmail, subject, body.String()); err != nil {
		log.Printf("Failed to send error alert email: %v", err)
	}
}

// Info logs an informational message
func (l *Logger) Info(message string) {
	l.Log(ErrorContext{
		Level:   LevelInfo,
		Message: message,
	})
}

// Warning logs a warning message
func (l *Logger) Warning(message string, err error) {
	l.Log(ErrorContext{
		Level:   LevelWarning,
		Message: message,
		Error:   err,
	})
}

// Error logs an error message
func (l *Logger) Error(message string, err error) {
	l.Log(ErrorContext{
		Level:      LevelError,
		Message:    message,
		Error:      err,
		StackTrace: getStackTrace(),
	})
}

// Critical logs a critical error and sends an email alert
func (l *Logger) Critical(message string, err error, details map[string]interface{}) {
	l.Log(ErrorContext{
		Level:      LevelCritical,
		Message:    message,
		Error:      err,
		Details:    details,
		StackTrace: getStackTrace(),
	})
}

// CriticalWithContext logs a critical error with full context
func (l *Logger) CriticalWithContext(ctx ErrorContext) {
	ctx.Level = LevelCritical
	if ctx.StackTrace == "" {
		ctx.StackTrace = getStackTrace()
	}
	l.Log(ctx)
}

// getStackTrace captures the current stack trace
func getStackTrace() string {
	buf := make([]byte, 4096)
	n := runtime.Stack(buf, false)
	return string(buf[:n])
}

// Close closes the log file
func (l *Logger) Close() error {
	if l.file != nil {
		return l.file.Close()
	}
	return nil
}

// RotateLogs rotates log files if they exceed a certain size (10MB)
func (l *Logger) RotateLogs(logPath string) error {
	info, err := l.file.Stat()
	if err != nil {
		return err
	}

	// If file is less than 10MB, don't rotate
	if info.Size() < 10*1024*1024 {
		return nil
	}

	// Close current file
	if err := l.file.Close(); err != nil {
		return err
	}

	// Rename current log to timestamped backup
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	backupPath := strings.TrimSuffix(logPath, ".log") + "_" + timestamp + ".log"
	if err := os.Rename(logPath, backupPath); err != nil {
		return err
	}

	// Open new log file
	file, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}

	l.file = file
	l.Info("Log file rotated")
	return nil
}
