package main

import (
	"case_study/internal/api"
	"case_study/internal/config"
	"case_study/internal/database"
	"case_study/internal/sender"
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	cfg := config.Load()

	// DB connect
	db, err := database.Connect(cfg)
	if err != nil {
		log.Fatalf("db connect error: %v", err)
	}

	// Redis
	var redisCache *sender.RedisCache
	if cfg.RedisEnabled {
		redisCache, err = sender.NewRedisCache(cfg)
		if err != nil {
			log.Fatalf("redis connect error: %v", err)
		}
	}

	// Scheduler
	sch := sender.NewScheduler(db, redisCache, cfg)
	// Start automatically on boot as required
	sch.Start()

	r := api.SetupRouter(db, sch, redisCache, cfg)

	srv := &http.Server{
		Addr:              ":" + cfg.HTTPPort,
		Handler:           r,
		ReadTimeout:       15 * time.Second,
		ReadHeaderTimeout: 15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	go func() {
		log.Printf("HTTP server listening on :%s", cfg.HTTPPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	// shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = srv.Shutdown(ctx)
	sch.Stop()
	if redisCache != nil {
		_ = redisCache.Close()
	}
}
