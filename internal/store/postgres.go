package store

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type URLRecord struct {
	ID        int64
	LongURL   string
	CreatedAt time.Time
}

type URLStore interface {
	InsertURL(ctx context.Context, longURL string) (int64, error)
	GetURLByID(ctx context.Context, id int64) (*URLRecord, error)
	Close()
}

type PostgresURLStore struct {
	pool *pgxpool.Pool
}

func NewPostgresURLStore(ctx context.Context, databaseURL string) (*PostgresURLStore, error) {
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to create pgx pool: %w", err)
	}
	return &PostgresURLStore{pool: pool}, nil
}

func (s *PostgresURLStore) Close() {
	s.pool.Close()
}

func (s *PostgresURLStore) InsertURL(ctx context.Context, longURL string) (int64, error) {
	var id int64
	err := s.pool.
		QueryRow(ctx, `INSERT INTO urls (long_url) VALUES ($1) RETURNING id`, longURL).
		Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("failed to insert url: %w", err)
	}
	return id, nil
}

func (s *PostgresURLStore) GetURLByID(ctx context.Context, id int64) (*URLRecord, error) {
	var rec URLRecord
	err := s.pool.
		QueryRow(ctx, `SELECT id, long_url, created_at FROM urls WHERE id = $1`, id).
		Scan(&rec.ID, &rec.LongURL, &rec.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to get url by id: %w", err)
	}
	return &rec, nil
}
