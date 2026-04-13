package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"room-booking/internal/config"
	"room-booking/internal/db"
	"room-booking/internal/db/repository/postgres"
	"room-booking/internal/infrastructure/custom_validator"
	"room-booking/internal/interface/http/router"
	"room-booking/internal/server"
	"room-booking/internal/service"

	govalidator "github.com/go-playground/validator/v10"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func main() {
	ctx := context.Background()

	cfg, _ := config.Load()

	logger := setupLogger(cfg.Env)

	validate := govalidator.New()
	custom_validator.Register(validate)

	pool, err := db.SetupDB(cfg.DatabaseURL(), cfg.DBMaxOpenConns, cfg.DBMaxIdleConns, cfg.DBConnMaxLifetime)
	if err != nil {
		logger.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer pool.Close()
	logger.Info("connected to database")

	jwtService := service.NewJWTService("super-secret")

	roomRepo := postgres.NewRoomRepository(pool)
	scheduleRepo := postgres.NewScheduleRepository(pool)
	slotRepo := postgres.NewSlotRepository(pool)
	bookingRepo := postgres.NewBookingRepository(pool)

	roomService := service.NewRoomService(roomRepo)
	scheduleService := service.NewScheduleService(roomRepo, scheduleRepo, slotRepo)
	slotService := service.NewSlotListService(roomRepo, slotRepo)
	bookingService := service.NewBookingCreateService(slotRepo, bookingRepo)

	mux := router.NewRouter(jwtService, roomService, scheduleService, slotService, bookingService, logger)
	srv := server.New(":8080", mux)

	go func() {
		if err := srv.Start(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	log.Println("Server started on :8080")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited gracefully")
}

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger

	switch env {
	case envLocal:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envDev:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envProd:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	default:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	}
	return log
}
