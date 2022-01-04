package main

import (
	"fmt"
	"strconv"
	"strings"
)

type FormatElo map[Format]int

func NewFormatElo() FormatElo {
	return make(map[Format]int)
}

func (f *FormatElo) MarshalJSON() ([]byte, error) {
	elos := "["
	i := 0
	for format, elo := range *f {
		if i == len(*f)-1 {
			elos = fmt.Sprintf("%s{\"%s\": %d}", elos, format, elo)
		} else {
			elos = fmt.Sprintf("%s{\"%s\": %d}, ", elos, format, elo)
		}
		i++
	}
	elos = fmt.Sprintf("%s] ", elos)

	return []byte(elos), nil
}

func (f *FormatElo) UnmarshalJSON(b []byte) error {
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
			fmt.Printf("error while unmarshalling player elo %s\n", err)
			continue
		}
		format := Format(strings.Trim(ss[0], "\""))
		(*f)[format] = i
	}

	return nil
}
