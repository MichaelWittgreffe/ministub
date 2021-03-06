package logger

// Logger defines an object suitable for performing logging
type Logger interface {
	Info(msg string)
	Error(msg string)
}
