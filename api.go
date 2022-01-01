package main

import (
	"encoding/json"
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
	"golang.org/x/net/context"
	"net/http"
	"time"
)

type api struct {
	echo *echo.Echo
}

type Api interface {
	Stop()
	Start(port int)
}

func (a *api) middleware() {
	// Middleware
	a.echo.Use(middleware.Logger())
	a.echo.Use(middleware.Recover())
}

func (a *api) routes() {
	v1 := a.echo.Group("/v1")
	v1.GET("/stats", getStatus)
}

func NewApi() Api {
	api := &api{
		echo: echo.New(),
	}

	api.echo.Logger.SetLevel(log.DEBUG)
	api.middleware()
	api.routes()

	if data, err := json.MarshalIndent(api.echo.Routes(), "", "  "); err == nil {
		fmt.Printf("%s\n", string(data))
	}

	return api
}

func (a *api) Start(port int) {
	go func() {
		if err := a.echo.Start(fmt.Sprintf(":%d", port)); err != nil && err != http.ErrServerClosed {
			a.echo.Logger.Fatal("shutting down the server")
		}
	}()
}

func (a *api) Stop() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	defer cancel()

	if err := a.echo.Shutdown(ctx); err != nil {
		a.echo.Logger.Fatal(err)
	}
}

func getStatus(c echo.Context) error {
	statss, err := json.Marshal(stats)

	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	return c.String(http.StatusOK, string(statss))
}
