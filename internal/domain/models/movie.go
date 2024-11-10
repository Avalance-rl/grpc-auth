package models

import "time"

type Movie struct {
	ID             int
	Title          string
	Duration       time.Duration
	CurrentViewers int
}
