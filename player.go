package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
)

type PlayerName string

type Player struct {
	Name  PlayerName `json:"name,omitempty"`
	Elo   PlayerElo  `json:"elo,omitempty"`
	Title string     `json:"title,omitempty"`
}

func (p *Player) GetBody () io.Reader {
	return strings.NewReader(string(p.Name))
}

func (p *Player) GetUrl () string {
	return "https://lichess.org/api/users"
}

func (p *Player) Initialize() {
	Request(p)
}

func (p *Player) ParseHttpResponse(r *http.Response) {
	b := make([]byte, 0)
	b, err := ioutil.ReadAll(r.Body)
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(r.Body)

	if err != nil || len(b) <= 0 {
		fmt.Println(err)
		return
	}

	player := struct {
		Perfs struct {
			Blitz struct {
				Rating int `json:"rating"`
			} `json:"blitz"`
			Bullet struct {
				Rating int `json:"rating"`
			} `json:"bullet"`
			Rapid struct {
				Rating int `json:"rating"`
			} `json:"rapid"`
			Classical struct {
				Rating int `json:"rating"`
			} `json:"classical"`
			Puzzles struct {
				Rating int `json:"rating"`
			} `json:"puzzles"`
		} `json:"perfs"`
		Title string `json:"title"`
	}{}

	err = json.Unmarshal(b, &player)
	if err != nil {
		fmt.Println(err)
		return
	}

	p.Elo = newPlayerElo()
	p.Elo[Blitz] = player.Perfs.Blitz.Rating
	p.Elo[Bullet] = player.Perfs.Bullet.Rating
	p.Elo[Rapid] = player.Perfs.Rapid.Rating
	p.Elo[Classical] = player.Perfs.Classical.Rating

	p.Title = player.Title
}
