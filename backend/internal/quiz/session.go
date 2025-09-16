package quiz

import (
	"sync"
	"time"
)

type Session struct {
	ID            string
	PokemonID     int
	PokemonName   string
	DisplayName   string
	AcceptAnswers []string
	RegionKey     string
	Types         []string
	StartedAt     time.Time
	LastGuessAt   time.Time
	Solved        bool
	GaveUp        bool
	AllowMega     bool
	AllowPrimal   bool
}

type Store struct {
	mu sync.RWMutex
	m  map[string]*Session
}

func NewStore() *Store { return &Store{m: make(map[string]*Session)} }

func (s *Store) Get(id string) (*Session, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	v, ok := s.m[id]
	return v, ok
}

func (s *Store) Set(sess *Session) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.m[sess.ID] = sess
}
