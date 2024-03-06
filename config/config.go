package config

import (
	"context"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/go-ping/ping"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type ConfigHttp struct {
	CORES  string `env:"CORES,default=*"`
	Port   int64  `env:"Port,default=80"`
	Path   string `env:"Path,default=/"`
	Enable bool   `env:"Enable,default=false"`
	//MaxBufferSize is the max size of a task in bytes, default 10M
	MaxBufferSize int64 `env:"MaxBufferSize,default=10485760"`
}
type ConfigRedis struct {
	Name     string
	Username string `env:"Username"`
	Password string `env:"Password"`
	Host     string `env:"Host,required=true"`
	Port     int64  `env:"Port,required=true"`
	DB       int64  `env:"DB,required=true"`
}
type ConfigJWT struct {
	Secret string `env:"Secret"`
	Fields string `env:"Fields"`
}
type ConfigAPI struct {
	//ServiceBatchSize is the number of tasks that a service can read from redis at the same time
	ServiceBatchSize int64 `env:"ServiceBatchSize,default=64"`
}
type ConfigData struct {
	//AutoAuth should never be true in production
	AutoAuth bool `env:"AutoAuth,default=false"`
}
type Configuration struct {
	ConfigUrl string
	//redis server, format: username:password@address:port/db
	Redis []*ConfigRedis
	Jwt   ConfigJWT
	Http  ConfigHttp
	Api   ConfigAPI
	Data  ConfigData
	//{"DebugLevel": 0,"InfoLevel": 1,"WarnLevel": 2,"ErrorLevel": 3,"FatalLevel": 4,"PanicLevel": 5,"NoLevel": 6,"Disabled": 7	  }
	LogLevel int8 `env:"LogLevel,default=1"`
}

// set default values
var Cfg Configuration = Configuration{
	ConfigUrl: "",
	Redis:     []*ConfigRedis{},
	Jwt:       ConfigJWT{Secret: "", Fields: "*"},
	Http:      ConfigHttp{CORES: "*", Port: 80, Path: "/", Enable: false, MaxBufferSize: 10485760},
	Api:       ConfigAPI{ServiceBatchSize: 64},
	Data:      ConfigData{AutoAuth: false},
	LogLevel:  1,
}

var Rds map[string]*redis.Client = map[string]*redis.Client{}

func GetRdsClientByName(name string) (rds *redis.Client, err error) {
	var (
		ok bool
	)
	if rds, ok = Rds[name]; !ok {
		err = fmt.Errorf("redis client with name %s not found", name)
		return nil, err
	}

	return rds, nil
}

func init() {
	log.Info().Msg("Step1.0: App Start! load config from OS env")
	//step1: load config from file
	LoadConfig_FromFile()
	//step2: load config from env. this will overwrite the config from file
	LoadConfig_FromEnv()
	//step3: load config from web. this will overwrite the config from env.
	//warning local config will be overwritten by the config from web, to prevent falldown of config from web.
	LoadConfig_FromWeb()

	if _, ok := Rds[""]; !ok {
		log.Info().Msg("Step1.0 ERROR LoadConfig")
		log.Info().Msg("goflow data & api will no be able to be used. please check your env and restart the app if you want to use it")
		return
	}
	zerolog.SetGlobalLevel(zerolog.Level(Cfg.LogLevel))

	if Cfg.Jwt.Fields != "" {
		Cfg.Jwt.Fields = strings.ToLower(Cfg.Jwt.Fields)
	}
	log.Info().Any("Step1.1 Current Envs:", Cfg).Msg("Load config from env success")

	log.Info().Str("Step1.2 Checking Redis", "Start").Send()

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
			log.Fatal().Err(err).Any("Step1.3 Redis server not rechable", rdsCfg.Host).Send()
			return //if redis server is not valid, exit
		}
		//save to the list
		log.Info().Str("Step1.3 Redis Load ", "Success").Any("RedisUsername", rdsCfg.Username).Any("RedisPassword", rdsCfg.Password).Any("RedisHost", rdsCfg.Host).Any("RedisPort", rdsCfg.Port).Send()
		Rds[rdsCfg.Name] = rdsClient
		timeCmd := rdsClient.Time(context.Background())
		log.Info().Any("Step1.4 Redis server time: ", timeCmd.Val().String()).Send()
		//ping the address of redisAddress, if failed, print to log
		pingServer(rdsCfg.Host)

	}

	log.Info().Msg("Step1.E: App loaded done")

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
		log.Info().AnErr("Step1.5 ERROR NewPinger", err).Send()
	}
	pinger.Count = 4
	pinger.Timeout = time.Second * 10
	pinger.OnRecv = func(pkt *ping.Packet) {}

	pinger.OnFinish = func(stats *ping.Statistics) {
		// fmt.Printf("\n--- %s ping statistics ---\n", stats.Addr)
		log.Info().Str("Step1.5 Ping ", fmt.Sprintf("--- %s ping statistics ---", stats.Addr)).Send()
		// fmt.Printf("%d Ping packets transmitted, %d packets received, %v%% packet loss\n",
		// 	stats.PacketsSent, stats.PacketsRecv, stats.PacketLoss)
		log.Info().Str("Step1.5 Ping", fmt.Sprintf("%d/%d/%v%%", stats.PacketsSent, stats.PacketsRecv, stats.PacketLoss)).Send()

		// fmt.Printf("Ping round-trip min/avg/max/stddev = %v/%v/%v/%v\n",
		// 	stats.MinRtt, stats.AvgRtt, stats.MaxRtt, stats.StdDevRtt)
		log.Info().Str("Step1.5 Ping", fmt.Sprintf("%v/%v/%v/%v", stats.MinRtt, stats.AvgRtt, stats.MaxRtt, stats.StdDevRtt)).Send()
	}
	go func() {
		if err := pinger.Run(); err != nil {
			log.Info().AnErr("Step1.5 ERROR Ping", err).Send()
		}
	}()
}
