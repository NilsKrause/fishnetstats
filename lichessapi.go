package main

import (
	"fmt"
	"io"
	"net/http"
	"time"
)

type Gettable interface {
	ParseHttpResponse (response *http.Response)
	GetUrl () string
	GetBody () io.Reader
}

type lichessapi struct {
	c http.Client
	q chan Gettable
}

var api *lichessapi

func init () {
	api = &lichessapi{
		c: http.Client{},
		q: make(chan Gettable, 10000),
	}
	fmt.Println("starting request queue")
	go api.handleRequests()
}

func (l *lichessapi) handleRequests () {
	for req := range l.q {
		res, err := l.c.Post(req.GetUrl(), "text/plain", req.GetBody())
		if err != nil {
			fmt.Printf("error while requesting game %s\n", err)
			return
		}

		if res.StatusCode == 429 {
			fmt.Println("Got 429 StatusCode - pausing for a while..")
			<-time.After(time.Second * 75)
		}

		req.ParseHttpResponse(res)

		<-time.After(time.Second * 10)
	}

	fmt.Println("stopped request queue")
}

func Request (g Gettable) {
	go func () {
		api.q <- g
	}()
}
