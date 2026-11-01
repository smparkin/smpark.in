package paste

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"
	"unicode/utf8"

	"golang.org/x/crypto/bcrypt"
)

const (
	maxContentBytes  = 1_000_000
	maxTitleChars    = 200
	maxPasswordBytes = 72 // bcrypt ignores anything past this

	// JSON escaping can inflate a max-size paste well past 1 MB on the wire.
	maxBodyBytes = 8 << 20
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
	r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)

	var req createRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		var tooLarge *http.MaxBytesError
		if errors.As(err, &tooLarge) {
			writeError(w, http.StatusRequestEntityTooLarge, "content exceeds 1 MB")
			return
		}
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	switch {
	case req.Content == "":
		writeError(w, http.StatusBadRequest, "content is required")
		return
	case len(req.Content) > maxContentBytes:
		writeError(w, http.StatusRequestEntityTooLarge, "content exceeds 1 MB")
		return
	case utf8.RuneCountInString(req.Title) > maxTitleChars:
		writeError(w, http.StatusBadRequest, "title exceeds 200 characters")
		return
	case req.Expiry < 1 || req.Expiry > 10:
		writeError(w, http.StatusBadRequest, "expiry must be between 1 and 10 minutes")
		return
	case len(req.Password) > maxPasswordBytes:
		writeError(w, http.StatusBadRequest, "password exceeds 72 bytes")
		return
	}

	if req.Language == "" {
		req.Language = "plaintext"
	}

	now := time.Now().UTC()
	p := &Paste{
		Title:     req.Title,
		Content:   req.Content,
		Language:  req.Language,
		CreatedAt: now,
		ExpiresAt: now.Add(time.Duration(req.Expiry) * time.Minute),
	}

	if req.Password != "" {
		hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to process password")
			return
		}
		p.PasswordHash = hash
		p.Protected = true
	}

	h.store.Create(p)

	writeJSON(w, http.StatusCreated, map[string]string{
		"id":  p.ID,
		"url": "/paste/" + p.ID,
	})
}

func (h *Handlers) Get(w http.ResponseWriter, r *http.Request) {
	p, status, msg := h.load(r)
	switch status {
	case http.StatusOK:
		writeJSON(w, http.StatusOK, p)
	case http.StatusForbidden:
		writeJSON(w, status, map[string]any{"protected": true, "error": msg})
	default:
		writeError(w, status, msg)
	}
}

func (h *Handlers) GetRaw(w http.ResponseWriter, r *http.Request) {
	p, status, msg := h.load(r)
	if status != http.StatusOK {
		http.Error(w, msg, status)
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Write([]byte(p.Content))
}

// load looks up the paste named in the URL, enforcing expiry and the password
// header. When it can't return the paste it returns the status and message to send.
func (h *Handlers) load(r *http.Request) (*Paste, int, string) {
	id := r.PathValue("id")
	p, ok := h.store.Get(id)
	if !ok {
		return nil, http.StatusNotFound, "paste not found"
	}
	if p.expired(time.Now()) {
		h.store.Delete(id)
		return nil, http.StatusGone, "paste has expired"
	}
	if p.Protected && !p.passwordMatches(r.Header.Get("X-Paste-Password")) {
		return nil, http.StatusForbidden, "password required"
	}
	return p, http.StatusOK, ""
}
