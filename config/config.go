package config

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Netflix/go-env"
	"github.com/go-ping/ping"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type Configuration struct {
	RPCFirst bool `env:"RPCFirst,default=false"`
	//on redis server
	RedisAddress  string `env:"RedisAddress,required=true"`
	RedisPassword string `env:"RedisPassword"`
	RedisDb       int    `env:"RedisDb,required=true"`

	//AutoPermission should never be true in production
	AutoPermission bool   `env:"AutoPermission"`
	JWTSecret      string `env:"JWTSecret"`
	JwtFieldsKept  string `env:"JwtFieldsKept"`
	MaxBufferSize  int64  `env:"MaxBufferSize,default=10*1024*1024"`
	CORS           string `env:"CORS,default=*"`

	HTTPPort    int64  `env:"HTTPPort,default=80"`
	HTTPPath    string `env:"HTTPPath,default=/"`
	HTTPEnabled bool   `env:"HTTPEnabled,default=false"`

	//ServiceBatchSize is the number of tasks that a service can read from redis at the same time
	ServiceBatchSize int64 `env:"ServiceBatchSize"`
}

var Cfg Configuration = Configuration{}

var Rds *redis.Client = nil

func init() {
	log.Info().Msg("Step1: App Start! load config from OS env")
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	if _, err := env.UnmarshalFromEnviron(&Cfg); err != nil {
		log.Fatal().Err(err).Msg("Load config from env failed")
	}

	if Cfg.JwtFieldsKept != "" {
		Cfg.JwtFieldsKept = strings.ToLower(Cfg.JwtFieldsKept)
	}
	//if redisAddress  is not of format address:port , add default port 6379
	if !strings.Contains(Cfg.RedisAddress, ":") {
		Cfg.RedisAddress = Cfg.RedisAddress + ":6379"
	}
	log.Info().Any("Current Envs:", Cfg).Msg("Load config from env success")

	log.Info().Str("Redis connection Start checking ", Cfg.RedisAddress).Send()
	//ping the address of redisAddress, if failed, print to log
	go pingServer(strings.Split(Cfg.RedisAddress, ":")[0])
	//apply configuration
	redisOption := &redis.Options{
		Addr:         Cfg.RedisAddress,
		Password:     Cfg.RedisPassword, // no password set
		DB:           Cfg.RedisDb,       // use default DB
		PoolSize:     200,
		DialTimeout:  time.Second * 10,
		ReadTimeout:  time.Second * 30,
		WriteTimeout: time.Second * 30,
	}
	rds := redis.NewClient(redisOption)
	//test connection
	if _, err := rds.Ping(context.Background()).Result(); err != nil {
		log.Fatal().Err(err).Msg("Redis connection failed: " + Cfg.RedisAddress)
	}
	log.Info().Str("Redis connection Success", Cfg.RedisAddress)
	timeCmd := rds.Time(context.Background())
	log.Info().Any("Redis server time: ", timeCmd.Val().String()).Send()
	Rds = rds
	log.Info().Msg("Step1: App loaded configuration completed!")

}
func pingServer(domain string) {
	pinger, err := ping.NewPinger(domain)
	if err != nil {
		log.Info().AnErr("ERROR Ping", err).Send()
	}
	pinger.Count = 4
	pinger.Timeout = time.Second * 10

	pinger.OnRecv = func(pkt *ping.Packet) {
		fmt.Printf("Ping Received packet from %s: icmp_seq=%d time=%v\n",
			pkt.IPAddr, pkt.Seq, pkt.Rtt)
	}

	pinger.OnFinish = func(stats *ping.Statistics) {
		fmt.Printf("\n--- %s ping statistics ---\n", stats.Addr)
		fmt.Printf("%d Ping packets transmitted, %d packets received, %v%% packet loss\n",
			stats.PacketsSent, stats.PacketsRecv, stats.PacketLoss)
		fmt.Printf("Ping round-trip min/avg/max/stddev = %v/%v/%v/%v\n",
			stats.MinRtt, stats.AvgRtt, stats.MaxRtt, stats.StdDevRtt)
	}

	fmt.Printf("start pinging %s", domain)
	if err := pinger.Run(); err != nil {
		log.Info().AnErr("ERROR Ping", err).Send()
	}
}
