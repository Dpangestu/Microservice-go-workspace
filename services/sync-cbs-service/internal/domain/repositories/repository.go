package repositories

import (
	"bkc_microservice/services/sync-cbs-service/internal/domain/entities"
	"context"
)

type SycroneCoreRepository interface {
	// CRUD operations
	Create(ctx context.Context, sc *entities.SycroneCore) error
	GetByUserID(ctx context.Context, userID string) (*entities.SycroneCore, error)
	GetByUserCore(ctx context.Context, userCore string) (*entities.SycroneCore, error)
	Update(ctx context.Context, sc *entities.SycroneCore) error
	Delete(ctx context.Context, userID string) error

	// Search & Pagination
	ListByStatus(ctx context.Context, status string, page, size int) ([]*entities.SycroneCore, int, error)
}
