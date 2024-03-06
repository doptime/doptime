package config

import (
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/rs/zerolog/log"
)

func LoadConfig_FromWeb() {
	var (
		resp      *http.Response
		configUrl string
		err       error
		writer    *os.File
		// save temporary configuration
		_Cfg Configuration = Configuration{}
	)
	//read from env as primary source, and then from the configuration file as secondary source
	if configUrl = os.Getenv("CONFIG_URL"); configUrl == "" {
		configUrl = Cfg.ConfigUrl
	}
	//return if the url is not valid or empty
	if !strings.HasPrefix(strings.ToLower(configUrl), "http") {
		return
	}

	//download from the url and save to the file
	httpClient := &http.Client{Timeout: time.Second * 6}
	if resp, err = httpClient.Get(configUrl); err != nil {
		log.Error().Err(err).Str("Url", configUrl).Msg("LoadConfig_FromWeb failed")
	}
	defer resp.Body.Close()
	if _, err = toml.NewDecoder(resp.Body).Decode(&_Cfg); err != nil {
		log.Error().Err(err).Str("Url", configUrl).Msg("LoadConfig_FromWeb failed")
	}

	//save to the file
	localConfigFile := GetConfigFilePath()("/config.toml")
	if writer, err = os.OpenFile(localConfigFile, os.O_CREATE|os.O_WRONLY, 0644); err != nil {
		log.Error().Err(err).Str("Url", configUrl).Msg("LoadConfig_FromWeb failed")
	}
	defer writer.Close()
	if _, err = io.Copy(writer, resp.Body); err != nil {
		log.Error().Err(err).Str("Url", configUrl).Msg("LoadConfig_FromWeb failed")
	}

	//restore the configUrl, to prevent url drift, which is hard to trace
	_Cfg.ConfigUrl = configUrl
	//apply the configuration
	Cfg = _Cfg

	//reload the configuration every minute
	go func() {
		time.Sleep(time.Minute)
		LoadConfig_FromWeb()
	}()
}
