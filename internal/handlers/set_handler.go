package handlers

import (
	"encoding/json"
	"gym_tracker/internal/middleware"
	"gym_tracker/internal/repository"
	"net/http"
	"time"
)

type SetHandler struct {
	workoutRepo     *repository.WorkoutRepository
	workoutItemRepo *repository.WorkoutItemRepository
	setRepo         *repository.SetRepository
}

func NewSetHandler(workoutRepo *repository.WorkoutRepository, workoutItemRepo *repository.WorkoutItemRepository, setRepo *repository.SetRepository) *SetHandler {
	return &SetHandler{workoutRepo: workoutRepo, workoutItemRepo: workoutItemRepo, setRepo: setRepo}
}

type AddSetRequest struct {
	MachineID string  `json:"machine_id"`
	WeightKg  float64 `json:"weight_kg"`
	Reps      int     `json:"reps"`
}

type UpdateSetRequest struct {
	WeightKg float64 `json:"weight_kg"`
	Reps     int     `json:"reps"`
}

// AddSet is the core write path: every set gets attached to "today's" workout,
// creating the workout and the workout-machine item on the fly if this is the
// first set logged for that machine today. Set numbers reset per machine per day.
func (h *SetHandler) AddSet(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(string)

	var req AddSetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	today := time.Now().Truncate(24 * time.Hour)
	workoutID, err := h.workoutRepo.GetOrCreateWorkout(r.Context(), userID, today)
	if err != nil {
		http.Error(w, "failed to get or create workout", http.StatusInternalServerError)
		return
	}

	itemID, err := h.workoutItemRepo.GetOrCreateWorkoutItem(r.Context(), workoutID, req.MachineID)
	if err != nil {
		http.Error(w, "failed to get or create workout item", http.StatusInternalServerError)
		return
	}

	setNumber, err := h.setRepo.GetNextSetNumber(r.Context(), itemID)
	if err != nil {
		http.Error(w, "failed to get set number", http.StatusInternalServerError)
		return
	}

	setID, err := h.setRepo.CreateSet(r.Context(), itemID, setNumber, req.WeightKg, req.Reps)
	if err != nil {
		http.Error(w, "failed to create set", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{"id": setID, "set_number": setNumber})
}

func (h *SetHandler) GetSetsByMachine(w http.ResponseWriter, r *http.Request) {
	machineID := r.URL.Query().Get("machine_id")
	if machineID == "" {
		http.Error(w, "machine_id is required", http.StatusBadRequest)
		return
	}
	sets, err := h.setRepo.GetSetsByMachineID(r.Context(), machineID)
	if err != nil {
		http.Error(w, "failed to get sets", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(sets)
}

func (h *SetHandler) DeleteSet(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(string)
	setID := r.PathValue("id")

	err := h.setRepo.DeleteSet(r.Context(), setID, userID)
	if err != nil {
		http.Error(w, "set not found", http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *SetHandler) UpdateSet(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(string)
	setID := r.PathValue("id")

	var req UpdateSetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	err := h.setRepo.UpdateSet(r.Context(), setID, userID, req.WeightKg, req.Reps)
	if err != nil {
		http.Error(w, "set not found", http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *SetHandler) GetSet(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(string)
	setID := r.PathValue("id")
	set, err := h.setRepo.GetSetByID(r.Context(), setID, userID)
	if err != nil {
		http.Error(w, "set not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(set)
}

func (h *SetHandler) GetSetsByDate(w http.ResponseWriter, r *http.Request) {
	UserID := r.Context().Value(middleware.UserIDKey).(string)
	dateStr := r.URL.Query().Get("date")
	if dateStr == "" {
		http.Error(w, "date is required", http.StatusBadRequest)
		return
	}
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		http.Error(w, "invalid date format, expected YYYY-MM-DD", http.StatusBadRequest)
		return
	}

	sets, err := h.setRepo.GetSetsByDate(r.Context(), UserID, date)
	if err != nil {
		http.Error(w, "failed to get sets", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(sets)
}
