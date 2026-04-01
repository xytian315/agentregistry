package service

import (
	"context"

	"github.com/agentregistry-dev/agentregistry/pkg/models"
)

type providerServiceImpl struct {
	*registryServiceImpl
}

var _ ProviderService = (*providerServiceImpl)(nil)

func (s *registryServiceImpl) providerService() *providerServiceImpl {
	return &providerServiceImpl{registryServiceImpl: s}
}

func (s *providerServiceImpl) readStores() storeBundle {
	return s.registryServiceImpl.readStores()
}

func (s *registryServiceImpl) ListProviders(ctx context.Context, platform *string) ([]*models.Provider, error) {
	return s.providerService().ListProviders(ctx, platform)
}

func (s *registryServiceImpl) GetProviderByID(ctx context.Context, providerID string) (*models.Provider, error) {
	return s.providerService().GetProviderByID(ctx, providerID)
}

func (s *registryServiceImpl) CreateProvider(ctx context.Context, in *models.CreateProviderInput) (*models.Provider, error) {
	return s.providerService().CreateProvider(ctx, in)
}

func (s *registryServiceImpl) UpdateProvider(ctx context.Context, providerID string, in *models.UpdateProviderInput) (*models.Provider, error) {
	return s.providerService().UpdateProvider(ctx, providerID, in)
}

func (s *registryServiceImpl) DeleteProvider(ctx context.Context, providerID string) error {
	return s.providerService().DeleteProvider(ctx, providerID)
}

// ProviderService defines provider lifecycle operations.
type ProviderService interface {
	// ListProviders retrieves deployment target providers, optionally filtered by provider platform type.
	ListProviders(ctx context.Context, platform *string) ([]*models.Provider, error)
	// GetProviderByID retrieves a provider by ID.
	GetProviderByID(ctx context.Context, providerID string) (*models.Provider, error)
	// CreateProvider creates a deployment target provider.
	CreateProvider(ctx context.Context, in *models.CreateProviderInput) (*models.Provider, error)
	// UpdateProvider updates mutable fields for a provider.
	UpdateProvider(ctx context.Context, providerID string, in *models.UpdateProviderInput) (*models.Provider, error)
	// DeleteProvider deletes a provider by ID.
	DeleteProvider(ctx context.Context, providerID string) error
}

// ListProviders lists providers, optionally filtered by platform.
func (s *providerServiceImpl) ListProviders(ctx context.Context, platform *string) ([]*models.Provider, error) {
	return s.readStores().providers.ListProviders(ctx, platform)
}

// GetProviderByID gets a provider by ID.
func (s *providerServiceImpl) GetProviderByID(ctx context.Context, providerID string) (*models.Provider, error) {
	return s.readStores().providers.GetProviderByID(ctx, providerID)
}

// CreateProvider creates a provider.
func (s *providerServiceImpl) CreateProvider(ctx context.Context, in *models.CreateProviderInput) (*models.Provider, error) {
	return s.readStores().providers.CreateProvider(ctx, in)
}

// UpdateProvider updates mutable provider fields.
func (s *providerServiceImpl) UpdateProvider(ctx context.Context, providerID string, in *models.UpdateProviderInput) (*models.Provider, error) {
	return s.readStores().providers.UpdateProvider(ctx, providerID, in)
}

// DeleteProvider removes a provider by ID.
func (s *providerServiceImpl) DeleteProvider(ctx context.Context, providerID string) error {
	return s.readStores().providers.DeleteProvider(ctx, providerID)
}
