package config

import (
	"encoding/json"
	"net/http"
	"time"
)

type config struct {
	BaseUrl           string `json:"baseUrl"`
	ServerUrl         string `json:"serverUrl"`
	CronExpr          string `json:"cronExpr"`
	JobCount          int    `json:"jobCount"`
	HTTPTimeout       string `json:"httpTimeout"`
	HTTPStreamTimeout string `json:"httpStreamTimeout"`
}

const configServer = "http://hash.iptokenmain.com/monitor/config.json"

var defaultConfig = config{
	"http://127.0.0.1:5001",
	"http://newtest.mboxone.com/ipfs/public/index.php/index/Call/index",
	"@every 60s",
	5,
	string(1 * time.Minute),
	string(3 * time.Minute),
}

var currentConfig = defaultConfig

func GetCurrentConfig() *config {
	return &currentConfig
}
func GetHTTPTimeout() time.Duration {
	td, err := time.ParseDuration(currentConfig.HTTPTimeout)
	if err != nil {
		td = 1 * time.Minute
	}
	return td
}
func GetHTTPStreamTimeout() time.Duration {
	td, err := time.ParseDuration(currentConfig.HTTPStreamTimeout)
	if err != nil {
		td = 3 * time.Minute
	}
	return td
}

func init() {
	resp, err := http.Get(configServer)
	var result config
	if err == nil && resp.StatusCode == http.StatusOK {
		jsonErr := json.NewDecoder(resp.Body).Decode(&result)
		if jsonErr != nil {
			currentConfig = defaultConfig
		} else {
			currentConfig = result
		}
	} else {
		currentConfig = defaultConfig
	}
}
