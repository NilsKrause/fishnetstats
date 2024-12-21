package dto

import (
	"github.com/joanlopez/go-lichess/lichess"
	"gorm.io/gorm"
)

type Game struct {
	gorm.Model
	Variant     string
	Game        string
	Fetched     bool
	WhiteName   string
	WhiteRating int
	BlackName   string
	BlackRating int
	Opening     string
	Speed       string
}

func getPtrValue[P int | string | lichess.GameOpening](val *P) P {
	if val != nil {
		return *val
	}

	var p P
	return p
}

func LiGameToDto(game *lichess.Game) *Game {
	opening := ""
	if game.Opening != nil {
		opening = getPtrValue(game.Opening.Name)
	}

	whiteName := ""
	blackName := ""
	if game.Players.White.User != nil {
		whiteName = game.Players.White.User.Name
	}

	if game.Players.Black.User != nil {
		blackName = game.Players.Black.User.Name
	}

	return &Game{
		Game:        game.Id,
		Variant:     string(game.Variant),
		WhiteName:   whiteName,
		WhiteRating: getPtrValue(game.Players.White.Rating),
		BlackName:   blackName,
		BlackRating: getPtrValue(game.Players.Black.Rating),
		Opening:     opening,
		Speed:       string(game.Speed),
	}
}
