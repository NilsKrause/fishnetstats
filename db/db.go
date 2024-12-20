package db

import (
	"de.nilskrau.fishnet.stats/dto"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

type Db struct {
	db *gorm.DB
}

type Config struct {
	Location string `yaml:"location"`
}

func New(config *Config) *Db {
	db, err := gorm.Open(sqlite.Open(config.Location))

	if err != nil {
		panic(err)
	}

	if err := db.AutoMigrate(&dto.Game{}); err != nil {
		panic(err)
	}

	return &Db{
		db: db,
	}
}

func (db *Db) SaveGame(game *dto.Game) error {
	tx := db.db.Save(game)
	return tx.Error
}

func (db *Db) DoesExist(id string) bool {
	tx := db.db.Where("Game = ?", id).First(&dto.Game{})

	if tx.RowsAffected == 0 {
		return false
	}

	return true
}
