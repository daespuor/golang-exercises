package repository

import (
	"context"
	"database/sql"
	"fmt"
)

type URLMappingDTO struct {
	ShortUrl string
	LongUrl  string
}

type URLRepository interface {
	List(ctx context.Context) ([]URLMappingDTO, error)
	Seed(ctx context.Context) error
}

type SQLiteURLRepository struct {
	db *sql.DB
}

func NewSQLiteURLRepository(db *sql.DB) SQLiteURLRepository {
	return SQLiteURLRepository{db: db}
}

const (
	createTableQuery = `
CREATE TABLE IF NOT EXISTS URL_MAPPING (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    shortUrl TEXT NOT NULL,
    longUrl TEXT NOT NULL
);`
	insertUrlMappingQuery = `
INSERT INTO URL_MAPPING (shortUrl, longUrl)
VALUES (?, ?)`
	dropTableQuery = `
DROP TABLE IF EXISTS URL_MAPPING;`
	selectAllQuery = `SELECT shortUrl, longUrl FROM URL_MAPPING`
)

func (repo SQLiteURLRepository) List(ctx context.Context) ([]URLMappingDTO, error) {
	result := make([]URLMappingDTO, 0)
	rows, err := repo.db.QueryContext(ctx, selectAllQuery)
	if err != nil {
		return nil, fmt.Errorf("error querying URL_MAPPING table: %w", err)
	}

	for rows.Next() {
		urlMappingDTO := URLMappingDTO{}
		if err = rows.Scan(&urlMappingDTO.ShortUrl, &urlMappingDTO.LongUrl); err != nil {
			return nil, fmt.Errorf("error scanning row: %w", err)
		}
		result = append(result, urlMappingDTO)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over rows: %w", err)
	}

	return result, nil
}

func (repo SQLiteURLRepository) Seed(ctx context.Context) error {
	dtos := []URLMappingDTO{
		{ShortUrl: "/urlshort", LongUrl: "https://github.com/gophercises/urlshort"},
		{ShortUrl: "/urlshort-final", LongUrl: "https://github.com/gophercises/urlshort/tree/solution"},
	}

	_, err := repo.db.Exec(dropTableQuery)
	if err != nil {
		return fmt.Errorf("error dropping the table URL_MAPPING %w", err)
	}

	_, err = repo.db.Exec(createTableQuery)
	if err != nil {
		return fmt.Errorf("error creating the table URL_MAPPING %w", err)
	}

	// Insert elements into the table
	tx, err := repo.db.Begin()
	if err != nil {
		return fmt.Errorf("error opening transaction %w", err)
	}

	for _, dto := range dtos {
		_, err = tx.Exec(insertUrlMappingQuery, dto.ShortUrl, dto.LongUrl)
		if err != nil {
			if err = tx.Rollback(); err != nil {
				return fmt.Errorf("error rolling back transaction %w", err)
			}
			return fmt.Errorf("error inserting element into URL_MAPPING %w", err)
		}
	}

	err = tx.Commit()
	if err != nil {
		if err = tx.Rollback(); err != nil {
			return fmt.Errorf("error rolling back transaction after commit failure: %w", err)
		}
		return fmt.Errorf("error committing transaction %w", err)
	}

	return nil
}
