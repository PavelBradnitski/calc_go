package services

import (
	"context"

	"github.com/PavelBradnitski/calc_go/internal/models"
	"github.com/PavelBradnitski/calc_go/internal/repositories"
)

type RateService struct {
	repo repositories.RateRepositoryInterface
}

type RateServiceInterface interface {
	Add(ctx context.Context, result float64) (int64, error)
	Get(ctx context.Context) ([]models.Expression, error)
	GetById(ctx context.Context, id int) (*models.Expression, error)
}

func NewRateService(repo repositories.RateRepositoryInterface) *RateService {
	return &RateService{repo: repo}
}

func (s *RateService) Add(ctx context.Context, result float64) (int64, error) {
	return s.repo.Add(ctx, result)
}

func (s *RateService) Get(ctx context.Context) ([]models.Expression, error) {
	return s.repo.Get(ctx)
}

func (s *RateService) GetById(ctx context.Context, id int) (*models.Expression, error) {
	return s.repo.GetById(ctx, id)
}
