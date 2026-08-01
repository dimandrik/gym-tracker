package models

type WorkoutItem struct {
	ID        string `json:"id"`
	WorkoutID string `json:"workout_id"`
	MachineID string `json:"machine_id"`
}
