package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"gym_tracker/internal/config"
	"gym_tracker/internal/db"
	"gym_tracker/internal/handlers"
	"gym_tracker/internal/middleware"
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

	allowedOrigins := middleware.ParseAllowedOrigins(cfg.AllowedOrigins)
	mux := router.New(authHandler, machineHandler, setHandler, workoutHandler, userHandler, cfg.JWTSecret, allowedOrigins)

	server := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: mux,
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	go func() {
		log.Printf("starting server on port %s", cfg.Port)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal(err)
		}
	}()

	<-quit
	log.Println("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("forced shutdown: %v", err)
	}

	log.Println("server stopped")
}
