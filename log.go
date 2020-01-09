package main

import (
	"github.com/sirupsen/logrus"
	"io"
	"os"
)

var Log *logrus.Entry

const LOG_FILE = "pvs.log"

func init() {
	logger := logrus.New()
	logger.Formatter = &logrus.JSONFormatter{}
	logger.Level = logrus.DebugLevel
	//logger.SetOutput(os.Stdout)

	file, err := os.OpenFile(LOG_FILE, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0755)
	if err != nil {
		logger.Fatal(err)
	}
	//defer file.Close()
	logger.SetReportCaller(true)
	mw := io.MultiWriter(os.Stdout, file)
	logger.SetOutput(mw)
	Log = logger.WithFields(logrus.Fields{"prefix": "pvs"})
}
