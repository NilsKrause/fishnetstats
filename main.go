package main

import (
	"bufio"
	"fmt"
	"github.com/fsnotify/fsnotify"
	"io"
	"log"
	"os"
	"os/signal"
	"os/user"
	"regexp"
)

var fishnetLineRegex *regexp.Regexp
var lichessGameUrlRegex *regexp.Regexp
var gameidRegexS = "\\w{8}"
var homeDir string
var gamesFile = ".fishnet-games"
var lichessUrl = "https://lichess.org/"
var syslogPath = "/var/log/syslog"
var syslogOffset int64
var signals chan os.Signal
var games map[Gameid]*Game

func getHomeDir() {
	usr, err := user.Current()
	if err != nil {
		log.Fatalf(err.Error())
	}
	homeDir = usr.HomeDir
}

func saveGamesToFile() {
	fName := fmt.Sprintf("%s/%s", homeDir, gamesFile)
	fBkName := fmt.Sprintf("%s/%s.bk", homeDir, gamesFile)
	_ = os.Rename(fName, fBkName)
	file, err := os.Create(fName)
	if err != nil {
		panic(err)
	}

	b := make([]byte, 0)
	i := 0
	for _, g := range games {
		if g.Id.String() == "iM76BKY1" {
			fmt.Printf("OK!!")
		}
		gb := g.ToByte()
		if gb == nil || len(gb) <= 0 {
			if g.Id.String() == "iM76BKY1" {
				fmt.Printf("NOPE!!")
			}
			continue
		}
		b = append(b, gb...)
		b = append(b, '\n')
		i++
	}
	fmt.Printf("Iterated over %d games", i)
	n, err := file.Write(b)
	if err != nil {
		panic(err)
	}
	file.Close()

	if n != len(b) {
		fmt.Printf("wrote %d bytes but had %d bytes", n, len(b))
		return
	}

	_ = os.Remove(fBkName)

	fmt.Printf("wrote %d games to %s", len(games), gamesFile)
}

func loadGamesFromFile() {
	file, err := os.Open(fmt.Sprintf("%s/%s", homeDir, gamesFile))
	if err != nil {
		panic(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	i := 0
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) <= 0 || line[0] == '\n' {
			continue
		}

		g := parseGameline(line)

		if g == nil {
			continue
		}

		i++
		addGame(g)
	}

	fmt.Printf("Read %d lines..", i)
}

func loadGamesFromFishnetLog() {
	var err error

	syslog, err := os.Open(syslogPath)
	if err != nil {
		panic(err)
	}

	scanner := bufio.NewScanner(syslog)

	for scanner.Scan() {
		analyzeLogLine(scanner.Text())
	}

	if syslogOffset, err = syslog.Seek(0, io.SeekCurrent); err != nil {
		panic(err)
	}
}

func bToGid(id []byte) Gameid {
	gid := [8]byte{}
	for i := range id {
		gid[i] = id[i]
	}
	return gid
}

func aToGid(id string) Gameid {
	gid := [8]byte{}
	bid := []byte(id)
	for i := range gid {
		gid[i] = bid[i]
	}
	return gid
}

func addGame(g *Game) {
	if _, ok := games[g.Id]; !ok {
		//fmt.Printf("added game from struct %s\n", g.Id)
		games[g.Id] = g
	}
}

func addGameFromId(idS string) {
	id := aToGid(idS)
	if _, ok := games[id]; !ok {
		//fmt.Printf("added game %s\n", idS)
		g := &Game{Id: id}
		g.Initialize()
		games[id] = g
	}
}

func getGameId(line string) string {
	matches := lichessGameUrlRegex.FindAllStringSubmatch(line, -1)
	groupNames := lichessGameUrlRegex.SubexpNames()
	for _, match := range matches {
		for groupIdx, group := range match {
			name := groupNames[groupIdx]
			if name == "id" {
				return group
			}
		}
	}

	return ""
}

func analyzeLogLine(line string) {
	if fishnetLineRegex.MatchString(line) {
		if id := getGameId(line); id != "" {
			addGameFromId(id)
		}
	}
}

func readLogChanges() {
	var err error
	syslog, err := os.Open(syslogPath)
	if err != nil {
		panic(err)
	}

	if syslogOffset, err = syslog.Seek(syslogOffset, io.SeekCurrent); err != nil {
		panic(err)
	}

	scanner := bufio.NewScanner(syslog)
	for scanner.Scan() {
		analyzeLogLine(scanner.Text())
	}

	if syslogOffset, err = syslog.Seek(0, io.SeekCurrent); err != nil {
		panic(err)
	}
}

func followLog() {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer func(watcher *fsnotify.Watcher) {
		err := watcher.Close()
		if err != nil {

		}
	}(watcher)

	done := make(chan bool)
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Op&fsnotify.Write == fsnotify.Write {
					readLogChanges()
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("error:", err)
			}
		}
	}()

	//err = watcher.Add("/var/log/syslog")
	if err != nil {
		log.Fatal(err)
	}
	<-done
}

func waitForSigInt() {
	go func() {
		for range signals {
			saveGamesToFile()
			os.Exit(0)
		}
	}()
}

func init() {
	games = make(map[Gameid]*Game)
	getHomeDir()

	var err error
	fishnetLineRegex, err = regexp.Compile("fishnet-x86_64-unknown-linux-gnu")
	if err != nil {
		panic(err)
	}

	lichessGameUrlRegex, err = regexp.Compile(fmt.Sprintf("%s(?P<id>%s)", lichessUrl, gameidRegexS))
	if err != nil {
		panic(err)
	}

	signals = make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)
	waitForSigInt()
}

func main() {
	loadGamesFromFile()
	loadGamesFromFishnetLog()

	fmt.Printf("Loaded %d already analyzed games.\n", len(games))
	followLog()
}
