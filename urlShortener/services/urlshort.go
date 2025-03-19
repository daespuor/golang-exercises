package services

import (
	"context"
	"time"
	"urlShortener/repository"
)

type URLService struct {
	repo *repository.SQLiteURLRepository
}

func NewURLService(repo *repository.SQLiteURLRepository) URLService {
	return URLService{repo: repo}
}

func (s *URLService) GetAllMappings() ([]repository.URLMappingDTO, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return s.repo.List(ctx)
}
