package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

type GettableType string

const (
	GameT   GettableType = "game"
	PlayerT GettableType = "player"
)

type Gettable interface {
	ParseResponseBody(body string)
	ParseHttpResponse(response *http.Response)
	HasId(id interface{}) bool
	GetType() GettableType
	GetUrl() string
	GetBody() io.Reader
	GetBodyString() string
}

type lichessRes struct {
	res *http.Response
	err error
}

type lichessReq struct {
	url  string
	body io.Reader
	res  chan<- *lichessRes
}

type lichessapi struct {
	c   http.Client
	gq  chan Gettable
	pq  chan Gettable
	req chan lichessReq
}

var lichess *lichessapi

func init() {
	lichess = &lichessapi{
		c:   http.Client{},
		gq:  make(chan Gettable, 1001),
		pq:  make(chan Gettable, 1001),
		req: make(chan lichessReq, 10),
	}
	fmt.Println("starting request queue")
	lichess.handleRequests()
}

func (l *lichessapi) nextLoad(gettableType GettableType) string {
	var c chan Gettable
	if gettableType == GameT {
		c = l.gq
	} else {
		c = l.pq
	}

	var body string

	for elem := range c {
		body = elem.GetBodyString()
		<-time.After(time.Millisecond * 500)
		for i := 0; i < 499; i++ {
			select {
			case elem := <-c:
				// if there is more frame immediately available, we add them to our slice
				body = fmt.Sprintf("%s,%s", body, elem.GetBodyString())
			default:
				// else we move on without blocking
				return body
			}
		}
		return body
	}

	return body
}

func getBody(r *http.Response) (string, error) {
	b := make([]byte, 0)
	b, err := ioutil.ReadAll(r.Body)
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(r.Body)

	if err != nil {
		err := fmt.Sprintf("error while reading game response body %s\n", err.Error())
		fmt.Printf(err)
		return "", errors.New(err)
	}

	if len(b) <= 0 {
		err := errors.New("error while reading game response body, length less or equal zero")
		fmt.Println(err)
		return "", err
	}

	return string(b), nil
}

type ApiPlayerResponse struct {
	Id        string `json:"id"`
	Username  string `json:"username"`
	Patron    bool   `json:"patron"`
	CreatedAt int    `json:"createdAt"`
	Profile   struct {
		Country    string `json:"country"`
		Location   string `json:"location"`
		Bio        string `json:"bio"`
		FirstName  string `json:"firstName"`
		LastName   string `json:"lastName"`
		FideRating int    `json:"fideRating"`
		UscfRating int    `json:"uscfRating"`
		EcfRating  int    `json:"ecfRating"`
		Links      string `json:"links"`
	} `json:"profile"`
	PlayTime struct {
		Total int `json:"total"`
		Tv    int `json:"tv"`
	} `json:"playTime"`
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
}

func (l *lichessapi) handleGameRequests() {
	fmt.Println("started game request handler")
	for {
		reqBody := l.nextLoad(GameT)
		resChan := make(chan *lichessRes)

		l.req <- lichessReq{
			url:  "https://lichess.org/games/export/_ids",
			body: strings.NewReader(reqBody),
			res:  resChan,
		}

		r, ok := <-resChan
		close(resChan)
		res := r.res
		err := r.err

		if !ok || err != nil {
			fmt.Printf("error while requesting games %s\n", err)
			return
		}

		resBody, err := getBody(res)
		if err != nil {
			continue
		}

		gameBodys := strings.Split(resBody, "\n\n")
		for _, gbody := range gameBodys {
			gid, err := getGameIdFromBody(gbody)
			if err != nil {
				continue
			}

			if game, ok := games[gid]; ok {
				game.ParseResponseBody(gbody)
			}
		}
	}
}

func (l *lichessapi) handlePlayerRequests() {
	fmt.Println("started player request handler")
	for {
		body := l.nextLoad(PlayerT)
		resChan := make(chan *lichessRes)

		l.req <- lichessReq{
			url:  "https://lichess.org/api/users",
			body: strings.NewReader(body),
			res:  resChan,
		}

		r, ok := <-resChan
		fmt.Println("got player response")

		close(resChan)
		res := r.res
		err := r.err

		if !ok || err != nil {
			fmt.Printf("error while requesting players %s\n", err)
			return
		}

		resBody, err := getBody(res)
		if err != nil {
			fmt.Printf("error while getting body %s\n", err)
			continue
		}

		playerRes := make([]*ApiPlayerResponse, 0)
		err = json.Unmarshal([]byte(resBody), &playerRes)
		if err != nil {
			fmt.Printf("error while unmarshalling response body %s\n", err)
			continue
		}

		fmt.Printf("got %d new players", len(playerRes))

		for _, pres := range playerRes {
			if player, ok := players[PlayerName(pres.Username)]; ok {
				player.FromJsonResponse(pres)
			} else {
				fmt.Printf("got unknown player %s\n", pres.Username)
			}
		}
	}
}

func (l *lichessapi) handlePostRequests() {
	fmt.Println("starting post request handler")
	for {
		select {
		case req := <-l.req:
			fmt.Printf("got request to %s\n", req.url)
			res, err := l.c.Post(req.url, "text/plain", req.body)

			if err == nil && res.StatusCode == 429 {
				err = errors.New("got 429 status")
			}

			req.res <- &lichessRes{
				res: res,
				err: err,
			}

			if res != nil && res.StatusCode == 429 {
				fmt.Println("Got 429 StatusCode - pausing for a while..")
				<-time.After(time.Second * 75)
			} else {
				fmt.Println("waiting 8 seconds to not flood the api")
				<-time.After(time.Second * 8)
			}
		}
	}
}

func (l *lichessapi) handleRequests() {
	go l.handlePostRequests()
	go l.handleGameRequests()
	go l.handlePlayerRequests()
}

func Request(g Gettable) {
	go func() {
		if g.GetType() == GameT {
			lichess.gq <- g
		} else if g.GetType() == PlayerT {
			lichess.pq <- g
		}
	}()
}
