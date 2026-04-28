package paste

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type Handlers struct {
	store *Store
}

func NewHandlers(store *Store) *Handlers {
	return &Handlers{store: store}
}

type createRequest struct {
	Title    string `json:"title"`
	Content  string `json:"content"`
	Language string `json:"language"`
	Expiry   int    `json:"expiry"`
	Password string `json:"password"`
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func (h *Handlers) Create(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1_100_000)

	var req createRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Content == "" {
		writeError(w, http.StatusBadRequest, "content is required")
		return
	}
	if len(req.Content) > 1_000_000 {
		writeError(w, http.StatusRequestEntityTooLarge, "content exceeds 1,000,000 characters")
		return
	}
	if len(req.Title) > 200 {
		writeError(w, http.StatusBadRequest, "title exceeds 200 characters")
		return
	}

	var passwordHash string
	if req.Password != "" {
		if len(req.Password) > 72 {
			writeError(w, http.StatusBadRequest, "password exceeds 72 characters")
			return
		}
		hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to process password")
			return
		}
		passwordHash = string(hash)
	}

	p := &Paste{
		Title:        req.Title,
		Content:      req.Content,
		Language:     req.Language,
		PasswordHash: passwordHash,
		Protected:    passwordHash != "",
	}

	if req.Expiry < 1 || req.Expiry > 10 {
		writeError(w, http.StatusBadRequest, "expiry must be between 1 and 10 minutes")
		return
	}
	t := time.Now().Add(time.Duration(req.Expiry) * time.Minute)
	p.ExpiresAt = &t

	if err := h.store.Create(p); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create paste")
		return
	}

	writeJSON(w, http.StatusCreated, map[string]string{
		"id":  p.ID,
		"url": "/paste/" + p.ID,
	})
}

func (h *Handlers) Get(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	p, err := h.store.Get(id)
	if errors.Is(err, ErrNotFound) {
		writeError(w, http.StatusNotFound, "paste not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to retrieve paste")
		return
	}
	if p.ExpiresAt != nil && time.Now().After(*p.ExpiresAt) {
		h.store.Delete(id)
		writeError(w, http.StatusGone, "paste has expired")
		return
	}
	if p.Protected {
		pw := r.Header.Get("X-Paste-Password")
		if pw == "" || bcrypt.CompareHashAndPassword([]byte(p.PasswordHash), []byte(pw)) != nil {
			writeJSON(w, http.StatusForbidden, map[string]any{"protected": true, "error": "password required"})
			return
		}
	}
	writeJSON(w, http.StatusOK, p)
}

func (h *Handlers) GetRaw(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	p, err := h.store.Get(id)
	if errors.Is(err, ErrNotFound) {
		http.Error(w, "paste not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, "failed to retrieve paste", http.StatusInternalServerError)
		return
	}
	if p.ExpiresAt != nil && time.Now().After(*p.ExpiresAt) {
		h.store.Delete(id)
		http.Error(w, "paste has expired", http.StatusGone)
		return
	}
	if p.Protected {
		pw := r.Header.Get("X-Paste-Password")
		if pw == "" || bcrypt.CompareHashAndPassword([]byte(p.PasswordHash), []byte(pw)) != nil {
			http.Error(w, "password required", http.StatusForbidden)
			return
		}
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Write([]byte(p.Content))
}
