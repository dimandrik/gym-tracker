package handlers

import (
	"encoding/json"
	"errors"
	"gym_tracker/internal/auth"
	"gym_tracker/internal/middleware"
	"gym_tracker/internal/repository"
	"net/http"
)

type UserHandler struct {
	userRepo *repository.UserRepository
}

type ProfileResponse struct {
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

type UpdateNameRequest struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

type UpdateEmailRequest struct {
	NewEmail        string `json:"new_email"`
	CurrentPassword string `json:"current_password"`
}

type UpdatePasswordRequest struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

type DeleteAccountRequest struct {
	Password string `json:"password"`
}

func NewUserHandler(userRepo *repository.UserRepository) *UserHandler {
	return &UserHandler{userRepo: userRepo}
}

func (h *UserHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(string)

	user, err := h.userRepo.GetUserByID(r.Context(), userID)
	if err != nil {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}

	response := ProfileResponse{
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func (h *UserHandler) UpdateName(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(string)
	var req UpdateNameRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if req.FirstName == "" || req.LastName == "" {
		http.Error(w, "first_name and last_name are required", http.StatusBadRequest)
		return
	}
	if err := h.userRepo.UpdateName(r.Context(), userID, req.FirstName, req.LastName); err != nil {
		http.Error(w, "failed to update name", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *UserHandler) UpdateEmail(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(string)

	var req UpdateEmailRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if !emailRegex.MatchString(req.NewEmail) {
		http.Error(w, "invalid email format", http.StatusBadRequest)
		return
	}

	user, err := h.userRepo.GetUserByID(r.Context(), userID)
	if err != nil {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}

	if err := auth.CheckPassword(req.CurrentPassword, user.PasswordHash); err != nil {
		http.Error(w, "incorrect password", http.StatusUnauthorized)
		return
	}

	if err := h.userRepo.UpdateEmail(r.Context(), userID, req.NewEmail); err != nil {
		if errors.Is(err, repository.ErrEmailAlreadyExists) {
			http.Error(w, "email already exists", http.StatusConflict)
			return
		}
		http.Error(w, "failed to update email", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *UserHandler) UpdatePassword(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(string)

	var req UpdatePasswordRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.CurrentPassword == "" || req.NewPassword == "" {
		http.Error(w, "current_password and new_password are required", http.StatusBadRequest)
		return
	}
	user, err := h.userRepo.GetUserByID(r.Context(), userID)
	if err != nil {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}
	if err := auth.CheckPassword(req.CurrentPassword, user.PasswordHash); err != nil {
		http.Error(w, "incorrect password", http.StatusUnauthorized)
		return
	}
	if err := validatePassword(req.NewPassword); err != nil {
		http.Error(w, "invalid password format", http.StatusBadRequest)
		return
	}
	newHash, err := auth.HashPassword(req.NewPassword)
	if err != nil {
		http.Error(w, "failed to hash password", http.StatusInternalServerError)
		return
	}
	if err := h.userRepo.UpdatePassword(r.Context(), userID, newHash); err != nil {
		http.Error(w, "failed to update password", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *UserHandler) DeleteAccount(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(string)

	var req DeleteAccountRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	user, err := h.userRepo.GetUserByID(r.Context(), userID)
	if err != nil {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}

	if err := auth.CheckPassword(req.Password, user.PasswordHash); err != nil {
		http.Error(w, "incorrect password", http.StatusUnauthorized)
		return
	}

	if err := h.userRepo.DeleteUser(r.Context(), userID); err != nil {
		http.Error(w, "failed to delete account", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
