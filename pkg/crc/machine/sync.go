package machine

import (
	"context"
	"errors"
	"time"

	"github.com/code-ready/crc/pkg/crc/logging"
	"github.com/code-ready/machine/libmachine/state"
	"golang.org/x/sync/semaphore"
)

const startCancelTimeout = 15 * time.Second

type Synchronized struct {
	startLock      *semaphore.Weighted
	stopDeleteLock *semaphore.Weighted
	startCancel    context.CancelFunc
	underlying     Client
}

func NewSynchronizedMachine(machine Client) *Synchronized {
	return &Synchronized{
		startLock:      semaphore.NewWeighted(1),
		stopDeleteLock: semaphore.NewWeighted(1),
		underlying:     machine,
	}
}

func (s *Synchronized) IsProgressing() bool {
	if !s.startLock.TryAcquire(1) {
		return true
	}
	defer s.startLock.Release(1)
	if !s.stopDeleteLock.TryAcquire(1) {
		return true
	}
	defer s.stopDeleteLock.Release(1)
	return false
}

func (s *Synchronized) Delete() error {
	cleanup, err := s.lockForStopOrDelete()
	defer cleanup()
	if err != nil {
		return err
	}
	return s.underlying.Delete()
}

func (s *Synchronized) Start(ctx context.Context, startConfig StartConfig) (*StartResult, error) {
	if !s.startLock.TryAcquire(1) {
		return nil, errors.New("cluster is starting")
	}
	defer s.startLock.Release(1)
	ctx, s.startCancel = context.WithCancel(ctx)
	return s.underlying.Start(ctx, startConfig)
}

func (s *Synchronized) Stop() (state.State, error) {
	cleanup, err := s.lockForStopOrDelete()
	defer cleanup()
	if err != nil {
		return state.Error, err
	}
	return s.underlying.Stop()
}

func (s *Synchronized) lockForStopOrDelete() (func(), error) {
	if !s.stopDeleteLock.TryAcquire(1) {
		return func() {

		}, errors.New("cluster is stopping or deleting")
	}

	if s.startCancel != nil {
		logging.Infof("Cancelling virtual machine start...")
		s.startCancel()
	}

	timeout, cancelFunc := context.WithTimeout(context.Background(), startCancelTimeout)
	defer cancelFunc()
	if err := s.startLock.Acquire(timeout, 1); err != nil {
		return func() {
			s.stopDeleteLock.Release(1)
		}, errors.New("cannot abort startup sequence quickly enough")
	}
	return func() {
		s.stopDeleteLock.Release(1)
		s.startLock.Release(1)
	}, nil
}

func (s *Synchronized) GetName() string {
	return s.underlying.GetName()
}
func (s *Synchronized) Exists() (bool, error) {
	return s.underlying.Exists()
}

func (s *Synchronized) GetConsoleURL() (*ConsoleResult, error) {
	return s.underlying.GetConsoleURL()
}

func (s *Synchronized) IP() (string, error) {
	return s.underlying.IP()
}

func (s *Synchronized) PowerOff() error {
	return s.underlying.PowerOff()
}

func (s *Synchronized) Status() (*ClusterStatusResult, error) {
	return s.underlying.Status()
}

func (s *Synchronized) IsRunning() (bool, error) {
	return s.underlying.IsRunning()
}
