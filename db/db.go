package db

import (
	"database/sql"
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

type PlayerCount struct {
	Player string `json:"player"`
	Count  int    `json:"count"`
}

func (db *Db) GetPlayerCount() ([]*PlayerCount, error) {
	countsMap := make(map[string]int, 0)

	processRows := func(rows *sql.Rows) {
		for rows.Next() {
			var player string
			var count int
			err := rows.Scan(&player, &count)
			if err != nil {
				continue
			}

			if c, ok := countsMap[player]; ok {
				countsMap[player] = c + count
			} else {
				countsMap[player] = count
			}
		}
	}

	whiteRows, err := db.db.Table("games").Select("white_name, count(*) as count").Group("white_name").Rows()
	if err != nil {
		return nil, err
	}
	defer whiteRows.Close()
	processRows(whiteRows)

	blackRows, err := db.db.Table("games").Select("black_name, count(*) as count").Group("black_name").Rows()
	if err != nil {
		return nil, err
	}
	defer blackRows.Close()
	processRows(blackRows)

	counts := make([]*PlayerCount, 0)
	for player, count := range countsMap {
		counts = append(counts, &PlayerCount{
			Player: player,
			Count:  count,
		})
	}

	return counts, nil
}

type OpeningCount struct {
	Opening string `json:"opening"`
	Count   int    `json:"count"`
}

func (db *Db) GetOpeningCount() ([]*OpeningCount, error) {
	rows, err := db.db.Table("games").Select("opening, count(*) as count").Group("opening").Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	openingCounts := make([]*OpeningCount, 0)
	for rows.Next() {
		oc := &OpeningCount{}
		err := rows.Scan(&oc.Opening, &oc.Count)
		if err != nil {
			continue
		}
		openingCounts = append(openingCounts, oc)
	}

	return openingCounts, nil
}

func (db *Db) GetAllGamesPaged(offset int, count int) ([]*dto.Game, error) {
	games := make([]*dto.Game, 0)
	tx := db.db.Limit(count).Offset(offset).Find(&games)
	if tx.Error != nil {
		return nil, tx.Error
	}

	return games, nil
}

func (db *Db) GetAllGames() ([]*dto.Game, error) {
	games := make([]*dto.Game, 0)
	tx := db.db.Find(&games)
	if tx.Error != nil {
		return nil, tx.Error
	}

	return games, nil
}

func (db *Db) GetHighestRatedGame() (*dto.Game, error) {
	highestWhiteGame := &dto.Game{}
	highestBlackGame := &dto.Game{}
	if tx := db.db.Order("white_rating DESC").First(highestWhiteGame); tx.Error != nil {
		return nil, tx.Error
	}

	if tx := db.db.Order("black_rating DESC").First(highestBlackGame); tx.Error != nil {
		return nil, tx.Error
	}

	if highestWhiteGame.WhiteRating > highestBlackGame.BlackRating {
		return highestWhiteGame, nil
	}

	return highestBlackGame, nil
}

func (db *Db) GetAverageRating() (int, error) {
	games, err := db.GetAllGames()
	if err != nil {
		return -1, err
	}

	sum := 0
	for _, g := range games {
		sum += g.WhiteRating
		sum += g.BlackRating
	}

	return sum / (len(games) * 2), nil
}

func (db *Db) GetLowestRatedGame() (*dto.Game, error) {
	lowestWhiteGame := &dto.Game{}
	lowestBlackGame := &dto.Game{}
	if tx := db.db.Order("white_rating").First(lowestWhiteGame); tx.Error != nil {
		return nil, tx.Error
	}

	if tx := db.db.Order("black_rating").First(lowestBlackGame); tx.Error != nil {
		return nil, tx.Error
	}

	if lowestWhiteGame.WhiteRating < lowestBlackGame.BlackRating {
		return lowestWhiteGame, nil
	}

	return lowestBlackGame, nil
}
