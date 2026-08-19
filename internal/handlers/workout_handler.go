package handlers

import (
	"encoding/json"
	"gym_tracker/internal/middleware"
	"gym_tracker/internal/repository"
	"net/http"
)

type WorkoutHandler struct {
	workoutRepo *repository.WorkoutRepository
}

func NewWorkoutHandler(workoutRepo *repository.WorkoutRepository) *WorkoutHandler {
	return &WorkoutHandler{workoutRepo}
}

func (h *WorkoutHandler) GetHistory(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(string)
	workouts, err := h.workoutRepo.GetWorkoutHistory(r.Context(), userID)
	if err != nil {
		serverError(w, err, "failed to get workout history")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(workouts)
}
