package api

import (
	"context"
	"de.nilskrau.fishnet.stats/db"
	"errors"
	"fmt"
	"github.com/labstack/echo/v4"
	"net/http"
	"time"
)

type Api struct {
	c *Config
	e *echo.Echo
}

type Config struct {
	Port int
	Host string
}

type baseContext struct {
	echo.Context
	db *db.Db
}

func New(conf *Config, db *db.Db) Api {
	e := echo.New()

	// populate the context with database
	ee := e.Group("", func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			return next(&baseContext{
				Context: c,
				db:      db,
			})
		}
	})

	// list games
	ee.GET("games", listGames)

	// highes elo
	ee.GET("highest", highesRatedGame)

	// lowest elo
	ee.GET("lowest", lowestRatedGame)

	// average elo
	ee.GET("average", averageRating)

	// opening count
	ee.GET("openings", openingCount)

	// player with most analysed games
	ee.GET("players", playerCount)

	return Api{
		c: conf,
		e: e,
	}
}

func (a *Api) Start(parent context.Context) {
	go func() {
		if err := a.e.Start(fmt.Sprintf("%s:%d", a.c.Host, a.c.Port)); err != nil && !errors.Is(err, http.ErrServerClosed) {
			a.e.Logger.Fatal("shutting down the server", err)
		}
	}()

	// listens if parent context stops, this also stops the api server
	go func() {
		done := parent.Done()
		if done == nil {
			return
		}
		<-done
		_ = a.stop(parent)
	}()
}

func (a *Api) stop(parent context.Context) error {
	var ctx context.Context
	var cancel context.CancelFunc

	if parent == nil {
		ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
	} else {
		ctx, cancel = context.WithTimeout(parent, 10*time.Second)
	}

	defer cancel()

	if err := a.e.Shutdown(ctx); err != nil {
		a.e.Logger.Fatal(err)
	}

	return nil
}

func (a *Api) Stop() {
	_ = a.stop(context.TODO())
}
