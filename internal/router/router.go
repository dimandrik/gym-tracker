package router

import (
	"gym_tracker/internal/handlers"
	"gym_tracker/internal/middleware"
	"net/http"
	"time"
)

func New(
	authHandler *handlers.AuthHandler,
	machineHandler *handlers.MachineHandler,
	setHandler *handlers.SetHandler,
	workoutHandler *handlers.WorkoutHandler,
	userHandler *handlers.UserHandler,
	jwtSecret string,
) http.Handler {
	mux := http.NewServeMux()

	loginLimiter := middleware.NewRateLimiter(5, time.Minute)
	registerLimiter := middleware.NewRateLimiter(5, time.Minute)

	// регистрация и логин — единственные маршруты без токена
	mux.HandleFunc("POST /api/login", middleware.RateLimitMiddleware(loginLimiter)(authHandler.Login))
	mux.HandleFunc("POST /api/register", middleware.RateLimitMiddleware(registerLimiter)(authHandler.Register))

	authMiddleware := middleware.AuthMiddleware(jwtSecret)
	mux.HandleFunc("POST /api/machines", authMiddleware(machineHandler.CreateMachine))
	mux.HandleFunc("GET /api/machines", authMiddleware(machineHandler.GetMachines))
	mux.HandleFunc("POST /api/sets", authMiddleware(setHandler.AddSet))
	mux.HandleFunc("GET /api/sets", authMiddleware(setHandler.GetSetsByMachine))
	mux.HandleFunc("GET /api/workouts/history", authMiddleware(workoutHandler.GetHistory))
	mux.HandleFunc("GET /api/machines/{id}", authMiddleware(machineHandler.GetMachine))
	mux.HandleFunc("DELETE /api/sets/{id}", authMiddleware(setHandler.DeleteSet))
	mux.HandleFunc("PUT /api/sets/{id}", authMiddleware(setHandler.UpdateSet))
	mux.HandleFunc("GET /api/sets/{id}", authMiddleware(setHandler.GetSet))
	mux.HandleFunc("PUT /api/machines/{id}", authMiddleware(machineHandler.UpdateMachine))
	mux.HandleFunc("DELETE /api/machines/{id}", authMiddleware(machineHandler.DeleteMachine))
	mux.HandleFunc("GET /api/workouts/day", authMiddleware(setHandler.GetSetsByDate))
	mux.HandleFunc("GET /api/user/profile", authMiddleware(userHandler.GetProfile))
	mux.HandleFunc("PUT /api/user/name", authMiddleware(userHandler.UpdateName))
	mux.HandleFunc("PUT /api/user/email", authMiddleware(userHandler.UpdateEmail))
	mux.HandleFunc("PUT /api/user/password", authMiddleware(userHandler.UpdatePassword))
	mux.HandleFunc("DELETE /api/user/account", authMiddleware(userHandler.DeleteAccount))
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	mux.Handle("/uploads/", http.StripPrefix("/uploads/", http.FileServer(http.Dir("uploads"))))

	return middleware.CORS(mux)
}
