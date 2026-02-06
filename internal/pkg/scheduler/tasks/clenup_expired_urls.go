package tasks

import (
	"context"

	"github.com/dzhordano/urlshortener/internal/core/domain/model"
	"github.com/dzhordano/urlshortener/internal/pkg/scheduler"
	"github.com/jackc/pgx/v5/pgxpool"
)

type CleanupExpiredURLsTask struct {
	db *pgxpool.Pool
}

// NewCleanupExpiredURLsTask returns cleanup task for expired urls.
// Deletes all non-valid entries (expired after model.ShortURLValidFor duration).
func NewCleanupExpiredURLsTask(
	db *pgxpool.Pool,
) scheduler.Task {
	return &CleanupExpiredURLsTask{
		db: db,
	}
}

func (t *CleanupExpiredURLsTask) Name() string {
	return "cleanup_expired_urls"
}

func (t *CleanupExpiredURLsTask) Execute(ctx context.Context) error {
	query := `
	DELETE FROM urls
	WHERE valid_until < NOW() - make_interval(secs => $1)
	`

	_, err := t.db.Exec(ctx, query, model.ShortURLValidFor.Seconds())
	if err != nil {
		return err
	}

	return nil
}
