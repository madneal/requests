package main

import (
	"github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"os"
)

var Log *logrus.Entry

const LOG_FILE = "pvs.log"

type WriterHook struct {
	Writer    io.Writer
	LogLevels []logrus.Level
}

func (hook *WriterHook) Fire(entry *logrus.Entry) error {
	line, err := entry.String()
	if err != nil {
		return err
	}
	_, err = hook.Writer.Write([]byte(line))
	return err
}

func (hook *WriterHook) Levels() []logrus.Level {
	return hook.LogLevels
}

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
	//logger.SetReportCaller(true)
	//mw := io.MultiWriter(os.Stdout, file)
	//logger.SetOutput(mw)
	fileWriter := io.Writer(file)
	logger.SetOutput(ioutil.Discard)
	logger.AddHook(&WriterHook{
		Writer:    fileWriter,
		LogLevels: []logrus.Level{logrus.PanicLevel, logrus.FatalLevel, logrus.ErrorLevel, logrus.WarnLevel},
	})
	logger.AddHook(&WriterHook{
		Writer:    os.Stdout,
		LogLevels: []logrus.Level{logrus.InfoLevel, logrus.DebugLevel, logrus.ErrorLevel},
	})
	Log = logger.WithFields(logrus.Fields{"prefix": "pvs"})
}
