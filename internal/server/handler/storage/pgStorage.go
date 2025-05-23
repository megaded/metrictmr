package storage

import (
	"context"
	"database/sql"
	"errors"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/megaded/metrictmr/internal/data"
	"github.com/megaded/metrictmr/internal/logger"
	"github.com/megaded/metrictmr/internal/retry"
	"github.com/megaded/metrictmr/internal/server/handler/config"
)

const (
	CreateTable = `create table if not exists metrics(
	id uuid primary key default gen_random_uuid() ,
	name text not null,
	type text not null,
	delta bigint null,
	value double precision null,
	constraint metrics_name_type unique (name, type));`
)

type PgStorage struct {
	dbConnString string
	db           *sql.DB
	retry        retry.Retry
}

func NewPgStorage(ctx context.Context, cfg config.Config) *PgStorage {
	db, err := sql.Open("pgx", cfg.DBConnString)
	if err != nil {
		logger.Log.Fatal(err.Error())
	}
	ping := db.Ping()
	if ping != nil {
		logger.Log.Fatal(err.Error())
	}
	err = migrate(db)
	if err != nil {
		logger.Log.Fatal(err.Error())
	}
	go func() {
		defer db.Close()
		<-ctx.Done()
	}()

	return &PgStorage{dbConnString: cfg.DBConnString, db: db, retry: retry.NewRetry(1, 2, 3)}
}

func migrate(db *sql.DB) error {
	_, err := db.Exec(CreateTable)
	return err
}

func store(db *sql.DB, m ...data.Metric) error {
	if len(m) == 0 {
		return nil
	}
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	for _, v := range m {
		_, err = tx.Exec(`insert into metrics as m (name, type, delta, value) values ($1, $2, $3, $4) 
ON conflict(name ,type) do 
update set value = EXCLUDED.value, delta = m.delta + EXCLUDED.delta;`, v.ID, v.MType, v.Delta, v.Value)
		if err != nil {
			logger.Log.Info(err.Error())
			tx.Rollback()
			return err
		}
	}
	return tx.Commit()
}

func (s *PgStorage) GetGauge(name string) (metric data.Metric, exist bool, err error) {
	result := data.Metric{
		ID:    name,
		MType: gauge,
	}
	row := s.db.QueryRow(`select m.delta, m.value
	from metrics m 
	where m."name" =$1 and m."type" = $2;`, name, gauge)
	var value sql.NullFloat64
	var delta sql.NullInt64
	err = row.Scan(&delta, &value)
	if err != nil {
		logger.Log.Info(err.Error())
		if errors.Is(err, sql.ErrNoRows) {
			return result, false, nil
		}
		return result, false, err
	}
	result.Value = &value.Float64
	return result, true, nil
}
func (s *PgStorage) Store(ctx context.Context, metric ...data.Metric) error {
	return store(s.db, metric...)
}

func (s *PgStorage) GetCounter(name string) (metric data.Metric, exist bool, err error) {
	result := data.Metric{
		ID:    name,
		MType: counter,
	}
	row := s.db.QueryRow(`select m.delta, m.value
	from metrics m 
	where m."name" =$1 and m."type" = $2;`, name, counter)
	var value sql.NullFloat64
	var delta sql.NullInt64
	err = row.Scan(&delta, &value)
	if err != nil {
		logger.Log.Info(err.Error())
		if err == sql.ErrNoRows {
			return result, false, nil
		}
		return result, false, err
	}

	result.Delta = &delta.Int64
	return result, true, nil
}

func (s *PgStorage) GetMetrics() ([]data.Metric, error) {
	result := make([]data.Metric, 0)
	rows, err := s.db.Query(`select m.name, m.type, m.delta, m.value from metrics m;`)

	if err != nil {
		return result, err
	}
	defer rows.Close()
	for rows.Next() {
		var m data.Metric
		var value sql.NullFloat64
		var delta sql.NullInt64
		err = rows.Scan(&m.ID, &m.MType, &delta, &value)
		if err != nil {
			return nil, err
		}
		if m.MType == counter {
			m.Delta = &delta.Int64
		}
		if m.MType == gauge {
			m.Value = &value.Float64
		}

		result = append(result, m)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}
	return result, nil

}

func (s *PgStorage) HealthCheck() bool {
	err := s.db.Ping()
	return err == nil
}
