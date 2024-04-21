package dlog

import (
	"context"
	"math/big"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/cespare/xxhash/v2"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
)

type dWriter struct {
}

var RdsClientToLog *redis.Client = nil

var SavedText = map[string]bool{}
var saveLogTextMutex = &sync.Mutex{} // 使用 sync.Mutex 替代通道
var keyLogName = "doptimelog:" + getMachineName()
var keyLogTextName = "doptimelog:" + getMachineName() + ":text"

func (dr dWriter) WriteLevel(level zerolog.Level, p []byte) (n int, err error) {
	var (
		ok bool
	)

	if RdsClientToLog != nil {
		redisPipeline := RdsClientToLog.Pipeline()
		now := time.Now()
		timeStr := strconv.FormatInt(now.UnixMicro(), 10)
		xxhash64 := big.NewInt(int64(xxhash.Sum64(p) & 0x7FFFFFFFFFFFFFFF)).Text(62)

		redisPipeline.LPush(context.Background(), keyLogName, timeStr+":"+xxhash64)
		//keep 32768 log items only
		redisPipeline.LTrim(context.Background(), keyLogName, -32768, -1)
		//lock saveLogTextMutextLock to read and write SavedText
		saveLogTextMutex.Lock()
		if _, ok = SavedText[xxhash64]; !ok {
			redisPipeline.HSet(context.Background(), keyLogTextName, xxhash64, p)
			SavedText[xxhash64] = true
		}
		saveLogTextMutex.Unlock()

		redisPipeline.Exec(context.Background())
	}
	return dr.Write(p)
}

func (dr dWriter) Write(p []byte) (n int, err error) {
	os.Stdout.Write([]byte(time.Now().Format("2006-01-02 15:04:05") + " "))
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
