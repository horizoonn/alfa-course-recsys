package catalog

import "github.com/horizoonn/alfa-course-recsys/backend/platform/postgres"

type Repository struct {
	executor postgres.ExecutorFunc
}

func NewRepository(executor postgres.ExecutorFunc) *Repository {
	return &Repository{
		executor: executor,
	}
}
