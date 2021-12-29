package main

import (
	"fmt"
	"strconv"
	"strings"
)

type PlayerElo map[Format]int

func newPlayerElo () PlayerElo {
	return make(map[Format]int)
}

func (pe *PlayerElo) MarshalJSON() ([]byte, error) {
	elos := "["
	i := 0
	for format, elo := range *pe {
		if i == len(*pe)-1 {
			elos = fmt.Sprintf("%s{\"%s\": %d}", elos, format, elo)
		} else {
			elos = fmt.Sprintf("%s{\"%s\": %d}, ", elos, format, elo)
		}
	}
	elos = fmt.Sprintf("%s] ", elos)

	return []byte(elos), nil
}

func (pe *PlayerElo) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), "[]")
	elos := strings.Split(s, ",")

	// {"blitz": "1234"}
	for _, eloA := range elos {
		ss := strings.Split(strings.Trim(eloA, "{}"), ": ")
		if len(ss) != 2 {
			continue
		}
		i, err := strconv.Atoi(strings.Trim(ss[1], "\""))
		if err != nil {
			fmt.Printf("error while unmarshelling player elo %s", err)
			continue
		}
		format := Format(strings.Trim(ss[0], "\""))
		(*pe)[format] = i
	}

	return nil
}
