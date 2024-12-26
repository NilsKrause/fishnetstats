package api

import (
	"github.com/labstack/echo/v4"
	"net/http"
	"strconv"
)

func listGames(cc echo.Context) error {
	c := cc.(*baseContext)

	pageStr := c.QueryParam("page")
	page, err := strconv.Atoi(pageStr)
	if err != nil {
		page = 0
	} else {
		page -= 1
	}

	pageCount := 20
	offset := page * pageCount

	games, err := c.db.GetAllGamesPaged(offset, pageCount)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, games)
}

func highesRatedGame(cc echo.Context) error {
	c := cc.(*baseContext)
	highest, err := c.db.GetHighestRatedGame()
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, highest)
}

func lowestRatedGame(cc echo.Context) error {
	c := cc.(*baseContext)
	lowest, err := c.db.GetLowestRatedGame()
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, lowest)
}

func averageRating(cc echo.Context) error {
	c := cc.(*baseContext)
	average, err := c.db.GetAverageRating()
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, struct {
		Rating int `json:"rating"`
	}{Rating: average})
}

func openingCount(cc echo.Context) error {
	c := cc.(*baseContext)

	count, err := c.db.GetOpeningCount()
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, count)
}

func playerCount(cc echo.Context) error {
	c := cc.(*baseContext)

	count, err := c.db.GetPlayerCount()
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, count)
}
