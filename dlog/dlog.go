package dlog

import (
	"context"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
)

type dWriter struct {
}

var RdsClientToLog *redis.Client = nil

func (dr dWriter) WriteLevel(level zerolog.Level, p []byte) (n int, err error) {

	key := "doptimelog:" + getMachineName()
	if RdsClientToLog != nil {
		RdsClientToLog.ZAdd(context.Background(), key, redis.Z{Score: float64(time.Now().UnixNano()), Member: string(p)})
	}
	return dr.Write(p)
}
func (dr dWriter) Write(p []byte) (n int, err error) {
	_, err = os.Stdout.Write(p)
	if err != nil {
		return 0, err
	}
	return len(p), nil
}

var levelWriter dWriter = dWriter{}

var Logger = zerolog.New(levelWriter)

func Debug() *zerolog.Event {
	return Logger.Debug()
}
func Info() *zerolog.Event {
	return Logger.Info()
}
func Warn() *zerolog.Event {
	return Logger.Warn()
}
func Error() *zerolog.Event {
	return Logger.Error()
}
func Fatal() *zerolog.Event {
	return Logger.Fatal()
}
func Panic() *zerolog.Event {
	return Logger.Panic()
}
func Log() *zerolog.Event {
	return Logger.Log()
}
