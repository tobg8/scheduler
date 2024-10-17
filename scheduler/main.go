package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/tobg/scheduler/controllers"
	"github.com/tobg/scheduler/database"
	"github.com/tobg/scheduler/repositories"
	"github.com/tobg/scheduler/usecases"
)

// App represents our app structure
type App struct {
	Port               string
	RegisterController *controllers.RegisterController
}

func main() {
	app, err := initApp()
	if err != nil {
		log.Fatal("could not init application: %w", err)
	}

	app.SetupRoutes()
	err = app.Serve()
	if err != nil {
		log.Fatal("could not start application: %w", err)
	}
}

// initialization of app (port, controllers etc...)
func initApp() (*App, error) {
	if err := godotenv.Load(); err != nil {
		return nil, fmt.Errorf("error loading .env file: %w", err)
	}

	db, err := database.InitializeDB()
	if err != nil {
		return nil, fmt.Errorf("could not initialize database: %w", err)
	}

	rr := repositories.NewRegisterRepository(db)
	ru := usecases.NewRegisterUsecase(rr)
	rc := controllers.NewRegisterController(ru)

	err = rc.ReloadJobs()
	if err != nil {
		return nil, fmt.Errorf("could not reload jobs from database: %w", err)
	}

	return &App{
		Port:               os.Getenv("PORT"),
		RegisterController: rc,
	}, nil
}

// SetupRoutes binds the routes to the appropriate handlers
func (app *App) SetupRoutes() {
	http.Handle("/register", http.HandlerFunc(app.RegisterController.Register))
	http.Handle("/get-jobs", http.HandlerFunc(app.RegisterController.GetJobs))
}

// Graceful shutdown setup
func (app *App) Serve() error {
	srv := &http.Server{
		Addr:    app.Port,
		Handler: http.DefaultServeMux,
	}

	// Listen for OS signals for graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Run the server in a goroutine
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %s", err)
		}
	}()
	fmt.Printf("Listening on %s\n", app.Port)

	// Wait for an interrupt or terminate signal
	<-quit

	// Create a context with a timeout to ensure the server shuts down properly
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return srv.Shutdown(ctx)
}
