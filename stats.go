package main

import (
	"fmt"
)

var stats *Stats

func init() {
	stats = &Stats{
		gamesCount: make(map[Format]int),
		eloSum:     NewFormatElo(),
		averageElo: NewFormatElo(),
		highestElo: NewFormatElo(),
		lowestElo:  NewFormatElo(),

		titledPlayers: make(map[string]string),

		eloPlayers:    make(map[string]byte),
		numberOfGames: 0,
	}
}

type Stats struct {
	gamesCount map[Format]int
	eloSum     FormatElo

	eloPlayers map[string]byte

	averageElo FormatElo
	highestElo FormatElo
	lowestElo  FormatElo

	titledPlayers map[string]string

	numberOfGames int
}

func (s *Stats) MarshalJSONElos() []byte {
	elos := "\"elos\":["
	aelos := make([]string, 0)
	for _, format := range formats {
		elo := fmt.Sprintf("{\"%s\":{", format)
		h, hok := s.highestElo[format]
		l, lok := s.lowestElo[format]
		a, aok := s.averageElo[format]

		if !hok && !lok && !aok {
			continue
		}

		if hok {
			if lok || aok {
				elo = fmt.Sprintf("%s\"high\":%d,", elo, h)
			} else {
				elo = fmt.Sprintf("%s\"high\":%d", elo, h)
			}
		}

		if lok {
			if aok {
				elo = fmt.Sprintf("%s\"low\":%d,", elo, l)
			} else {
				elo = fmt.Sprintf("%s\"low\":%d", elo, l)
			}
		}

		if aok {
			elo = fmt.Sprintf("%s\"average\":%d", elo, a)
		}

		elo = fmt.Sprintf("%s}}", elo)

		aelos = append(aelos, elo)
	}

	for i, elo := range aelos {
		if i == len(aelos)-1 {
			elos = fmt.Sprintf("%s%s", elos, elo)
		} else {
			elos = fmt.Sprintf("%s%s,", elos, elo)
		}
	}

	elos = fmt.Sprintf("%s] ", elos)

	return []byte(elos)
}

func (s *Stats) MarshalJSONPlayers() []byte {
	ps := make(map[string]string)
	pn := make(map[string]int)
	for p, t := range s.titledPlayers {
		tp, ok := ps[t]
		if ok {
			ps[t] = fmt.Sprintf("%s,\"%s\"", tp, p)
		} else {
			ps[t] = fmt.Sprintf("\"%s\"", p)
		}
		if n, ok := pn[t]; ok {
			pn[t] = n + 1
		} else {
			pn[t] = 1
		}
	}
	ts := make([]string, 0)
	for t, tps := range ps {
		ts = append(ts, fmt.Sprintf("{\"title\":\"%s\",\"nplayers\":%d,\"players\":[%s]}", t, pn[t], tps))
	}
	p := ""
	if len(ts) == 0 {
		p = "\"players\":[]"
	} else {
		p = "\"players\":["
	}
	for i, t := range ts {
		if i == len(ts)-1 {
			p = fmt.Sprintf("%s%s]", p, t)
		} else {
			p = fmt.Sprintf("%s%s,", p, t)
		}
	}
	return []byte(p)
}

func (s *Stats) MarshalJSON() ([]byte, error) {
	/*
		{
			"elos": [
				"blitz" :{
					"high": 2600,
					"low": 800,
					"average": 1500
				},
				"bullet" :{
					"high": 2600,
					"low": 800,
					"average": 1500
				},
			],
			"players": [
				{
					"title": "gm",
					"nplayers": 20,
					"players": [
						"hugo",
						"bernd"
					]
				},
				{
					"title": "im",
					"nplayers": 300,
					"players": [
						"hugo",
						"bernd"
					]
				}
			]
		}
	*/

	stats := make([]byte, 0)

	stats = append(stats, '{')
	stats = append(stats, s.MarshalJSONElos()...)
	stats = append(stats, ',')
	stats = append(stats, s.MarshalJSONPlayers()...)
	stats = append(stats, ',')
	stats = append(stats, []byte(fmt.Sprintf("\"gamesCount\":%d", s.numberOfGames))...)
	stats = append(stats, ',')
	stats = append(stats, []byte{'}', ' '}...)

	return stats, nil
}

func (s *Stats) updateAverage(format Format, elo int) {
	nSum := elo
	nCnt := 1

	if sum, ok := s.eloSum[format]; ok {
		nSum += sum
	}

	if cnt, ok := s.gamesCount[format]; ok {
		nCnt += cnt
	}

	s.eloSum[format] = nSum
	s.gamesCount[format] = nCnt
	s.averageElo[format] = nSum / nCnt
}

func (s *Stats) updateLowesElo(format Format, elo int) {
	if lowest, ok := s.lowestElo[format]; ok {
		if elo >= lowest {
			return
		}
	}

	s.lowestElo[format] = elo
}

func (s *Stats) updateHighestElo(format Format, elo int) {
	if lowest, ok := s.highestElo[format]; ok {
		if elo <= lowest {
			return
		}
	}

	s.highestElo[format] = elo
}

func (s *Stats) updateElos(g *Game) {
	if _, ok := s.eloPlayers[g.Black.Id]; !ok {
		s.eloPlayers[g.Black.Id] = 0x0
		for format, elo := range g.Black.Elo {
			if elo == 0 {
				continue
			}
			s.updateLowesElo(format, elo)
			s.updateHighestElo(format, elo)
			s.updateAverage(format, elo)
		}
	}

	if _, ok := s.eloPlayers[g.White.Id]; !ok {
		s.eloPlayers[g.White.Id] = 0x0
		for format, elo := range g.White.Elo {
			if elo == 0 {
				continue
			}
			s.updateLowesElo(format, elo)
			s.updateHighestElo(format, elo)
			s.updateAverage(format, elo)
		}
	}
}

func (s *Stats) updateTitledPlayers(p *Player) {
	if p.Title == "" {
		return
	}

	s.titledPlayers[string(p.Name)] = p.Title
}

func (s *Stats) addGame(g *Game) {
	if !g.IsInitialized() {
		return
	}

	s.updateElos(g)
	s.updateTitledPlayers(g.White)
	s.updateTitledPlayers(g.Black)
	s.numberOfGames++
}

func (s *Stats) Initialize() {
	for _, g := range games {
		s.updateElos(g)
	}
	s.numberOfGames = len(games)
}

func (s *Stats) GetTitledPlayers() map[string]string {
	return s.titledPlayers
}

func (s *Stats) GetNumberOfGames() int {
	return len(games)
}

func (s *Stats) GetAverageElo() int {
	return 0
}
