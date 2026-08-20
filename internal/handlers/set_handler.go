package handlers

import (
	"encoding/json"
	"errors"
	"gym_tracker/internal/middleware"
	"gym_tracker/internal/repository"
	"net/http"
	"time"
)

type SetHandler struct {
	workoutRepo     *repository.WorkoutRepository
	workoutItemRepo *repository.WorkoutItemRepository
	setRepo         *repository.SetRepository
	machineRepo     *repository.MachineRepository
}

func NewSetHandler(workoutRepo *repository.WorkoutRepository, workoutItemRepo *repository.WorkoutItemRepository, setRepo *repository.SetRepository, machineRepo *repository.MachineRepository) *SetHandler {
	return &SetHandler{workoutRepo: workoutRepo, workoutItemRepo: workoutItemRepo, setRepo: setRepo, machineRepo: machineRepo}
}

type AddSetRequest struct {
	MachineID string  `json:"machine_id"`
	WeightKg  float64 `json:"weight_kg"`
	Reps      int     `json:"reps"`
	Date      string  `json:"date"`
}

type UpdateSetRequest struct {
	WeightKg float64 `json:"weight_kg"`
	Reps     int     `json:"reps"`
}

// подход создаёт тренировку и её элемент "на лету", если это первый подход по машине за день;
// нумерация подходов сбрасывается для каждой машины каждый день
func (h *SetHandler) AddSet(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(string)

	var req AddSetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if err := validateSetInput(req.WeightKg, req.Reps); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	today := time.Now().Truncate(24 * time.Hour)
	if req.Date != "" {
		parsedDate, err := time.Parse("2006-01-02", req.Date)
		if err != nil {
			http.Error(w, "invalid date format, expected YYYY-MM-DD", http.StatusBadRequest)
			return
		}
		today = parsedDate
	}

	if _, err := h.machineRepo.GetMachineByID(r.Context(), req.MachineID, userID); err != nil {
		http.Error(w, "machine not found", http.StatusNotFound)
		return
	}

	workoutID, err := h.workoutRepo.GetOrCreateWorkout(r.Context(), userID, today)
	if err != nil {
		serverError(w, err, "failed to get or create workout")
		return
	}

	itemID, err := h.workoutItemRepo.GetOrCreateWorkoutItem(r.Context(), workoutID, req.MachineID)
	if err != nil {
		serverError(w, err, "failed to get or create workout item")
		return
	}

	setNumber, err := h.setRepo.GetNextSetNumber(r.Context(), itemID)
	if err != nil {
		serverError(w, err, "failed to get set number")
		return
	}

	setID, err := h.setRepo.CreateSet(r.Context(), itemID, setNumber, req.WeightKg, req.Reps)
	if err != nil {
		serverError(w, err, "failed to create set")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{"id": setID, "set_number": setNumber})
}

func (h *SetHandler) GetSetsByMachine(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(string)
	machineID := r.URL.Query().Get("machine_id")
	if machineID == "" {
		http.Error(w, "machine_id is required", http.StatusBadRequest)
		return
	}
	sets, err := h.setRepo.GetSetsByMachineID(r.Context(), machineID, userID)
	if err != nil {
		serverError(w, err, "failed to get sets")
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

	if err := validateSetInput(req.WeightKg, req.Reps); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err := h.setRepo.UpdateSet(r.Context(), setID, userID, req.WeightKg, req.Reps)
	if err != nil {
		http.Error(w, "set not found", http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// верхние границы — не физический лимит, а защита от мусорных/ошибочных значений
// (опечатка на порядок, баг на фронте), реальные подходы никогда их не достигают
func validateSetInput(weightKg float64, reps int) error {
	if weightKg <= 0 {
		return errors.New("weight_kg must be greater than 0")
	}
	if weightKg > 1000 {
		return errors.New("weight_kg must not exceed 1000")
	}
	if reps <= 0 {
		return errors.New("reps must be greater than 0")
	}
	if reps > 1000 {
		return errors.New("reps must not exceed 1000")
	}
	return nil
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
		serverError(w, err, "failed to get sets")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(sets)
}
