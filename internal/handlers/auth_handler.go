package handlers

import (
	"encoding/json"
	"errors"
	"gym_tracker/internal/auth"
	"gym_tracker/internal/repository"
	"log"
	"net/http"
	"regexp"
	"strings"
)

var emailRegex = regexp.MustCompile(`^[^\s@]+@[^\s@]+\.[^\s@]+$`)

type AuthHandler struct {
	userRepo  *repository.UserRepository
	jwtSecret string
}

type RegisterRequest struct {
	Email     string `json:"email"`
	Password  string `json:"password"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func NewAuthHandler(userRepo *repository.UserRepository, jwtSecret string) *AuthHandler {
	return &AuthHandler{userRepo: userRepo, jwtSecret: jwtSecret}
}

// сразу возвращаем JWT, чтобы фронтенду не нужен был отдельный логин после регистрации
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := decodeJSON(w, r, &req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.FirstName == "" || req.LastName == "" {
		http.Error(w, "first name and last name are required", http.StatusBadRequest)
		return
	}
	if !emailRegex.MatchString(req.Email) {
		http.Error(w, "invalid email format", http.StatusBadRequest)
		return
	}
	if err := validatePassword(req.Password); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		serverError(w, err, "failed to hash password")
		return
	}

	userID, err := h.userRepo.CreateUser(r.Context(), req.Email, hash, req.FirstName, req.LastName)
	if err != nil {
		if errors.Is(err, repository.ErrEmailAlreadyExists) {
			http.Error(w, "email already exists", http.StatusConflict)
			return
		}
		serverError(w, err, "failed to create user")
		return
	}

	token, err := auth.GenerateToken(userID, h.jwtSecret)
	if err != nil {
		serverError(w, err, "failed to generate token")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(map[string]string{"token": token}); err != nil {
		log.Println("failed to encode response:", err)
	}
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := decodeJSON(w, r, &req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// одна и та же ошибка для неизвестного email и неверного пароля — чтобы не палить, какие email зарегистрированы
	user, err := h.userRepo.GetUserByEmail(r.Context(), req.Email)
	if err != nil {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	if err = auth.CheckPassword(req.Password, user.PasswordHash); err != nil {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	token, err := auth.GenerateToken(user.ID, h.jwtSecret)
	if err != nil {
		serverError(w, err, "failed to generate token")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(map[string]string{"token": token}); err != nil {
		log.Println("failed to encode response:", err)
	}
}

func validatePassword(password string) error {
	if len(password) < 8 {
		return errors.New("password must be at least 8 characters")
	}
	hasDigit := false
	hasSpecial := false
	for _, ch := range password {
		if ch >= '0' && ch <= '9' {
			hasDigit = true
		}
		if strings.ContainsRune("!@#$%^&*(),.?\":{}|<>", ch) {
			hasSpecial = true
		}
	}
	if !hasDigit {
		return errors.New("password must contain at least one digit")
	}
	if !hasSpecial {
		return errors.New("password must contain at least one special character")
	}
	return nil
}
