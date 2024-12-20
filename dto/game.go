package dto

import (
	"github.com/joanlopez/go-lichess/lichess"
	"gorm.io/gorm"
)

type Game struct {
	gorm.Model
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
	return &Game{
		Game:        game.Id,
		WhiteName:   game.Players.White.User.Name,
		WhiteRating: getPtrValue(game.Players.White.Rating),
		BlackName:   game.Players.Black.User.Name,
		BlackRating: getPtrValue(game.Players.Black.Rating),
		Opening:     getPtrValue(game.Opening.Name),
		Speed:       string(game.Speed),
	}
}
