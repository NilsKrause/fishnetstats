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
	Id    string     `json:"id,omitempty"`
	Name  PlayerName `json:"name,omitempty"`
	Elo   PlayerElo  `json:"elo,omitempty"`
	Title string     `json:"title,omitempty"`

	games       []*Game
	Initialized bool `json:"initialized,omitempty"`
	requesting  bool
}

func NewPlayer(name string, game *Game) *Player {
	if player, ok := players[PlayerName(name)]; ok {
		if game != nil {
			player.games = append(player.games, game)
		}
		return player
	}

	player := &Player{
		Id:         strings.ToLower(name),
		Name:       PlayerName(name),
		games:      make([]*Game, 1),
		requesting: false,
	}

	addPlayer(player)

	if game != nil {
		player.games = append(player.games, game)
	}

	return player
}

func (p *Player) IsInitialized() bool {
	return p.Initialized
}

func (p *Player) HasId(id interface{}) bool {
	nid := id.(PlayerName)
	if nid == p.Name {
		return true
	}

	return false
}

func (p *Player) GetBodyString() string {
	return string(p.Name)
}

func (p *Player) GetType() GettableType {
	return PlayerT
}

func (p *Player) GetBody() io.Reader {
	return strings.NewReader(string(p.Name))
}

func (p *Player) GetUrl() string {
	return "https://lichess.org/api/users"
}

func (p *Player) Initialize() {
	if p.IsInitialized() {
		return
	}
	if !p.requesting {
		p.requesting = true
		fmt.Printf("initializing player %s\n", p.Name)
		Request(p)
	}
}

func (p *Player) FromJsonResponse(player *ApiPlayerResponse) {
	p.Elo = newPlayerElo()
	p.Elo[Blitz] = player.Perfs.Blitz.Rating
	p.Elo[Bullet] = player.Perfs.Bullet.Rating
	p.Elo[Rapid] = player.Perfs.Rapid.Rating
	p.Elo[Classical] = player.Perfs.Classical.Rating

	p.Title = player.Title
	p.Initialized = true
	for _, g := range p.games {
		if g != nil {
			g.playerUpdated()
		}
	}
}

func (p *Player) ParseResponseBody(b string) {
	player := &ApiPlayerResponse{}

	err := json.Unmarshal([]byte(b), &player)
	if err != nil {
		fmt.Println(err)
		return
	}

	p.FromJsonResponse(player)
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

	p.ParseResponseBody(string(b))
}
