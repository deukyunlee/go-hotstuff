package logging

import (
	"github.com/sirupsen/logrus"
	"path/filepath"
	"runtime"
	"strconv"
)

var logging *logrus.Logger

func init() {
	logging = logrus.New()
	logging.SetLevel(logrus.DebugLevel)

	logging.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05.000000",
		CallerPrettyfier: func(f *runtime.Frame) (string, string) {
			filename := filepath.Base(f.File)
			return "", filename + ":" + strconv.Itoa(f.Line)
		},
	})

	logging.SetReportCaller(true)
}

func GetLogger() *logrus.Logger {
	return logging
}
