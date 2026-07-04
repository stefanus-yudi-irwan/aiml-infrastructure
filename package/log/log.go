package log

import (
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var logger *zap.Logger

func init() {
	SetProductionLogger()
}

func Logger() *zap.Logger {
	return logger
}

func Info(msg string, fields ...zapcore.Field) {
	Logger().Info(msg, fields...)
}

func Error(msg string, fields ...zapcore.Fields) {
	Logger().Error(msg, fields...)
}

// SetProductionLogger set current logger in production mode.
func SetProductionLogger(outputPaths ...string) {
	for _, outputPath := range outputPaths {
		err := os.MkdirAll(filepath.Dir(outputPath), os.ModePerm)
		if err != nil {
			panic(err)
		}
	}
	cfg := zap.NewProductionConfig()
	cfg.EncoderConfig.TimeKey = "timestamp"
	cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	cfg.OutputPaths = append(cfg.OutputPaths, outputPaths...)
	var err error
	logger, err = cfg.Build()
	if err != nil {
		panic(err)
	}
}

// SetDevelopmentLogger set current logger in development mode.
func SetDevelopmentLogger(outputPaths ...string) {
	for _, outputPath := range outputPaths {
		err := os.MkdirAll(filepath.Dir(outputPath), os.ModePerm)
		if err != nil {
			panic(err)
		}
	}
	cfg := zap.NewDevelopmentConfig()
	cfg.OutputPaths = append(cfg.OutputPaths, outputPaths...)
	var err error
	logger, err = cfg.Build()
	if err != nil {
		panic(err)
	}
}

func CloseLogger() {
	cfg := zap.NewProductionConfig()
	cfg.Level = zap.NewAtomicLevelAt(zap.FatalLevel)
	var err error
	logger, err = cfg.Build()
	if err != nil {
		panic(err)
	}
}

func RedactDBURL(rawURL string) string {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}
	username := parsed.User.Username()
	password, _ := parsed.User.Password()
	parsed.User = url.UserPassword(strings.Repeat("x", len(username)), strings.Repeat("x", len(password)))
	return parsed.String()
}
