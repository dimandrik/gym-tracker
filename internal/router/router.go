package router

import (
	"gym_tracker/internal/handlers"
	"gym_tracker/internal/middleware"
	"net/http"
)

func New(
	authHandler *handlers.AuthHandler,
	machineHandler *handlers.MachineHandler,
	setHandler *handlers.SetHandler,
	workoutHandler *handlers.WorkoutHandler,
	jwtSecret string,
) http.Handler {
	mux := http.NewServeMux()

	// register/login are the only routes that don't require a token yet
	mux.HandleFunc("POST /api/register", authHandler.Register)
	mux.HandleFunc("POST /api/login", authHandler.Login)

	// everything below reads the user from the JWT, so it goes through authMiddleware
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

	// serves uploaded machine photos directly from disk
	mux.Handle("/uploads/", http.StripPrefix("/uploads/", http.FileServer(http.Dir("uploads"))))

	return middleware.CORS(mux)
}
