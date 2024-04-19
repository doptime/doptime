package config

import (
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"github.com/doptime/doptime/dlog"
)

func GetConfigFilePath() func(filename string) string {
	var (
		tomlFilePath string
		err          error
	)
	//tomlPath is same path as the binary
	if tomlFilePath, err = os.Executable(); err != nil {
		tomlFilePath = ""
	} else {
		tomlFilePath = filepath.Dir(tomlFilePath)
	}
	return func(filename string) string {
		return tomlFilePath + filename
	}

}

// step1: load config from file
func LoadConfig_FromFile() {
	var (
		TomlPath = GetConfigFilePath()
	)
	//tomlPath is same path as the binary
	if TomlPath("") == "" {
		return
	}

	var (
		tomlFile, demoConfigFile string = TomlPath("/config.toml"), TomlPath("/config.demo.toml")
		err                      error
		writer                   *os.File
	)
	//return if success load config from file
	if _, err := toml.DecodeFile(tomlFile, &Cfg); err == nil {
		dlog.Info().Str("filename", tomlFile).Msg("LoadConfigFromFile success")
		return
	}

	//if toml file exist, but with bad format, then return
	if _, err := os.Stat(tomlFile); os.IsExist(err) {
		dlog.Error().Err(err).Str("filename", tomlFile).Msg("LoadConfigFromFile failed, see example format in config.demo.toml")
	}

	defer func() {
		Cfg.Redis = []*ConfigRedis{}
	}()

	// create default config file as demo
	Cfg.Redis = append(Cfg.Redis, &ConfigRedis{Name: "redis_server_demo", Username: "doptime", Password: "yourpasswordhere", Host: "drangonflydb.local", Port: 6379, DB: 0})

	//open file to write
	if writer, err = os.OpenFile(demoConfigFile, os.O_CREATE|os.O_WRONLY, 0644); err != nil {
		dlog.Error().Err(err).Str("filename", demoConfigFile).Msg("save demo toml config file failed")
		return
	}
	defer writer.Close()
	//encode toml
	encoder := toml.NewEncoder(writer)
	if err := encoder.Encode(Cfg); err != nil {
		dlog.Error().Err(err).Str("filename", demoConfigFile).Msg("toml Encode demo toml config file failed")
	}
}
