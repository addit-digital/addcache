package addcache

import (
	"errors"
	"strings"
	"sync"
	"time"
)

const (
	defaultDelimiter string        = ":"
	defaultCleanup   time.Duration = 30 * time.Second
)

var ErrCacheKeyNotFound = errors.New("exception.cache.key.not-found")

// Cache implementation core structure
type Cache interface {
	Set(key string, data any)
	SetEx(key string, data any, duration time.Duration)
	Get(key string) (any, error)
	Delete(key string)
	CreateKey(args ...string) string
	CreateKeyWithDelimiter(delimiter string, args ...string) string
	StopCleanup()
}

// local handling of cache implementation
type storage struct {
	stop chan struct{}
	wg   sync.WaitGroup
	mu   sync.RWMutex
	data map[string]storageData
}

type storageData struct {
	isPersistence  bool
	setTime        time.Time
	expireDuration time.Duration
	data           any
}

func NewCache() Cache {
	return NewCacheWithCleanup(defaultCleanup)
}

func NewCacheWithCleanup(cleanupInterval time.Duration) Cache {
	storage := storage{
		stop: make(chan struct{}),
		data: make(map[string]storageData),
	}

	storage.wg.Add(1)
	go func(cleanupInterval time.Duration) {
		defer storage.wg.Done()
		storage.cleanupLoop(cleanupInterval)
	}(cleanupInterval)

	return &storage
}

func (s *storage) Set(key string, data any) {
	s.data[key] = storageData{
		isPersistence:  true,
		setTime:        time.Now(),
		expireDuration: 0,
		data:           data,
	}
}

func (s *storage) SetEx(key string, data any, duration time.Duration) {
	s.data[key] = storageData{
		isPersistence:  false,
		setTime:        time.Now(),
		expireDuration: duration,
		data:           data,
	}
}

func (s *storage) Get(key string) (any, error) {
	if value, ok := s.data[key]; ok {
		if s.removeIfExpired(key, value) {
			return nil, ErrCacheKeyNotFound
		}
		return value.data, nil
	}
	return nil, ErrCacheKeyNotFound
}

func (s *storage) Delete(key string) {
	delete(s.data, key)
}

func (s *storage) CreateKey(args ...string) string {
	return s.CreateKeyWithDelimiter(defaultDelimiter, args...)
}

func (s *storage) CreateKeyWithDelimiter(delimiter string, args ...string) string {
	return strings.Join(args, delimiter)
}

func (s *storage) StopCleanup() {
	s.stop <- struct{}{}
}

func (s *storage) cleanupLoop(interval time.Duration) {
	t := time.NewTicker(interval)
	defer t.Stop()
	for {
		select {
		case <-s.stop:
			return
		case <-t.C:
			s.mu.Lock()
			for key, sd := range s.data {
				s.removeIfExpired(key, sd)
			}
			s.mu.Unlock()
		}
	}
}

func (s storage) removeIfExpired(key string, sd storageData) bool {
	if sd.isPersistence {
		return false
	}
	if sd.setTime.Add(sd.expireDuration).Unix() <= time.Now().Unix() {
		delete(s.data, key)
		return true
	}
	return false
}
