package sync

import (
	"context"
	"sync"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/pkg/errors"
	"go.uber.org/atomic"
)

var GroupStopping = errors.New("ErrGroup is stopping") // Stoppable is stopping signal

// Stoppable lifecycle stop manager interface
type Stoppable interface {
	WaitStoppable
	StopTriggerable

	// DoStop execute stop
	DoStop(f func())
	// Stopping listen stop is started
	Stopping() <-chan struct{}
	// IsStopping check stop is started
	IsStopping() bool
}

// StopTriggerable trigger stop interface
type StopTriggerable interface {
	// TriggerStop trigger stop
	TriggerStop()
	// StopTriggered listen stop is triggered
	StopTriggered() <-chan struct{}
}

// WaitStoppable wait stop completed
type WaitStoppable interface {
	WaitStopped()
}

var _ Stoppable = (*Stopper)(nil)

// Stopper implements Stoppable interface
type Stopper struct {
	_triggerLock  sync.Mutex
	stopTrigger   chan struct{} // the notification of stop triggered
	stopTriggered *atomic.Bool  // stop is triggered

	_stoppingLock sync.Mutex
	stoppingChan  chan struct{} // the notification of starting to stop
	isStopping    *atomic.Bool  // stop is started

	stoppedChan chan struct{} // the notification of stopping completed
	stopTimeout time.Duration // the timeout of stop
}

func NewStopper(stopTimeout time.Duration) *Stopper {
	return &Stopper{
		stopTrigger:   make(chan struct{}),
		stopTriggered: atomic.NewBool(false),

		stoppingChan: make(chan struct{}),
		isStopping:   atomic.NewBool(false),

		stoppedChan: make(chan struct{}),
		stopTimeout: stopTimeout,
	}
}
func (s *Stopper) DoStop(f func()) {
	if s.IsStopping() {
		return
	}

	func() {
		s._stoppingLock.Lock()
		defer s._stoppingLock.Unlock()

		if s.IsStopping() {
			return
		}

		close(s.stoppingChan)
		s.isStopping.Store(true)
	}()

	defer close(s.stoppedChan)

	ctx, cancel := context.WithTimeout(context.Background(), s.stopTimeout)
	defer cancel()

	done := make(chan struct{})
	go func() {
		f()
		close(done)
	}()

	select {
	case <-done:
	case <-ctx.Done():
		log.Errorf("[threading.Stopper.DoStop] Stop operation timed out after %.2fs", s.stopTimeout.Seconds())
	}
}

func (s *Stopper) TriggerStop() {
	if s.isStopTriggered() {
		return
	}

	s._triggerLock.Lock()
	defer s._triggerLock.Unlock()

	if s.isStopTriggered() {
		return
	}

	close(s.stopTrigger)
	s.stopTriggered.Store(true)
}

func (s *Stopper) isStopTriggered() bool {
	return s.stopTriggered.Load()
}

func (s *Stopper) StopTriggered() <-chan struct{} {
	return s.stopTrigger
}

func (s *Stopper) IsStopping() bool {
	return s.isStopping.Load()
}

func (s *Stopper) Stopping() <-chan struct{} {
	return s.stoppingChan
}

func (s *Stopper) WaitStopped() {
	<-s.stoppedChan
}
