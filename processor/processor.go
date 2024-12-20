package processor

import (
	"context"
	"de.nilskrau.fishnet.stats/db"
	"de.nilskrau.fishnet.stats/dto"
	"github.com/joanlopez/go-lichess/lichess"
)

type Processor struct {
	db     *db.Db
	client *lichess.Client
}

func New(d *db.Db) *Processor {
	return &Processor{
		db:     d,
		client: lichess.NewClient(nil),
	}
}

func (p *Processor) ProcessGameById(id string) {
	// check whether the game was already archived
	if p.db.DoesExist(id) {
		return
	}

	// request from lichess api
	game, _, err := p.client.Games.ExportById(context.TODO(), id, apiOpts())

	// save just the id in the db to restore the data later
	if err != nil {
		_ = p.db.SaveGame(&dto.Game{
			Game:    id,
			Fetched: false,
		})
		return
	}

	// store the game with meta information in the db
	dtoGame := dto.LiGameToDto(game)
	dtoGame.Fetched = true
	_ = p.db.SaveGame(dtoGame)
}

func apiOpts() *lichess.ExportOptions {
	var t = true
	var f = false
	return &lichess.ExportOptions{
		Moves:     &f,
		PgnInJson: &f,
		Tags:      &t,
		Clocks:    &f,
		Evals:     &f,
		Accuracy:  &f,
		Opening:   &t,
		Literate:  &f,
		Players:   nil,
	}
}
