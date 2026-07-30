package service

import (
	"context"
	"sync"
	"time"

	"api-service/repository"
	logger "api-service/utils"
)

// SessionCleaner manages periodic background cleanup of expired admin sessions
type SessionCleaner struct {
	adminRepo repository.AdminRepository
	interval  time.Duration
	stopChan  chan struct{}
	wg        sync.WaitGroup
}

// NewSessionCleaner initializes a new SessionCleaner instance
func NewSessionCleaner(adminRepo repository.AdminRepository, interval time.Duration) *SessionCleaner {
	if interval <= 0 {
		interval = 30 * time.Minute
	}
	return &SessionCleaner{
		adminRepo: adminRepo,
		interval:  interval,
		stopChan:  make(chan struct{}),
	}
}

// Start launches the background cleanup worker goroutine
func (c *SessionCleaner) Start() {
	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		logger.Infof("[SessionCleaner] Background session cleaner worker started (Interval: %v)", c.interval)

		// Perform an initial cleanup pass on startup
		c.clean()

		ticker := time.NewTicker(c.interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				c.clean()
			case <-c.stopChan:
				logger.Info("[SessionCleaner] Background session cleaner worker stopped.")
				return
			}
		}
	}()
}

// Stop gracefully signals the cleaner goroutine to finish and waits for termination
func (c *SessionCleaner) Stop() {
	close(c.stopChan)
	c.wg.Wait()
}

func (c *SessionCleaner) clean() {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	deleted, err := c.adminRepo.CleanExpiredSessions(ctx)
	if err != nil {
		logger.Errorf("[SessionCleaner] Failed to clean expired sessions: %v", err)
		return
	}
	if deleted > 0 {
		logger.Infof("[SessionCleaner] Successfully cleaned %d expired session(s)", deleted)
	}
}
