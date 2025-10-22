package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"gorm.io/gorm"

	"case_study/internal/config"
	"case_study/internal/sender"
)

func SetupRouter(db *gorm.DB, sch *sender.Scheduler, cache *sender.RedisCache, cfg config.Config) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	h := NewHandler(db, sch, cache, cfg)

	r.Route("/api", func(r chi.Router) {
		r.Post("/start", h.Start)
		r.Post("/stop", h.Stop)
		r.Post("/sent-messages", h.SendMessage)

	})

	// Serve Swagger static YAML for documentation reference
	r.Get("/swagger.yaml", h.ServeSwagger)

	return r
}
