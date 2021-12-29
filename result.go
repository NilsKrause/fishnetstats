package main

import (
	"errors"
	"fmt"
	"strings"
)

type Result bool

func aToRes(a string) *Result {
	return bToRes([]byte(a))
}

func bToRes(b []byte) *Result {
	rs := string(b)
	s := strings.Split(rs, "-")
	var res Result = false
	if len(s) != 2 {
		return &res
	}

	res = s[0] == "1"

	return &res // true = white won
}

func (r *Result) String () string {
	if *r {
		return "1-0"
	}

	return "0-1"
}

func (r *Result) MarshalJSON() ([]byte, error) {
	if *r {
		return []byte("\"1-0\" "), nil
	}

	return []byte("\"0-1\" "), nil
}

func (r *Result) UnmarshalJSON(b []byte) error {
	rs := string(b)
	s := strings.Split(rs, "-")
	if len(s) != 2 {
		return errors.New(fmt.Sprintf("could not convert %s to result.", rs))
	}

	if s[0] == "1" {
		*r = true
	} else {
		*r = false
	}

	return nil
}
