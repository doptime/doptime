package config

import (
	"bytes"
	"context"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/doptime/doptime/dlog"
	"github.com/go-ping/ping"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
)

type ConfigHttp struct {
	CORES string
	Port  int64  `env:"Port,default=80"`
	Path  string `env:"Path,default=/"`
	//MaxBufferSize is the max size of a task in bytes, default 10M
	MaxBufferSize int64  `env:"MaxBufferSize,default=10485760"`
	JwtSecret     string `env:"Secret"`
	//AutoAuth should never be true in production
	AutoAuth bool `env:"AutoAuth,default=false"`
}
type ConfigRedis struct {
	Name     string
	Username string `env:"Username"`
	Password string `env:"Password"`
	Host     string `env:"Host,required=true"`
	Port     int64  `env:"Port,required=true"`
	DB       int64  `env:"DB,required=true"`
}

// the http rpc server
type ApiSourceHttp struct {
	Name    string
	UrlBase string
	Jwt     string
}
type ConfigSettings struct {
	//{"DebugLevel": 0,"InfoLevel": 1,"WarnLevel": 2,"ErrorLevel": 3,"FatalLevel": 4,"PanicLevel": 5,"NoLevel": 6,"Disabled": 7	  }
	LogLevel int8
	//super user token, this is used to bypass the security check in data access
	//SUToken is designed to allow debugging in production environment without  change the permission table permanently
	SUToken string
}

type Configuration struct {
	ConfigUrl string
	Http      ConfigHttp
	//redis server, format: username:password@address:port/db
	Redis    []*ConfigRedis
	HttpRPC  []*ApiSourceHttp
	Settings ConfigSettings
}

// ServiceBatchSize is the number of tasks that a service can read from redis at the same time
var ServiceBatchSize int64 = 64

func (c Configuration) String() string {
	var (
		c1  Configuration
		buf bytes.Buffer
	)
	HideCharsButLat4 := func(s string) string {
		if len(s) <= 4 {
			return strings.Repeat("*", len(s))
		}
		return strings.Repeat("*", len(s)-4) + s[len(s)-4:]
	}
	//use gob to deep copy, to prevent error modification of the original secret
	if err := gob.NewEncoder(&buf).Encode(c); err != nil {
		return "error: " + err.Error() + " when encoding config to gob string"
	}
	gob.NewDecoder(&buf).Decode(&c1)
	//hide the secret , but leaving last 4 chars
	c1.Http.JwtSecret = HideCharsButLat4(c1.Http.JwtSecret)
	//hide the password, but leaving last 4 chars
	for _, rds := range c1.Redis {
		rds.Password = HideCharsButLat4(rds.Password)
	}
	for _, rpc := range c1.HttpRPC {
		rpc.Jwt = HideCharsButLat4(rpc.Jwt)
	}
	c1.Settings.SUToken = HideCharsButLat4(c1.Settings.SUToken)
	//convert c1 to json string
	jsonstr, _ := json.Marshal(c1)
	return string(jsonstr)
}

// set default values
var Cfg Configuration = Configuration{
	ConfigUrl: "",
	Redis:     []*ConfigRedis{},
	Http:      ConfigHttp{CORES: "*", Port: 80, Path: "/", MaxBufferSize: 10485760},
	HttpRPC:   []*ApiSourceHttp{{Name: "doptime", UrlBase: "https://api.doptime.com", Jwt: ""}},
	Settings:  ConfigSettings{LogLevel: 1},
}

var ErrNoSuchRedisDB = fmt.Errorf("no such redis db")
var Rds map[string]*redis.Client = map[string]*redis.Client{}
var ErrNoSuchRpcServer = fmt.Errorf("no such http Rpc Server")
var HttpRpc map[string]*ApiSourceHttp = map[string]*ApiSourceHttp{}

