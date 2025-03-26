package storage

import (
	"database/sql"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/megaded/metrictmr/internal/data"
	"github.com/megaded/metrictmr/internal/logger"
	"github.com/megaded/metrictmr/internal/server/handler/config"
)

const (
	CreateTable = `create table if not exists metrics(
	id uuid primary key default gen_random_uuid() ,
	name text,
	type text,
	delta int,
	value double precision,
	constraint metrics_name_type unique (name, type));`
)

type PgStorage struct {
	fStorage     FileStorage
	dbConnString string
	db           *sql.DB
}

func NewPgStorage(cfg config.Config) *PgStorage {
	db, err := sql.Open("pgx", cfg.DbConnString)
	if err != nil {
		logger.Log.Info(err.Error())
	}
	ping := db.Ping()
	if ping != nil {
		logger.Log.Fatal(err.Error())
	}
	err = migrate(db)
	if err != nil {
		logger.Log.Fatal(err.Error())
	}

	return &PgStorage{fStorage: *NewFileStorage(cfg), dbConnString: cfg.DbConnString, db: db}
}

func migrate(db *sql.DB) error {
	_, err := db.Exec(CreateTable)
	return err
}

func (s *PgStorage) GetGauge(name string) (metric data.Metric, exist bool) {
	return s.fStorage.GetGauge(name)
}
func (s *PgStorage) Store(metric data.Metric) {
	s.fStorage.Store(metric)
}

func (s *PgStorage) GetCounter(name string) (metric data.Metric, exist bool) {
	return s.fStorage.GetCounter(name)
}

func (s *PgStorage) GetGaugeMetrics() []data.Metric {
	return s.fStorage.GetGaugeMetrics()
}
func (s *PgStorage) GetCounterMetrics() []data.Metric {
	return s.fStorage.GetCounterMetrics()
}
func (s *PgStorage) HealthCheck() bool {
	err := s.db.Ping()
	return err == nil
}
