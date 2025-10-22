package sender

import (
	"context"
	"log"
	"strings"
	"time"

	"gorm.io/gorm"

	"case_study/internal/config"
	"case_study/internal/database"
)

type Scheduler struct {
	cfg       config.Config
	db        *gorm.DB
	repo      *database.MessageRepo
	cache     *RedisCache
	sender    *HTTPSender
	ticker    *time.Ticker
	quit      chan struct{}
	isRunning bool
}

func NewScheduler(db *gorm.DB, cache *RedisCache, cfg config.Config) *Scheduler {
	return &Scheduler{
		cfg:    cfg,
		db:     db,
		repo:   database.NewMessageRepo(db),
		cache:  cache,
		sender: NewHTTPSender(cfg),
		quit:   make(chan struct{}),
	}
}

func (s *Scheduler) Start() {
	if s.isRunning {
		return
	}
	s.isRunning = true
	s.ticker = time.NewTicker(time.Duration(s.cfg.TickerSeconds) * time.Second)
	go func() {
		log.Printf("scheduler started (every %ds)", s.cfg.TickerSeconds)
		// run immediately once on start
		s.runOnce()
		for {
			select {
			case <-s.ticker.C:
				s.runOnce()
			case <-s.quit:
				log.Println("scheduler stopped")
				return
			}
		}
	}()
}

func (s *Scheduler) Stop() {
	if !s.isRunning {
		return
	}
	s.isRunning = false
	if s.ticker != nil {
		s.ticker.Stop()
	}
	close(s.quit)
	s.quit = make(chan struct{}) // reset channel for next start
}

func (s *Scheduler) runOnce() {
	ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
	defer cancel()

	// Claim next batch
	msgs, err := s.repo.ClaimNextPending(ctx, s.cfg.BatchSize)
	if err != nil {
		log.Printf("claim error: %v", err)
		return
	}
	if len(msgs) == 0 {
		log.Printf("no pending messages")
		return
	}

	for _, m := range msgs {
		// Character limit enforcement
		if s.cfg.MsgCharLimit > 0 && len([]rune(m.Content)) > s.cfg.MsgCharLimit {
			err := s.repo.MarkFailed(ctx, m.ID, "content exceeds character limit")
			if err != nil {
				log.Printf("mark failed err: %v", err)
			}
			continue
		}

		messageID, code, body, err := s.sender.Send(ctx, m)
		if err != nil {
			// keep reason
			be := strings.TrimSpace(string(body))
			if be == "" {
				be = err.Error()
			}
			_ = s.repo.MarkFailed(ctx, m.ID, be)
			log.Printf("send failed id=%d code=%d err=%v body=%s", m.ID, code, err, string(body))
			continue
		}

		if err := s.repo.MarkSent(ctx, m.ID, messageID, time.Now()); err != nil {
			log.Printf("mark sent err: %v", err)
			continue
		}
		if s.cache != nil {
			_ = s.cache.SetSent(messageID, time.Now())
		}
		log.Printf("sent id=%d remoteId=%s", m.ID, messageID)
	}
}
