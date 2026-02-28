package paste

import (
	"crypto/rand"
	"errors"
	"sync"
	"time"
)

var ErrNotFound = errors.New("paste not found")

type Paste struct {
	ID        string     `json:"id"`
	Title     string     `json:"title"`
	Content   string     `json:"content"`
	Language  string     `json:"language"`
	CreatedAt time.Time  `json:"created_at"`
	ExpiresAt *time.Time `json:"expires_at"`
}

type Store struct {
	mu     sync.RWMutex
	pastes map[string]*Paste
}

func NewStore() *Store {
	return &Store{pastes: make(map[string]*Paste)}
}

func (s *Store) Create(p *Paste) error {
	id, err := generateID()
	if err != nil {
		return err
	}
	p.ID = id
	p.CreatedAt = time.Now().UTC()

	s.mu.Lock()
	s.pastes[id] = p
	s.mu.Unlock()
	return nil
}

func (s *Store) Get(id string) (*Paste, error) {
	s.mu.RLock()
	p, ok := s.pastes[id]
	s.mu.RUnlock()
	if !ok {
		return nil, ErrNotFound
	}
	return p, nil
}

func (s *Store) Delete(id string) error {
	s.mu.Lock()
	delete(s.pastes, id)
	s.mu.Unlock()
	return nil
}

func (s *Store) ForEach(fn func(p *Paste) error) error {
	s.mu.RLock()
	ids := make([]string, 0, len(s.pastes))
	for id := range s.pastes {
		ids = append(ids, id)
	}
	s.mu.RUnlock()

	for _, id := range ids {
		s.mu.RLock()
		p, ok := s.pastes[id]
		s.mu.RUnlock()
		if !ok {
			continue
		}
		if err := fn(p); err != nil {
			return err
		}
	}
	return nil
}

const base62Chars = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

func generateID() (string, error) {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	result := make([]byte, 8)
	for i, v := range b {
		result[i] = base62Chars[int(v)%62]
	}
	return string(result), nil
}
