package log

import (
	"context"
	"github.com/mj37yhyy/gowb/pkg/config"
	"github.com/mj37yhyy/gowb/pkg/constant"
	logger "github.com/sirupsen/logrus"
	"os"
)

func InitLogger(c context.Context) error {
	// 获取配置
	conf := c.Value(constant.ConfigKey).(config.Config)

	// 日志json格式
	if conf.Log.Formatter == "json" {
		// Log as JSON instead of the default ASCII formatter.
		logger.SetFormatter(&logger.JSONFormatter{
			TimestampFormat: "2006-01-02 15:04:05",
		})
	} else if conf.Log.Formatter == "text" {
		logger.SetFormatter(&logger.TextFormatter{
			TimestampFormat:           "2006-01-02 15:04:05",
			ForceColors:               true,
			EnvironmentOverrideColors: true,
		})
	}

	// Output to stdout instead of the default stderr
	// Can be any io.Writer, see below for File example
	logger.SetOutput(os.Stdout)

	// 日志级别
	// Only log the warning severity or above.
	level, err := logger.ParseLevel(conf.Log.Level)
	if err != nil {
		return err
	}
	logger.SetLevel(level)

	// 打印函数与文件
	logger.SetReportCaller(conf.Log.PrintMethod)

	return nil
}
