package main

import (
	"github.com/sirupsen/logrus"
	"testing"
)

func TestFormat(t *testing.T) {

	t.Log("Testing logs for level prefix")
	formatter := interacterLogFormatter{}
	debugLogLine := "[DEBUG]\tThis is a debug message!"
	infoLogLine := "[INFO]\tThis is an info message!"
	warningLogLine := "[WARNING]\tThis is a warning message!"
	errorLogLine := "[ERROR]\tThis is an error message!"

	testEntry := logrus.Entry{
		Level:   logrus.DebugLevel,
		Message: "This is a debug message!",
	}
	logLine, err := formatter.Format(&testEntry)
	if err != nil || string(logLine) != debugLogLine {
		t.Errorf("Expected logLine: %s; got %s", debugLogLine, logLine)
	}

	testEntry = logrus.Entry{
		Level:   logrus.InfoLevel,
		Message: "This is an info message!",
	}
	logLine, err = formatter.Format(&testEntry)
	if err != nil || string(logLine) != infoLogLine {
		t.Errorf("Expected logLine: %s; got %s", infoLogLine, logLine)
	}

	testEntry = logrus.Entry{
		Level:   logrus.WarnLevel,
		Message: "This is a warning message!",
	}
	logLine, err = formatter.Format(&testEntry)
	if err != nil || string(logLine) != warningLogLine {
		t.Errorf("Expected logLine: %s; got %s", warningLogLine, logLine)
	}

	testEntry = logrus.Entry{
		Level:   logrus.ErrorLevel,
		Message: "This is an error message!",
	}
	logLine, err = formatter.Format(&testEntry)
	if err != nil || string(logLine) != errorLogLine {
		t.Errorf("Expected logLine: %s; got %s", errorLogLine, logLine)
	}
}
