package main

import (
	"log"
	"net/http"

	"gym_tracker/internal/config"
	"gym_tracker/internal/db"
	"gym_tracker/internal/handlers"
	"gym_tracker/internal/repository"
	"gym_tracker/internal/router"
)

func main() {
	cfg := config.Load()
	pool, err := db.NewPool(cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()
	log.Println("connected to database")

	// репозитории отделены от хендлеров, чтобы SQL не проникал в хендлеры
	userRepo := repository.NewUserRepository(pool)
	machineRepo := repository.NewMachineRepository(pool)
	workoutRepo := repository.NewWorkoutRepository(pool)
	workoutItemRepo := repository.NewWorkoutItemRepository(pool)
	setRepo := repository.NewSetRepository(pool)

	authHandler := handlers.NewAuthHandler(userRepo, cfg.JWTSecret)
	machineHandler := handlers.NewMachineHandler(machineRepo, "uploads")
	setHandler := handlers.NewSetHandler(workoutRepo, workoutItemRepo, setRepo, machineRepo)
	workoutHandler := handlers.NewWorkoutHandler(workoutRepo)
	userHandler := handlers.NewUserHandler(userRepo)

	mux := router.New(authHandler, machineHandler, setHandler, workoutHandler, userHandler, cfg.JWTSecret)

	log.Printf("starting server on port %s", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, mux); err != nil {
		log.Fatal(err)
	}
}
