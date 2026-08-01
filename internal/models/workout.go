package models

import "time"

type Workout struct {
	ID          string    `json:"id"`
	UserID      string    `json:"user_id"`
	WorkoutDate time.Time `json:"workout_date"`
}
