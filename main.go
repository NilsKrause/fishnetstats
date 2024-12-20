package main

import (
	"de.nilskrau.fishnet.stats/db"
	"de.nilskrau.fishnet.stats/processor"
	"de.nilskrau.fishnet.stats/watcher"
	"gopkg.in/yaml.v3"
	"os"
)

type Config struct {
	Db      *db.Config      `yaml:"db"`
	Watcher *watcher.Config `yaml:"watcher"`
}

func readConfig() *Config {
	config := &Config{
		Db:      &db.Config{Location: "./db.sqlite"},
		Watcher: &watcher.Config{Location: "./fishnet.log"},
	}

	configDir := "./config.yaml"
	if len(os.Args) > 2 && os.Args[1] == "-c" {
		configDir = os.Args[2]
	}

	if data, err := os.ReadFile(configDir); err == nil {
		_ = yaml.Unmarshal(data, &config)
	}

	return config
}

func main() {
	config := readConfig()

	d := db.New(config.Db)
	p := processor.New(d)
	w := watcher.New(config.Watcher)

	w.Watch()

	for g := range w.NextGameId() {
		p.ProcessGameById(g)
	}
}
