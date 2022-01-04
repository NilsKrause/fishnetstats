package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
)

type Game struct {
	Id          Gameid       `json:"-"`
	White       *Player      `json:"white,omitempty"`
	Black       *Player      `json:"black,omitempty"`
	Link        string       `json:"-"`
	Result      *Result      `json:"result,omitempty"`
	Termination string       `json:"termination,omitempty"`
	Format      *Timecontrol `json:"format,omitempty"`

	Initialized bool `json:"initialized,omitempty"`
}

func getGameIdFromBody(body string) (Gameid, error) {
	lines := strings.Split(body, "\n")
	for _, line := range lines {
		if len(line) <= 1 {
			continue
		}
		if line[0] != '[' {
			continue
		}

		line = strings.Trim(line, "[]")
		s := strings.Split(line, " ")

		value := strings.Trim(s[1], "\"")
		if s[0] == "Site" {
			elems := strings.Split(value, "/")
			id := elems[len(elems)-1]
			if len(id) != 8 {
				return Gameid{}, errors.New("no valid gameid found")
			}
			return bToGid([]byte(id)), nil
		}
	}

	return Gameid{}, errors.New("no gameid found")
}

func (g *Game) playerUpdated() {
	if g.IsInitialized() {
		stats.addGame(g)
	}
}

func (g *Game) Initialize() {
	if g.IsInitialized() {
		return
	}

	if !g.Initialized {
		Request(g)
		return
	}

	if g.White != nil && !g.White.IsInitialized() {
		g.White.Initialize()
	} else if g.White == nil {
		gs, err := json.Marshal(g)
		if err != nil {
			gs = []byte("[unknown]")
		}
		fmt.Printf("weirdly have a nil player object for white: %s\n", gs)
	}

	if !g.Black.IsInitialized() {
		g.Black.Initialize()
	} else if g.Black == nil {
		gs, err := json.Marshal(g)
		if err != nil {
			gs = []byte("[unknown]")
		}
		fmt.Printf("weirdly have a nil player object for black: %s\n", gs)
	}
}

func (g *Game) HasId(id interface{}) bool {
	nid := id.([8]byte)

	if g.Id == nid {
		return true
	}

	return false
}

func (g *Game) GetType() GettableType {
	return GameT
}

func (g *Game) IsInitialized() bool {
	return g.Initialized && g.White != nil && g.White.IsInitialized() && g.Black != nil && g.Black.IsInitialized()
}

func (g *Game) GetBodyString() string {
	return g.Id.String()
}

func (g *Game) GetBody() io.Reader {
	return strings.NewReader(g.Id.String())
}

func (g *Game) GetUrl() string {
	return "https://lichess.org/games/export/_ids"
}

func (g *Game) ParseResponseBody(body string) {
	lines := strings.Split(body, "\n")
	for _, line := range lines {
		if len(line) <= 1 {
			continue
		}
		if line[0] != '[' {
			continue
		}

		line = strings.Trim(line, "[]")
		s := strings.Split(line, " ")

		value := strings.Trim(s[1], "\"")
		switch s[0] {
		case "White":
			g.White = NewPlayer(value, g)
			g.White.Initialize()
			break
		case "Black":
			g.Black = NewPlayer(value, g)
			g.Black.Initialize()
			break
		case "Result":
			g.Result = aToRes(value)
			break
		case "Termination":
			g.Termination = value
			break
		case "TimeControl":
			g.Format = aToTC(value)
		}
	}

	g.Initialized = true
}

func (g *Game) ParseHttpResponse(r *http.Response) {
	b := make([]byte, 0)
	b, err := ioutil.ReadAll(r.Body)
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(r.Body)

	if err != nil {
		fmt.Printf("error while reading game response body %s\n", err.Error())
		return
	}

	if len(b) <= 0 {
		fmt.Printf("error while reading game response body, length less or equal zero.\n")
		return
	}

	g.ParseResponseBody(string(b))
}

func parseGameline(b []byte) *Game {
	game := &Game{}
	err := json.Unmarshal(b[9:], game)
	if err != nil {
		fmt.Printf("Error (%s) parsing gmaestring: %s\n", err.Error(), string(b))
		return nil
	}

	game.Id = bToGid(b[:8])
	game.Link = fmt.Sprintf("%s%s", lichessUrl, game.Id)
	game.Initialize()

	return game
}

func (g *Game) ToByte() []byte {
	b, err := json.Marshal(g)
	if err != nil {
		fmt.Println(err)
		return nil
	}

	b = append([]byte(fmt.Sprintf("%s ", g.Id)), b...)

	return b
}
