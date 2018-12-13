package config

import (
	"encoding/json"
	"fmt"
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
const DebugConfigServer = "http://newtest.mboxone.com/monitor/config.json"

var defaultConfig = config{
	"http://127.0.0.1:5001",
	"http://newtest.mboxone.com/ipfs/public/index.php/index/Call/index",
	"@every 5s",
	2,
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

func GetMaxTaskNum() int {
	return currentConfig.JobCount
}

func GetServerConfig(debug bool, configUrl string) {
	var url = configServer
	if debug {
		if configUrl != "" {
			url = configUrl
		} else {
			url = DebugConfigServer
		}
		fmt.Println("--------------------start debug mode---------------------")
	}
	resp, err := http.Get(url)
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

func init() {
}
