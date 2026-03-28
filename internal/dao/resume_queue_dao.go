package dao

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ResumeQueueItem struct {
	JobID     uuid.UUID
	ApplyLink string
	Status    string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type ResumeQueueDAO interface {
	Enqueue(ctx context.Context, jobID uuid.UUID, applyLink string, status string) error
	ListByStatus(ctx context.Context, status string, limit int) ([]ResumeQueueItem, error)
	DeleteByJobID(ctx context.Context, jobID uuid.UUID) error
}

type PgxResumeQueueDAO struct {
	pool *pgxpool.Pool
}

func NewPgxResumeQueueDAO(pool *pgxpool.Pool) *PgxResumeQueueDAO {
	return &PgxResumeQueueDAO{pool: pool}
}

func (d *PgxResumeQueueDAO) Enqueue(ctx context.Context, jobID uuid.UUID, applyLink string, status string) error {
	_, err := d.pool.Exec(ctx, `
		INSERT INTO resume_generation_queue (job_id, apply_link, status)
		VALUES ($1, $2, $3)
		ON CONFLICT (job_id)
		DO UPDATE SET apply_link = EXCLUDED.apply_link, status = EXCLUDED.status, updated_at = NOW()
	`, jobID, applyLink, status)
	if err != nil {
		if isQueueApplyLinkConflict(err) {
			return fmt.Errorf("queue apply_link already exists: %w", err)
		}
		return err
	}
	return nil
}

func (d *PgxResumeQueueDAO) ListByStatus(ctx context.Context, status string, limit int) ([]ResumeQueueItem, error) {
	rows, err := d.pool.Query(ctx, `
		SELECT job_id, apply_link, status, created_at, updated_at
		FROM resume_generation_queue
		WHERE status = $1
		ORDER BY created_at ASC
		LIMIT $2
	`, status, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]ResumeQueueItem, 0, limit)
	for rows.Next() {
		var item ResumeQueueItem
		if err := rows.Scan(&item.JobID, &item.ApplyLink, &item.Status, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}

func (d *PgxResumeQueueDAO) DeleteByJobID(ctx context.Context, jobID uuid.UUID) error {
	commandTag, err := d.pool.Exec(ctx, `
		DELETE FROM resume_generation_queue
		WHERE job_id = $1
	`, jobID)
	if err != nil {
		return err
	}

	if commandTag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func isQueueApplyLinkConflict(err error) bool {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return false
	}
	return pgErr.Code == "23505" && pgErr.ConstraintName == "uq_resume_generation_queue_apply_link"
}
