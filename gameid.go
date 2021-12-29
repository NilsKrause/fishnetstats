package main

type Gameid [8]byte

func (g *Gameid) String () string {
	return string((*g)[:])
}