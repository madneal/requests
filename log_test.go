package main

import (
	"github.com/sirupsen/logrus"
	"testing"
)

func TestLog(t *testing.T) {
	Log = Log.WithFields(logrus.Fields{"url": "adfasd412384198jj"})
	Log.Info("test1234")
	Log.Info("afasdfqwerqw")
}
