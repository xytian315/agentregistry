package types

import (
	"context"

	"github.com/agentregistry-dev/agentregistry/pkg/models"
	apiv0 "github.com/modelcontextprotocol/registry/pkg/api/v0"
)

// ProviderRuntimeService defines the provider operations consumed by platform materialization.
type ProviderRuntimeService interface {
	GetProviderByID(ctx context.Context, providerID string) (*models.Provider, error)
}

// ServerRuntimeService defines the server operations consumed by platform materialization.
type ServerRuntimeService interface {
	GetServerByNameAndVersion(ctx context.Context, serverName string, version string) (*apiv0.ServerResponse, error)
}

// AgentRuntimeService defines the agent operations consumed by platform materialization.
type AgentRuntimeService interface {
	GetAgentByNameAndVersion(ctx context.Context, agentName string, version string) (*models.AgentResponse, error)
	ResolveAgentManifestSkills(ctx context.Context, manifest *models.AgentManifest) ([]AgentSkillRef, error)
	ResolveAgentManifestPrompts(ctx context.Context, manifest *models.AgentManifest) ([]ResolvedPrompt, error)
}
