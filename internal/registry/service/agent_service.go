package service

import (
	"context"

	api "github.com/agentregistry-dev/agentregistry/internal/registry/platforms/types"
	"github.com/agentregistry-dev/agentregistry/pkg/models"
	"github.com/agentregistry-dev/agentregistry/pkg/registry/database"
)

// AgentService defines agent catalog and mutation operations.
type AgentService interface {
	// ListAgents retrieve all agents with optional filtering
	ListAgents(ctx context.Context, filter *database.AgentFilter, cursor string, limit int) ([]*models.AgentResponse, string, error)
	// GetAgentByName retrieve latest version of an agent by name
	GetAgentByName(ctx context.Context, agentName string) (*models.AgentResponse, error)
	// GetAgentByNameAndVersion retrieve specific version of an agent by name and version
	GetAgentByNameAndVersion(ctx context.Context, agentName string, version string) (*models.AgentResponse, error)
	// GetAllVersionsByAgentName retrieve all versions of an agent by name
	GetAllVersionsByAgentName(ctx context.Context, agentName string) ([]*models.AgentResponse, error)
	// CreateAgent creates a new agent version
	CreateAgent(ctx context.Context, req *models.AgentJSON) (*models.AgentResponse, error)
	// ResolveAgentManifestSkills resolves manifest skill refs to concrete image or repo refs.
	ResolveAgentManifestSkills(ctx context.Context, manifest *models.AgentManifest) ([]api.AgentSkillRef, error)
	// ResolveAgentManifestPrompts resolves manifest prompt refs to concrete prompt content.
	ResolveAgentManifestPrompts(ctx context.Context, manifest *models.AgentManifest) ([]api.ResolvedPrompt, error)
	// DeleteAgent permanently removes an agent version from the registry
	DeleteAgent(ctx context.Context, agentName, version string) error
	// UpsertAgentEmbedding stores semantic embedding metadata for an agent version
	UpsertAgentEmbedding(ctx context.Context, agentName, version string, embedding *database.SemanticEmbedding) error
	// GetAgentEmbeddingMetadata retrieves the embedding metadata for an agent version
	GetAgentEmbeddingMetadata(ctx context.Context, agentName, version string) (*database.SemanticEmbeddingMetadata, error)
}

// AgentRouteService defines the subset of agent operations used by the HTTP routing layer.
type AgentRouteService interface {
	ListAgents(ctx context.Context, filter *database.AgentFilter, cursor string, limit int) ([]*models.AgentResponse, string, error)
	GetAgentByName(ctx context.Context, agentName string) (*models.AgentResponse, error)
	GetAgentByNameAndVersion(ctx context.Context, agentName string, version string) (*models.AgentResponse, error)
	GetAllVersionsByAgentName(ctx context.Context, agentName string) ([]*models.AgentResponse, error)
	CreateAgent(ctx context.Context, req *models.AgentJSON) (*models.AgentResponse, error)
	DeleteAgent(ctx context.Context, agentName, version string) error
}
