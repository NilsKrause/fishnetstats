package main

import (
	"de.nilskrau.fishnet.stats/db"
	"de.nilskrau.fishnet.stats/processor"
	"de.nilskrau.fishnet.stats/watcher"
)

type Config struct {
	Db      *db.Config
	Watcher *watcher.Config
}

func main() {
	config := &Config{
		Db:      &db.Config{Location: "./db.sqlite"},
		Watcher: &watcher.Config{Location: "./fishnet.log"},
	}
	d := db.New(config.Db)
	p := processor.New(d)
	w := watcher.New(config.Watcher)

	w.Watch()

	for g := range w.NextGameId() {
		p.ProcessGameById(g)
	}
}
