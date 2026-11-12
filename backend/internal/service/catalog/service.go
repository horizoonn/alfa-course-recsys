package catalog

import (
	"context"

	"github.com/horizoonn/alfa-course-recsys/backend/internal/domain"
)

type Repository interface {
	ListArticles(context.Context) ([]domain.Material, error)
	ListCourses(context.Context) ([]domain.Material, error)
	ListDocuments(context.Context) ([]domain.Material, error)
	ListTerms(context.Context) ([]domain.Material, error)
	ListCalculators(context.Context) ([]domain.Material, error)
	ListPodcasts(context.Context) ([]domain.Material, error)
	ListWebinars(context.Context) ([]domain.Material, error)
}

type TxManager interface {
	WithinTransaction(context.Context, func(context.Context) error) error
}

type Service struct {
	repository Repository
	txManager  TxManager
}

func NewService(repository Repository, txManager TxManager) *Service {
	return &Service{
		repository: repository,
		txManager:  txManager,
	}
}
