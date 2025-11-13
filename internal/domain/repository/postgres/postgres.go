package postgres

import (
	"context"
	"fmt"
	"pr-service/config"
	"time"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"go.uber.org/zap"
)

type Repository struct {
	ctx context.Context
	log *zap.Logger
	cfg *config.ConfigModel
	DB  PgxPool
}

type PgxPool interface {
	Exec(context.Context, string, ...interface{}) (pgconn.CommandTag, error)
	Query(context.Context, string, ...interface{}) (pgx.Rows, error)
	QueryRow(context.Context, string, ...interface{}) pgx.Row
	Begin(context.Context) (pgx.Tx, error)
	Close()
}

func NewRepository(log *zap.Logger, cfg *config.ConfigModel, ctx context.Context) (*Repository, error) {
	return &Repository{ctx: ctx, log: log, cfg: cfg}, nil
}

func (r *Repository) OnStart(_ context.Context) error {
	u := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		r.cfg.Postgres.Host, r.cfg.Postgres.Port, r.cfg.Postgres.User,
		r.cfg.Postgres.Password, r.cfg.Postgres.DBName, r.cfg.Postgres.SSLMode)
	var pool *pgxpool.Pool
	var err error
	for i := 0; i < 5; i++ {
		pool, err = pgxpool.Connect(r.ctx, u)
		if err == nil {
			r.DB = pool
			return nil
		}
		r.log.Warn("db retry", zap.Int("try", i+1), zap.Error(err))
		time.Sleep(2 * time.Second)
	}
	return err
}

func (r *Repository) OnStop(_ context.Context) error {
	if r.DB != nil {
		r.DB.Close()
	}
	return nil
}
