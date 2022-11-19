package main

import (
	"fmt"
	"github.com/jmoiron/sqlx"
	"time"
)

func NewDb(dsn string, connMaxLifetime time.Duration, maxOpenConns int, maxIdleConns int) (*sqlx.DB, error) {
	dbInstance, err := sqlx.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("DB initialization with specified DN or DSN: %w", err)
	}

	dbInstance.SetConnMaxLifetime(connMaxLifetime)
	dbInstance.SetMaxOpenConns(maxOpenConns)
	dbInstance.SetMaxIdleConns(maxIdleConns)

	err = dbInstance.Ping()
	if err != nil {
		return nil, fmt.Errorf("ping DB: %w", err)
	}
	return dbInstance, nil
}
