package main

import (
	"encoding/json"
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
}

func (g *Game) Initialize() {
	Request(g)
}

func (g *Game) GetBody () io.Reader {
	return strings.NewReader(g.Id.String())
}

func (g *Game) GetUrl () string {
	return "https://lichess.org/games/export/_ids"
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

	lines := strings.Split(string(b), "\n")
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
			g.White = &Player{Name: PlayerName(value)}
			g.White.Initialize()
			fmt.Printf("White: %s ", g.White.Name)
			break
		case "Black":
			g.Black = &Player{Name: PlayerName(value)}
			g.Black.Initialize()
			fmt.Printf("Black: %s ", g.Black.Name)
			break
		case "Result":
			g.Result = aToRes(value)
			fmt.Printf("Resul: %v ", g.Result)
			break
		case "Termination":
			g.Termination = value
			fmt.Printf("Termination: %s ", g.Termination)
			break
		case "TimeControl":
			g.Format = aToTC(value)
			fmt.Printf("Format: %d+%d ", g.Format.Seconds, g.Format.Bonus)
		}
	}

	fmt.Printf("\n")

	gb, err := json.Marshal(g)
	if err == nil {
		fmt.Printf("parsed game %s and got this data: %s\n", g.Id, string(gb))
	}
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
	if g.Id.String() == "iM76BKY1" {
		fmt.Printf("MARSHALING :D")
	}
	if err != nil {
		fmt.Println(err)
		return nil
	}

	//fmt.Printf("Game %s to Bytes: %s\n", g.Id, string(b))
	b = append([]byte(fmt.Sprintf("%s ", g.Id)), b...)

	return b
}
