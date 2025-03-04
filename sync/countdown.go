package sync

import (
	"errors"
	"sync"
	"time"
)

var ErrCountdownTimerExpired = errors.New("countdown timer expired")

// CountdownStopper the interface for countdown stoppers
type CountdownStopper interface {
	ExpiryTime() time.Time
	SetExpiryTime(time.Time)
	Reset()
}

var _ CountdownStopper = (*countdownTimer)(nil)

// countdownTimer implements the countdown stopper
type countdownTimer struct {
	sync.RWMutex
	expiryTime time.Time
}

// NewCountdownStopper creates a new countdown stopper
func NewCountdownStopper() CountdownStopper {
	return &countdownTimer{
		expiryTime: time.Time{},
	}
}

func (c *countdownTimer) ExpiryTime() time.Time {
	c.RLock()
	defer c.RUnlock()
	return c.expiryTime
}

func (c *countdownTimer) SetExpiryTime(newTime time.Time) {
	c.Lock()
	defer c.Unlock()
	c.expiryTime = newTime
}

func (c *countdownTimer) Reset() {
	c.Lock()
	defer c.Unlock()
	c.expiryTime = time.Time{}
}
