package main

import (
	"context"
	"de.nilskrau.fishnet.stats/api"
	"de.nilskrau.fishnet.stats/db"
	"de.nilskrau.fishnet.stats/processor"
	"de.nilskrau.fishnet.stats/watcher"
	"gopkg.in/yaml.v3"
	"os"
	"os/signal"
)

type Config struct {
	Db      *db.Config      `yaml:"db"`
	Watcher *watcher.Config `yaml:"watcher"`
	Api     *api.Config     `yaml:"api"`
}

func readConfig() *Config {
	config := &Config{
		Db:      &db.Config{Location: "./data/db.sqlite"},
		Watcher: &watcher.Config{Location: "./data/fishnet.log"},
		Api: &api.Config{
			Port: 4321,
			Host: "0.0.0.0",
		},
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
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	config := readConfig()

	d := db.New(config.Db)
	p := processor.New(d)
	w := watcher.New(config.Watcher)
	a := api.New(config.Api, d)

	w.Watch()
	a.Start(ctx)

	// handle passing of gameid's to processor in background
	go func() {
		for g := range w.NextGameId() {
			p.ProcessGameById(g)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	// Use a buffered channel to avoid missing signals as recommended for signal.Notify
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit

	// cancel the context
	cancel()

}
