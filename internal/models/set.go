package models

import "time"

type Set struct {
	ID            string  `json:"id"`
	WorkoutItemID string  `json:"workout_item_id"`
	SetNumber     int     `json:"set_number"`
	WeightKg      float64 `json:"weight_kg"`
	Reps          int     `json:"reps"`
}

// денормализованное представление подхода с данными машины и датой тренировки — для эндпоинтов истории
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
