package api

import (
	"case_study/internal/models"
	"context"
	"encoding/json"
	"net/http"
	"time"

	"gorm.io/gorm"

	"case_study/internal/config"
	"case_study/internal/database"
	"case_study/internal/sender"
)

type sendMessageReq struct {
	To      string `json:"to"`
	Content string `json:"content"`
}
type Handler struct {
	DB    *gorm.DB
	Repo  *database.MessageRepo
	Sched *sender.Scheduler
	Cfg   config.Config
	Cache *sender.RedisCache
}

func NewHandler(db *gorm.DB, sch *sender.Scheduler, cache *sender.RedisCache, cfg config.Config) *Handler {
	return &Handler{DB: db, Repo: database.NewMessageRepo(db), Sched: sch, Cfg: cfg, Cache: cache}
}
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// POST /api/sent-messages
func (h *Handler) SendMessage(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()

	var req sendMessageReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON"})
		return
	}
	if req.To == "" || req.Content == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "missing 'to' or 'content'"})
		return
	}
	if h.Cfg.MsgCharLimit > 0 && len([]rune(req.Content)) > h.Cfg.MsgCharLimit {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "content exceeds character limit"})
		return
	}

	// 1) create DB row as processing
	m, err := h.Repo.Create(ctx, req.To, req.Content, models.StatusProcessing)
	if err != nil {
		writeJSON(w, 500, map[string]string{"error": err.Error()})
		return
	}

	// 2) send to webhook
	s := sender.NewHTTPSender(h.Cfg)
	msgID, code, _, sendErr := s.Send(ctx, models.Message{ID: m.ID, To: req.To, Content: req.Content})
	if sendErr != nil {
		_ = h.Repo.MarkFailed(ctx, m.ID, sendErr.Error())
		writeJSON(w, code, map[string]string{"error": sendErr.Error()})
		return
	}

	// 3) mark sent
	_ = h.Repo.MarkSent(ctx, m.ID, msgID, time.Now())
	if h.Cache != nil {
		_ = h.Cache.SetSent(msgID, time.Now())
	}

	// 4) respond exactly as your spec/example
	writeJSON(w, http.StatusAccepted, map[string]any{
		"message":   "Accepted",
		"messageId": msgID,
	})
}

// POST /api/start
func (h *Handler) Start(w http.ResponseWriter, r *http.Request) {
	h.Sched.Start()
	writeJSON(w, http.StatusOK, map[string]string{"status": "started"})
}

// POST /api/stop
func (h *Handler) Stop(w http.ResponseWriter, r *http.Request) {
	h.Sched.Stop()
	writeJSON(w, http.StatusOK, map[string]string{"status": "stopped"})
}

// GET /swagger.yaml
func (h *Handler) ServeSwagger(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/yaml")
	http.ServeFile(w, r, "internal/docs/swagger.yaml")
}
