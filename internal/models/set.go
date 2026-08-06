package models

import "time"

type Set struct {
	ID            string  `json:"id"`
	WorkoutItemID string  `json:"workout_item_id"`
	SetNumber     int     `json:"set_number"`
	WeightKg      float64 `json:"weight_kg"`
	Reps          int     `json:"reps"`
}

// SetHistoryEntry is a denormalized view of a set joined with its machine
// and workout date — what the history endpoints return, as opposed to Set.
type SetHistoryEntry struct {
	ID          string    `json:"id"`
	MachineName string    `json:"machine_name"`
	WorkoutDate time.Time `json:"workout_date"`
	SetNumber   int       `json:"set_number"`
	WeightKg    float64   `json:"weight_kg"`
	Reps        int       `json:"reps"`
}

type SetDetail struct {
	ID          string    `json:"id"`
	MachineID   string    `json:"machine_id"`
	MachineName string    `json:"machine_name"`
	WorkoutDate time.Time `json:"workout_date"`
	SetNumber   int       `json:"set_number"`
	WeightKg    float64   `json:"weight_kg"`
	Reps        int       `json:"reps"`
}

type DaySetEntry struct {
	MachineID       string  `json:"machine_id"`
	MachineName     string  `json:"machine_name"`
	MachinePhotoURL string  `json:"machine_photo_url"`
	SetNumber       int     `json:"set_number"`
	WeightKg        float64 `json:"weight_kg"`
	Reps            int     `json:"reps"`
}
