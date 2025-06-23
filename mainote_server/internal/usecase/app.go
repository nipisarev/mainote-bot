package usecase

import (
	"context"

	"mainote-server/internal/domain"
	"mainote-server/internal/repository"
)

// AppUsecase defines the interface for app business logic.
type AppUsecase interface {
	GetAllApps(ctx context.Context) ([]domain.App, error)
}

// NewAppUsecase creates a new instance of AppUsecase.
func NewAppUsecase(appRepo repository.AppRepository) AppUsecase {
	return &appUsecase{appRepo: appRepo}
}

type appUsecase struct {
	appRepo repository.AppRepository
}

// GetAllApps retrieves all available apps for customer installation.
func (uc *appUsecase) GetAllApps(ctx context.Context) ([]domain.App, error) {
	return uc.appRepo.GetAll(ctx)
}
