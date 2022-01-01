package main

import (
	"fmt"
	"strconv"
	"strings"
)

type PlayerElo map[Format]int

func newPlayerElo() PlayerElo {
	return make(map[Format]int)
}

func (pe *PlayerElo) MarshalJSON() ([]byte, error) {
	elos := "["
	i := 0
	for format, elo := range *pe {
		if i == len(*pe)-1 {
			elos = fmt.Sprintf("%s{\"%s\": %d}", elos, format, elo)
		} else {
			elos = fmt.Sprintf("%s{\"%s\": %d},", elos, format, elo)
		}
		i++
	}
	elos = fmt.Sprintf("%s] ", elos)

	return []byte(elos), nil
}

func (pe *PlayerElo) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), "[]")
	elos := strings.Split(s, ",")

	*pe = newPlayerElo()

	// {"blitz": "1234"}
	for _, eloA := range elos {
		trimed := strings.Trim(eloA, "{}")
		ss := strings.Split(trimed, ":")
		if len(ss) != 2 {
			fmt.Printf("error while splitting string, got %d from %s\n", len(ss), eloA)
			continue
		}
		elostr := strings.Trim(ss[1], "\"")
		i, err := strconv.Atoi(elostr)
		if err != nil {
			fmt.Printf("error while unmarshalling player elo %s\n", err)
			continue
		}
		format := Format(strings.Trim(ss[0], "\""))
		(*pe)[format] = i
	}

	return nil
}
