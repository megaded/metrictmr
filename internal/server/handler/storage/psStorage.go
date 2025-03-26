package storage

import (
	"database/sql"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/megaded/metrictmr/internal/data"
	"github.com/megaded/metrictmr/internal/logger"
	migration "github.com/megaded/metrictmr/internal/server/db"
	"github.com/megaded/metrictmr/internal/server/handler/config"
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
	_, err := db.Exec(migration.CreateTable)
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
