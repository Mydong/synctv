package bootstrap

import (
	"context"
	"io"
	"log"
	"os"

	"github.com/natefinch/lumberjack"
	"github.com/sirupsen/logrus"
	"github.com/synctv-org/synctv/cmd/flags"
	"github.com/synctv-org/synctv/internal/conf"
	"github.com/synctv-org/synctv/utils"
	"github.com/zijiren233/go-colorable"
)

func setLog(l *logrus.Logger) {
	if flags.Dev {
		l.SetLevel(logrus.DebugLevel)
		l.SetReportCaller(true)
	} else {
		l.SetLevel(logrus.InfoLevel)
		l.SetReportCaller(false)
	}
}

func InitLog(ctx context.Context) error {
	setLog(logrus.StandardLogger())
	if conf.Conf.Log.Enable {
		utils.OptFilePath(&conf.Conf.Log.FilePath)
		var l = &lumberjack.Logger{
			Filename:   conf.Conf.Log.FilePath,
			MaxSize:    conf.Conf.Log.MaxSize,
			MaxBackups: conf.Conf.Log.MaxBackups,
			MaxAge:     conf.Conf.Log.MaxAge,
			Compress:   conf.Conf.Log.Compress,
		}
		if err := l.Rotate(); err != nil {
			logrus.Fatalf("log: rotate log file error: %v", err)
		}
		var w io.Writer = colorable.NewNonColorableWriter(l)
		if flags.Dev || flags.LogStd {
			logrus.SetOutput(io.MultiWriter(os.Stdout, w))
			logrus.Infof("log: enable log to stdout and file: %s", conf.Conf.Log.FilePath)
		} else {
			logrus.SetOutput(w)
			logrus.Infof("log: disable log to stdout, only log to file: %s", conf.Conf.Log.FilePath)
		}
	}
	switch conf.Conf.Log.LogFormat {
	case "json":
		logrus.SetFormatter(&logrus.JSONFormatter{})
	default:
		if conf.Conf.Log.LogFormat != "text" {
			logrus.Warnf("unknown log format: %s, use default: text", conf.Conf.Log.LogFormat)
		}
		if colorable.IsTerminal(os.Stdout.Fd()) {
			logrus.SetFormatter(&logrus.TextFormatter{
				ForceColors: true,
			})
		} else {
			logrus.SetFormatter(&logrus.TextFormatter{})
		}
	}
	log.SetOutput(logrus.StandardLogger().Writer())
	return nil
}

func InitStdLog(ctx context.Context) error {
	logrus.StandardLogger().SetOutput(os.Stdout)
	log.SetOutput(os.Stdout)
	setLog(logrus.StandardLogger())
	return nil
}

func InitDiscardLog(ctx context.Context) error {
	logrus.StandardLogger().SetOutput(io.Discard)
	log.SetOutput(io.Discard)
	return nil
}
