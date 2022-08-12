package main

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"strings"
)

type interacterLogFormatter struct {
	logrus.TextFormatter
}

func (f *interacterLogFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	return []byte(fmt.Sprintf("[%s]\t%s\n", strings.ToUpper(entry.Level.String()), entry.Message)), nil
}
