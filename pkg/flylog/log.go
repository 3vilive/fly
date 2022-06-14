package flylog

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/spf13/viper"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	Logger         *zap.Logger
	logCtx, cancel = context.WithCancel(context.Background())
)

func init() {
	Logger = zap.NewExample()
}

type LoggerConfig struct {
	Rotate time.Duration
	Level  string
	Logger lumberjack.Logger
}

func InitLog() error {
	config := LoggerConfig{
		Rotate: 24 * time.Hour,
		Level:  "debug",
		Logger: lumberjack.Logger{
			Filename:   "app.log",
			MaxSize:    500, // mega bytes
			MaxBackups: 3,
			MaxAge:     30, // days
		},
	}

	if viper.IsSet("log") {
		fmt.Println("config `log` not set")
		fmt.Printf("viper.Sub(\"log\") = %#v\n", viper.GetStringMap("log"))

		if err := viper.Sub("log").Unmarshal(&config); err != nil {
			fmt.Printf("unmarshal `log` error: %s\n", err.Error())
			return err
		}

		fmt.Printf("config: %#v\n", config)
	}

	// init zap logger
	var zapLogLevel = zap.DebugLevel
	switch config.Level {
	case "debug":
		zapLogLevel = zap.DebugLevel
	case "info":
		zapLogLevel = zap.InfoLevel
	case "warn":
		zapLogLevel = zap.WarnLevel
	case "error":
		zapLogLevel = zap.ErrorLevel
	case "dpanic":
		zapLogLevel = zap.DPanicLevel
	case "panic":
		zapLogLevel = zap.PanicLevel
	case "fatal":
		zapLogLevel = zap.FatalLevel
	}

	w := zapcore.AddSync(&config.Logger)
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
		w,
		zapLogLevel,
	)
	Logger = zap.New(core)

	go func() {
		for {
			select {
			case <-logCtx.Done():
				return
			case <-time.After(config.Rotate):
				if err := config.Logger.Rotate(); err != nil {
					fmt.Fprintf(os.Stderr, "error when rotating log: %v\n", err)
				}
			}
		}
	}()

	return nil
}

func DeinitLog() {
	cancel()

	if err := Logger.Sync(); err != nil {
		fmt.Fprintf(os.Stderr, "Error when closing zap logger: %v\n", err)
	}
}

// proxy all methods
func Sugar() *zap.SugaredLogger                                 { return Logger.Sugar() }
func Named(s string) *zap.Logger                                { return Logger.Named(s) }
func WithOptions(opts ...zap.Option) *zap.Logger                { return Logger.WithOptions(opts...) }
func With(fields ...zap.Field) *zap.Logger                      { return Logger.With(fields...) }
func Check(lvl zapcore.Level, msg string) *zapcore.CheckedEntry { return Logger.Check(lvl, msg) }
func Debug(msg string, fields ...zap.Field)                     { Logger.Debug(msg, fields...) }
func Info(msg string, fields ...zap.Field)                      { Logger.Info(msg, fields...) }
func Warn(msg string, fields ...zap.Field)                      { Logger.Warn(msg, fields...) }
func Error(msg string, fields ...zap.Field)                     { Logger.Error(msg, fields...) }
func DPanic(msg string, fields ...zap.Field)                    { Logger.DPanic(msg, fields...) }
func Panic(msg string, fields ...zap.Field)                     { Logger.Panic(msg, fields...) }
func Fatal(msg string, fields ...zap.Field)                     { Logger.Fatal(msg, fields...) }
func Sync() error                                               { return Logger.Sync() }
func Core() zapcore.Core                                        { return Logger.Core() }