func init() {
	dlog.Info().Msg("Step1.0: App Start! load config from OS env")
	//step1: load config from file
	LoadConfig_FromFile()
	dlog.Info().Str("Step1.1.1 Current config after apply config.toml", Cfg.String()).Send()
	//step2: load config from env. this will overwrite the config from file
	//step3: load config from web. this will overwrite the config from env.
	//warning local config will be overwritten by the config from web, to prevent falldown of config from web.
	LoadConfig_FromWeb()
	dlog.Info().Str("Step1.1.2 Current config after apply web config toml", Cfg.String()).Send()

	zerolog.SetGlobalLevel(zerolog.Level(Cfg.Settings.LogLevel))

	dlog.Info().Str("Step1.2 Checking Redis", "Start").Send()

	for _, rdsCfg := range Cfg.Redis {
		//apply configuration
		redisOption := &redis.Options{
			Addr:         rdsCfg.Host + ":" + strconv.Itoa(int(rdsCfg.Port)),
			Username:     rdsCfg.Username,
			Password:     rdsCfg.Password, // no password set
			DB:           int(rdsCfg.DB),  // use default DB
			PoolSize:     200,
			DialTimeout:  time.Second * 10,
			ReadTimeout:  -1,
			WriteTimeout: time.Second * 300,
		}
		rdsClient := redis.NewClient(redisOption)
		//test connection
		if _, err := rdsClient.Ping(context.Background()).Result(); err != nil {
			dlog.Fatal().Err(err).Any("Step1.3 Redis server ping error", rdsCfg.Host).Send()
			return //if redis server is not valid, exit
		}
		//save to the list
		dlog.Info().Str("Step1.3 Redis Load ", "Success").Any("RedisUsername", rdsCfg.Username).Any("RedisHost", rdsCfg.Host).Any("RedisPort", rdsCfg.Port).Send()
		Rds[rdsCfg.Name] = rdsClient
		timeCmd := rdsClient.Time(context.Background())
		dlog.Info().Any("Step1.4 Redis server time: ", timeCmd.Val().String()).Send()
		//ping the address of redisAddress, if failed, print to log
		pingServer(rdsCfg.Host)

	}
	//check if default redis is set
	if _rds, ok := Rds["default"]; !ok {
		dlog.Warn().Msg("Step1.0 \"default\" redis server missing in Configuration. RPC will can not be received. Please ensure this is what your want")
		return
	} else {
		Rds[""] = _rds
		dlog.RdsClientToLog = _rds
	}
	for _, rpc := range Cfg.HttpRPC {
		HttpRpc[rpc.Name] = rpc
	}
	dlog.Info().Msg("Step1.E: App loaded done")

}

var pingTaskServers = []string{}

func pingServer(domain string) {
	var (
		pinger *ping.Pinger
		err    error
	)
	if slices.Index(pingTaskServers, domain) != -1 {
		return
	}
	pingTaskServers = append(pingTaskServers, domain)

	if pinger, err = ping.NewPinger(domain); err != nil {
		dlog.Info().AnErr("Step1.5 ERROR NewPinger", err).Send()
	}
	pinger.Count = 4
	pinger.Timeout = time.Second * 10
	pinger.OnRecv = func(pkt *ping.Packet) {}

	pinger.OnFinish = func(stats *ping.Statistics) {
		// fmt.Printf("\n--- %s ping statistics ---\n", stats.Addr)
		dlog.Info().Str("Step1.5 Ping ", fmt.Sprintf("--- %s ping statistics ---", stats.Addr)).Send()
		// fmt.Printf("%d Ping packets transmitted, %d packets received, %v%% packet loss\n",
		// 	stats.PacketsSent, stats.PacketsRecv, stats.PacketLoss)
		dlog.Info().Str("Step1.5 Ping", fmt.Sprintf("%d/%d/%v%%", stats.PacketsSent, stats.PacketsRecv, stats.PacketLoss)).Send()

		// fmt.Printf("Ping round-trip min/avg/max/stddev = %v/%v/%v/%v\n",
		// 	stats.MinRtt, stats.AvgRtt, stats.MaxRtt, stats.StdDevRtt)
		dlog.Info().Str("Step1.5 Ping", fmt.Sprintf("%v/%v/%v/%v", stats.MinRtt, stats.AvgRtt, stats.MaxRtt, stats.StdDevRtt)).Send()
	}
	go func() {
		if err := pinger.Run(); err != nil {
			dlog.Info().AnErr("Step1.5 ERROR Ping", err).Send()
		}
	}()
}
