package config

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"path/filepath"
)

const (
	configPath = "./config.json"
)

type (
	Config struct {
		Mysql         Mysql  `json:"mysql"`
		TelegramToken string `json:"telegram_token"`
		Node          string `json:"node"`
	}
	Mysql struct {
		Host     string `json:"host"`
		Port     string `json:"port"`
		DB       string `json:"db"`
		User     string `json:"user"`
		Password string `json:"password"`
	}
)

func GetConfig() Config {
	path, _ := filepath.Abs(configPath)
	file, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatalln("Invalid config path : "+configPath, err)
	}
	var config Config
	err = json.Unmarshal(file, &config)
	if err != nil {
		log.Fatalln("Failed unmarshal config ", err)
	}
	return config
}
