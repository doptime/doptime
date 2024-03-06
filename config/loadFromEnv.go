package config

import (
	"encoding/json"
	"os"
	"strconv"
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
		var rdsCfg = &ConfigRedis{}
		//if it is Redis, then change it to Redis_default
		if key == "Redis" {
			key = "Redis_default"
		}
		//if it is not in the format of Redis_*, then skip
		if strings.Index(key, "Redis") != 0 || len(val) <= 6 || key[5] != '_' {
			continue
		}

		//read in value, if it is not in the format of {Username,Password,Host,Port,DB}, then skip
		if val = strings.TrimSpace(val); val[0] != '{' || val[len(val)-1] != '}' {
			continue
		}

		if err := json.Unmarshal([]byte(val), &rdsCfg); err != nil {
			correctFormat := "{Name,Username,Password,Host,Port,DB},{Name,Username,Password,Host,Port,DB}"
			log.Fatal().Err(err).Str("redis key", key).Str("redisEnv", val).Msg("Step1.0 Load Env/Redis failed, correct format: " + correctFormat)
		}
		// read in the name of the redis server
		//if the name is default, then set it to empty
		if rdsCfg.Name = key[6:]; rdsCfg.Name == "default" {
			rdsCfg.Name = ""
		}
		Cfg.Redis = append(Cfg.Redis, rdsCfg)
	}

	// Load and parse JWT config
	if jwtEnv, ok := envMap["Jwt"]; ok && jwtEnv != "" {
		if err := json.Unmarshal([]byte(jwtEnv), &Cfg.Jwt); err != nil {
			log.Fatal().Err(err).Str("jwtEnv", jwtEnv).Msg("Step1.0 Load Env/Jwt failed")
		}
	}

	// Load and parse HTTP config
	Cfg.Http.Enable, Cfg.Http.Path, Cfg.Http.CORES = true, "/", "*"
	if httpEnv, ok := envMap["Http"]; ok && len(httpEnv) > 0 {
		if err := json.Unmarshal([]byte(httpEnv), &Cfg.Http); err != nil {
			log.Fatal().Err(err).Str("httpEnv", httpEnv).Msg("Step1.0 Load Env/Http failed")
		}
	}

	// Load and parse API config
	if apiEnv, ok := envMap["Api"]; ok && apiEnv != "" {
		if err := json.Unmarshal([]byte(apiEnv), &Cfg.Api); err != nil {
			log.Fatal().Err(err).Str("apiEnv", apiEnv).Msg("Step1.0 Load Env/Api failed")
		}
	}
	if dataEnv, ok := envMap["Data"]; ok && dataEnv != "" {
		if err := json.Unmarshal([]byte(dataEnv), &Cfg.Data); err != nil {
			log.Fatal().Err(err).Str("dataEnv", dataEnv).Msg("Step1.0 Load Env/data env failed")
		}
	}

	// Load LogLevel
	if logLevelEnv, ok := envMap["LogLevel"]; ok && len(logLevelEnv) > 0 {
		if logLevel, err := strconv.ParseInt(logLevelEnv, 10, 8); err == nil {
			Cfg.LogLevel = int8(logLevel)
		}
	}
	return nil
}
