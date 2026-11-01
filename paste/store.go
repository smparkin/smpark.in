package paste

import (
	"crypto/rand"
	"sync"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type Paste struct {
	ID           string    `json:"id"`
	Title        string    `json:"title"`
	Content      string    `json:"content"`
	Language     string    `json:"language"`
	CreatedAt    time.Time `json:"created_at"`
	ExpiresAt    time.Time `json:"expires_at"`
	PasswordHash []byte    `json:"-"`
	Protected    bool      `json:"protected"`
}

func (p *Paste) expired(now time.Time) bool {
	return now.After(p.ExpiresAt)
}

func (p *Paste) passwordMatches(pw string) bool {
	return pw != "" && bcrypt.CompareHashAndPassword(p.PasswordHash, []byte(pw)) == nil
}

type Store struct {
	mu     sync.RWMutex
	pastes map[string]*Paste
}

func NewStore() *Store {
	return &Store{pastes: make(map[string]*Paste)}
}

// Create assigns p a fresh ID and stores it.
func (s *Store) Create(p *Paste) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for {
		id := generateID()
		if _, taken := s.pastes[id]; !taken {
			p.ID = id
			s.pastes[id] = p
			return
		}
	}
}

func (s *Store) Get(id string) (*Paste, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	p, ok := s.pastes[id]
	return p, ok
}

func (s *Store) Delete(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.pastes, id)
}

// DeleteExpired removes every paste that expired before now and reports how many it removed.
func (s *Store) DeleteExpired(now time.Time) int {
	s.mu.Lock()
	defer s.mu.Unlock()
	n := 0
	for id, p := range s.pastes {
		if p.expired(now) {
			delete(s.pastes, id)
			n++
		}
	}
	return n
}

const base62Chars = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

func generateID() string {
	id := make([]byte, 0, 8)
	buf := make([]byte, 16)
	for len(id) < cap(id) {
		rand.Read(buf)
		for _, b := range buf {
			// Rejecting bytes >= 248 (62*4) keeps every character equally likely.
			if b < 248 && len(id) < cap(id) {
				id = append(id, base62Chars[b%62])
			}
		}
	}
	return string(id)
}
