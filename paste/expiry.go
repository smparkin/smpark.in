package paste

import (
	"context"
	"log/slog"
	"time"
)

// StartExpiryWorker deletes expired pastes every interval until ctx is done.
// Reads also reject expired pastes, so this only bounds memory use.
func StartExpiryWorker(ctx context.Context, s *Store, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case now := <-ticker.C:
			if n := s.DeleteExpired(now); n > 0 {
				slog.Info("deleted expired pastes", "count", n)
			}
		}
	}
}
