package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/indienSs/tg-bot-go/internal/config"
	_ "github.com/lib/pq"
)

type Postgres struct {
	db *sql.DB
}

func New(cfg config.PostgresConfig) (*Postgres, error) {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open postgres: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping postgres: %w", err)
	}

	return &Postgres{db: db}, nil
}

func (p *Postgres) Close() error {
	return p.db.Close()
}

func (p *Postgres) SaveUser(ctx context.Context, telegramID int64, username, firstName, lastName string) error {
	query := `
		INSERT INTO users (telegram_id, username, first_name, last_name)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (telegram_id) DO UPDATE SET
			username = EXCLUDED.username,
			first_name = EXCLUDED.first_name,
			last_name = EXCLUDED.last_name
	`

	_, err := p.db.ExecContext(ctx, query, telegramID, username, firstName, lastName)
	return err
}

func (p *Postgres) SaveMessage(ctx context.Context, telegramID int64, text string) error {
	query := `
		INSERT INTO messages (user_id, text)
		VALUES ((SELECT id FROM users WHERE telegram_id = $1), $2)
	`

	_, err := p.db.ExecContext(ctx, query, telegramID, text)
	return err
}