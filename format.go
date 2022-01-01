package main

type Format string

var formats = []Format{
	Blitz,
	Puzzle,
	Bullet,
	Correspondence,
	Classical,
	Rapid,
}

const (
	Blitz          Format = "blitz"
	Puzzle         Format = "puzzle"
	Bullet         Format = "bullet"
	Correspondence Format = "correspondence"
	Classical      Format = "classical"
	Rapid          Format = "rapid"
)
