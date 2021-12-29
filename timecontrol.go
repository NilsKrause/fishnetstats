package main

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type Timecontrol struct {
	Seconds int
	Bonus   int
}

func (t *Timecontrol) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("\"%d+%d\" ", t.Seconds, t.Bonus)), nil
}

func (t *Timecontrol) UnmarshalJSON(b []byte) error {
	bs := string(b)
	tc := aToTC(bs)

	if tc == nil {
		return errors.New(fmt.Sprintf("could not convert %s to timecontrol.", bs))
	}

	t.Seconds = tc.Seconds
	t.Bonus = tc.Bonus

	return nil
}

func (t *Timecontrol) asFormat() Format {
	minutes := t.Seconds / 60

	if minutes < 3 {
		return Bullet
	}

	if minutes < 10 {
		return Blitz
	}

	if minutes < 30 {
		return Rapid
	}

	return Classical
}

func aToTC(timecontrol string) *Timecontrol {
	s := strings.Split(strings.Trim(timecontrol, "\""), "+")
	if len(s) != 2 {
		return nil
	}

	seconds, err := strconv.Atoi(s[0])
	if err != nil {
		return nil
	}

	bonus, err := strconv.Atoi(s[1])
	if err != nil {
		return nil
	}

	return &Timecontrol{
		Seconds: seconds,
		Bonus:   bonus,
	}
}
