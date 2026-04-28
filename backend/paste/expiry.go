package paste

import (
	"context"
	"log/slog"
	"time"
)

func StartExpiryWorker(ctx context.Context, s *Store, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			deleteExpired(s)
		}
	}
}

func deleteExpired(s *Store) {
	var expired []string
	err := s.ForEach(func(p *Paste) error {
		if p.ExpiresAt != nil && time.Now().After(*p.ExpiresAt) {
			expired = append(expired, p.ID)
		}
		return nil
	})
	if err != nil {
		slog.Error("expiry scan error", "err", err)
		return
	}
	for _, id := range expired {
		if err := s.Delete(id); err != nil {
			slog.Error("expiry delete error", "id", id, "err", err)
		}
	}
	if len(expired) > 0 {
		slog.Info("deleted expired pastes", "count", len(expired))
	}
}
