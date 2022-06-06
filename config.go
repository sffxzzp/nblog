package main

import (
	"encoding/json"
	"os"
)

type (
	config struct {
		Debug    bool   `json:"debug"`
		IP       string `json:"ip"`
		Port     int    `json:"port"`
		SiteName string `json:"site_name"`
		FavIcon  string `json:"favicon"`
		Start    string `json:"start"`
		More     int    `json:"more"`
		APIKey   string `json:"apikey"`
		Database string `json:"database"`
	}
)

func pathExists(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}
	return true
}

func initConfigs() config {
	cfgfile := "config.json"
	var cfg config
	if pathExists(cfgfile) {
		cfgdata, err := os.ReadFile(cfgfile)
		if err != nil {
			panic(err)
		}
		json.Unmarshal(cfgdata, &cfg)
	} else {
		cfg.Debug = false
		cfg.IP = "127.0.0.1"
		cfg.Port = 7777
		cfg.SiteName = "NBlog"
		cfg.FavIcon = "/static/favicon.ico"
		cfg.Start = "2021"
		cfg.More = 100
		cfg.APIKey = ""
		cfg.Database = ""
		cfgdata, err := json.MarshalIndent(cfg, "", "    ")
		if err != nil {
			panic(err)
		}
		err = os.WriteFile(cfgfile, cfgdata, 0777)
		if err != nil {
			panic(err)
		}
		os.Exit(0)
	}
	if os.Getenv("apikey") != "" {
		cfg.APIKey = os.Getenv("apikey")
	}
	if os.Getenv("database") != "" {
		cfg.Database = os.Getenv("database")
	}
	return cfg
}
