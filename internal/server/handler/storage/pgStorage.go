package storage

import (
	"database/sql"
	"fmt"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/megaded/metrictmr/internal/data"
	"github.com/megaded/metrictmr/internal/logger"
	"github.com/megaded/metrictmr/internal/server/handler/config"
)

const (
	CreateTable = `create table if not exists metrics(
	id uuid primary key default gen_random_uuid() ,
	name text not null,
	type text not null,
	delta int null,
	value double precision null,
	constraint metrics_name_type unique (name, type));`

	Upsert = `insert into metrics as m (name, type, delta, value) values ($1, $2, $3, $4) 
ON conflict(name ,type) do 
update set value = EXCLUDED.value, delta = m.delta + EXCLUDED.delta
where m.name = EXCLUDED.name and m.type = EXCLUDED.type;`
	Select = `select m.delta, m.value
   from metrics m 
   where m."name" =$1 and m."type" = $2;`
)

type PgStorage struct {
	fStorage     FileStorage
	dbConnString string
	db           *sql.DB
}

func NewPgStorage(cfg config.Config) *PgStorage {
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

	return &PgStorage{fStorage: *NewFileStorage(cfg), dbConnString: cfg.DBConnString, db: db}
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
		tx.Exec(Upsert, v.ID, v.MType, v.Delta, v.Value)
	}
	return tx.Commit()
}

func (s *PgStorage) GetGauge(name string) (metric data.Metric, exist bool) {
	result := data.Metric{
		ID:    name,
		MType: gauge,
	}
	row := s.db.QueryRow(Select, name, gauge)
	if row == nil {
		fmt.Println("nul")
		return result, false
	}
	var value sql.NullFloat64
	var delta sql.NullInt64
	err := row.Scan(&delta, &value)
	if err != nil {
		logger.Log.Info(err.Error())
		return result, false
	}
	result.Value = &value.Float64
	return result, true
}
func (s *PgStorage) Store(metric data.Metric) {
	store(s.db, metric)
}

func (s *PgStorage) GetCounter(name string) (metric data.Metric, exist bool) {
	result := data.Metric{
		ID:    name,
		MType: gauge,
	}
	row := s.db.QueryRow(Select, name, counter)
	if row == nil {
		fmt.Println("nul")
		return result, false
	}
	var value sql.NullFloat64
	var delta sql.NullInt64
	err := row.Scan(&delta, &value)
	if err != nil {
		logger.Log.Info(err.Error())
		return result, false
	}
	result.Delta = &delta.Int64
	return result, true
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
