package log

import "github.com/sirupsen/logrus"

type SLoger struct {
	Log *logrus.Logger
}

func InitServiceLoger(level string) SLoger {
	var Log *logrus.Logger = logrus.New()

	logLevel, err := logrus.ParseLevel(level)
	if err != nil {
		logrus.Panicln(err)
	}

	Log.SetLevel(logLevel)
	Log.SetFormatter(&logrus.TextFormatter{
		ForceColors:   true,
		FullTimestamp: true,
	})

	return SLoger{
		Log: Log,
	}
}
