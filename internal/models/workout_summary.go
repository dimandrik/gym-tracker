package models

import "time"

type WorkoutSummary struct {
	WorkoutDate   time.Time `json:"workout_date"`
	MachinesCount int       `json:"machines_count"`
	SetsCount     int       `json:"sets_count"`
}
