package logging

import "github.com/sirupsen/logrus"

type Event struct {
	ID      int
	Message string
	Error   error
}

type StandardLogger struct {
	*logrus.Logger
}

func NewLogger() *StandardLogger {
	var baseLogger = logrus.New()

	var standardLogger = &StandardLogger{baseLogger}

	standardLogger.Formatter = &logrus.JSONFormatter{}

	return standardLogger
}

// Misconfigured throws error if some object is misconfigured
func (l *StandardLogger) Misconfigured(message string, errorMsg error) {

	l.Errorf("Misconfigured: %s , Error: %v", message, errorMsg)
}

func (l *StandardLogger) Unsuccessful(message string, errorMsg error) {
	l.Errorf("Unsuccessful: %s, Error: %v", message, errorMsg)
}

// Success is a form of info that prints if there is an success
func (l *StandardLogger) Success(message string) {
	l.Infof("Successful %s", message)
}
