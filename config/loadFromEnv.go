package config

import (
	"encoding/json"
	"os"
	"strings"

	"github.com/rs/zerolog/log"
)

func LoadConfig_FromEnv() (err error) {
	var envMap = map[string]string{}

	for _, env := range os.Environ() {
		kvs := strings.SplitN(env, "=", 2)
		if len(kvs) == 2 && len(kvs[0]) > 0 && len(kvs[1]) > 0 {
			envMap[kvs[0]] = kvs[1]
		}
	}
	//load redis items
	for key, val := range envMap {
		var rdsCfg = ConfigRedis{}
		//if it is not in the format of Redis_*, then skip
		if strings.Index(key, "REDIS") != 0 || len(val) <= 6 || key[5] != '_' {
			continue
		}

		//read in value, if it is not in the format of {Username,Password,Host,Port,DB}, then skip
		if val = strings.TrimSpace(val); val[0] != '{' || val[len(val)-1] != '}' {
			continue
		}

		if err := json.Unmarshal([]byte(val), &rdsCfg); err != nil {
			correctFormat := "{Name,Username,Password,Host,Port,DB},{Name,Username,Password,Host,Port,DB}"
			log.Fatal().Err(err).Str("redis key", key).Str("redisEnv", val).Msg("Step1.1.2 Load Env/Redis failed, correct format: " + correctFormat)
		}
		//read redis name from env key
		rdsCfg.Name = key[6:]
		Cfg.Redis = append(Cfg.Redis, rdsCfg)
	}
	// Load and parse HTTP config
	if httpEnv, ok := envMap["HTTP"]; ok && len(httpEnv) > 0 {
		if err := json.Unmarshal([]byte(httpEnv), &Cfg.Http); err != nil {
			log.Fatal().Err(err).Str("httpEnv", httpEnv).Msg("Step1.1.2 Load Env/Http failed")
		}
	}

	// Load LogLevel
	if settingEnv, ok := envMap["SETTINGS"]; ok && len(settingEnv) > 0 {
		if err := json.Unmarshal([]byte(settingEnv), &Cfg.Settings); err != nil {
			log.Fatal().Err(err).Str("settingEnv", settingEnv).Msg("Step1.1.2 Load Env/Http failed")
		}
	}
	return nil
}
