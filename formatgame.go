package main

import (
	"fmt"
	"strings"
)

type FormatGame map[Format]Gameid

func NewFormatGame() FormatGame {
	return make(map[Format]Gameid)
}

func (f *FormatGame) MarshalJSON() ([]byte, error) {
	elos := "["
	i := 0
	for format, elo := range *f {
		if i == len(*f)-1 {
			elos = fmt.Sprintf("%s{\"%s\": \"%d\"}", elos, format, elo)
		} else {
			elos = fmt.Sprintf("%s{\"%s\": \"%d\"}, ", elos, format, elo)
		}
		i++
	}
	elos = fmt.Sprintf("%s] ", elos)

	return []byte(elos), nil
}

func (f *FormatGame) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), "[]")
	games := strings.Split(s, ",")

	// {"blitz": "1234"}
	for _, gameA := range games {
		ss := strings.Split(strings.Trim(gameA, "{}"), ": ")
		if len(ss) != 2 {
			fmt.Printf("split error while unmarshalling format game %s\n", string(b))
			continue
		}
		gameid := strings.Trim(ss[1], "\"")
		if len(gameid) != 8 {
			fmt.Printf("gid error while unmarshalling format game %s\n", string(b))
			continue
		}
		format := Format(strings.Trim(ss[0], "\""))
		(*f)[format] = aToGid(gameid)
	}

	return nil
}
