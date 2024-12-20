package watcher

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/fsnotify/fsnotify"
	"io"
	"os"
	"regexp"
)

type Config struct {
	Location string `yaml:"location"`
}

type Watcher struct {
	listeners chan string
	config    *Config
	watcher   *fsnotify.Watcher
	offset    int64
}

func New(config *Config) *Watcher {
	return &Watcher{
		config: config,
	}
}

func (w *Watcher) getListener() chan string {
	if w.listeners == nil {
		w.listeners = make(chan string)
	}

	return w.listeners
}

func (w *Watcher) NextGameId() <-chan string {
	return w.getListener()
}

func (w *Watcher) sendToListener(gameId string) {
	w.getListener() <- gameId
}

func (w *Watcher) readLogChanges() {
	var err error
	file, err := os.Open(w.config.Location)
	if err != nil {
		return
	}
	defer file.Close()

	if w.offset, err = file.Seek(w.offset, io.SeekCurrent); err != nil {
		return
	}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		if id, err := getGameId(scanner.Bytes()); err != nil {
			continue
		} else {
			w.sendToListener(id)
		}
	}

	if w.offset, err = file.Seek(0, io.SeekCurrent); err != nil {
		return
	}
}

func (w *Watcher) forwardLog() {
	var err error
	file, err := os.Open(w.config.Location)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		if id, err := getGameId(scanner.Bytes()); err != nil {
			continue
		} else {
			w.sendToListener(id)
		}
	}

	if w.offset, err = file.Seek(0, io.SeekCurrent); err != nil {
		panic(err)
	}
}

func (w *Watcher) watch() {
	if w.watcher != nil {
		return
	}

	// read log till end
	w.forwardLog()

	var err error
	w.watcher, err = fsnotify.NewWatcher()
	if err != nil {
		panic(err)
	}

	defer func() {
		w.watcher.Close()
		w.watcher = nil
	}()

	err = w.watcher.Add(w.config.Location)
	if err != nil {
		panic(errors.New("error while adding watcher to syslog"))
	}

	// watch log for changes
	for {
		select {
		case event, ok := <-w.watcher.Events:
			if ok && event.Op&fsnotify.Write == fsnotify.Write {
				w.readLogChanges()
			}
		}
	}
}

func (w *Watcher) Watch() {
	go w.watch()
}

var idRegex = regexp.MustCompile(`https:\/\/lichess\.org\/(.*) finished`)

func getGameId(b []byte) (string, error) {
	matches := idRegex.FindSubmatch(b)
	if len(matches) != 2 {
		return "", fmt.Errorf("no id found")
	}

	return string(matches[1]), nil
}
