package data

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "modernc.org/sqlite" // Side effect import: Helps sql to know how to connect to sqlite
)

type DB interface {
	Connect(ctx context.Context) error
	Disconnect(ctx context.Context) error
	GetConn() *sql.DB
}

type SqLiteDB struct {
	db  *sql.DB
	dsn string
}

func NewSQLiteDB(dsn string) SqLiteDB {
	return SqLiteDB{dsn: dsn}
}

func (db *SqLiteDB) Connect(ctx context.Context) error {
	connection, err := sql.Open("sqlite", db.dsn)

	if err != nil {
		return fmt.Errorf("error connecting to the database: %w", err)
	}

	connection.SetMaxOpenConns(1)
	connection.SetMaxIdleConns(1)
	connection.SetConnMaxLifetime(1 * 5 * time.Minute)

	if err := connection.PingContext(ctx); err != nil {
		return fmt.Errorf("error pinging the database: %w", err)
	}

	db.db = connection

	return nil
}

func (db *SqLiteDB) Disconnect(ctx context.Context) error {
	if db.db == nil {
		return fmt.Errorf("error disconnecting from the database: no active connection found")
	}

	return db.db.Close()
}

func (db *SqLiteDB) GetConn() *sql.DB {
	return db.db
}
