package service

import (
	"context"

	platformtypes "github.com/agentregistry-dev/agentregistry/internal/registry/platforms/types"
	"github.com/agentregistry-dev/agentregistry/pkg/models"
	"github.com/agentregistry-dev/agentregistry/pkg/registry/database"
	apiv0 "github.com/modelcontextprotocol/registry/pkg/api/v0"
)

// APIRouteView provides the focused service surface used by the HTTP routing layer.
type APIRouteView struct {
	ServerRouteService
	AgentRouteService
	SkillService
	PromptService
	ProviderService
	DeploymentService
}

func NewAPIRouteView(server ServerRouteService, agent AgentRouteService, skill SkillService, prompt PromptService, provider ProviderService, deployment DeploymentService) *APIRouteView {
	return &APIRouteView{
		ServerRouteService: server,
		AgentRouteService:  agent,
		SkillService:       skill,
		PromptService:      prompt,
		ProviderService:    provider,
		DeploymentService:  deployment,
	}
}

func NewAPIRouteViewFromSet(set *Set) *APIRouteView {
	return NewAPIRouteView(set.Server(), set.Agent(), set.Skill(), set.Prompt(), set.Provider(), set.Deployment())
}

type mcpRegistryServerView interface {
	ListServers(ctx context.Context, filter *database.ServerFilter, cursor string, limit int) ([]*apiv0.ServerResponse, string, error)
	GetServerByNameAndVersion(ctx context.Context, serverName string, version string) (*apiv0.ServerResponse, error)
	GetAllVersionsByServerName(ctx context.Context, serverName string) ([]*apiv0.ServerResponse, error)
	GetServerReadmeLatest(ctx context.Context, serverName string) (*database.ServerReadme, error)
	GetServerReadmeByVersion(ctx context.Context, serverName, version string) (*database.ServerReadme, error)
}

type mcpRegistryAgentView interface {
	ListAgents(ctx context.Context, filter *database.AgentFilter, cursor string, limit int) ([]*models.AgentResponse, string, error)
	GetAgentByName(ctx context.Context, agentName string) (*models.AgentResponse, error)
	GetAgentByNameAndVersion(ctx context.Context, agentName string, version string) (*models.AgentResponse, error)
}

type mcpRegistrySkillView interface {
	ListSkills(ctx context.Context, filter *database.SkillFilter, cursor string, limit int) ([]*models.SkillResponse, string, error)
	GetSkillByName(ctx context.Context, skillName string) (*models.SkillResponse, error)
	GetSkillByNameAndVersion(ctx context.Context, skillName string, version string) (*models.SkillResponse, error)
}

type mcpRegistryDeploymentView interface {
	GetDeployments(ctx context.Context, filter *models.DeploymentFilter) ([]*models.Deployment, error)
	GetDeploymentByID(ctx context.Context, id string) (*models.Deployment, error)
	DeployServer(ctx context.Context, serverName, version string, config map[string]string, preferRemote bool, providerID string) (*models.Deployment, error)
	DeployAgent(ctx context.Context, agentName, version string, config map[string]string, preferRemote bool, providerID string) (*models.Deployment, error)
	UndeployDeployment(ctx context.Context, deployment *models.Deployment) error
}

// MCPRegistryView provides the focused service surface used by the MCP bridge.
type MCPRegistryView struct {
	mcpRegistryServerView
	mcpRegistryAgentView
	mcpRegistrySkillView
	mcpRegistryDeploymentView
}

func NewMCPRegistryView(server mcpRegistryServerView, agent mcpRegistryAgentView, skill mcpRegistrySkillView, deployment mcpRegistryDeploymentView) *MCPRegistryView {
	return &MCPRegistryView{
		mcpRegistryServerView:     server,
		mcpRegistryAgentView:      agent,
		mcpRegistrySkillView:      skill,
		mcpRegistryDeploymentView: deployment,
	}
}

func NewMCPRegistryViewFromSet(set *Set) *MCPRegistryView {
	return NewMCPRegistryView(set.Server(), set.Agent(), set.Skill(), set.Deployment())
}

type platformRuntimeProviderView interface {
	GetProviderByID(ctx context.Context, providerID string) (*models.Provider, error)
}

type platformRuntimeServerView interface {
	GetServerByNameAndVersion(ctx context.Context, serverName string, version string) (*apiv0.ServerResponse, error)
}

type platformRuntimeAgentView interface {
	GetAgentByNameAndVersion(ctx context.Context, agentName string, version string) (*models.AgentResponse, error)
	ResolveAgentManifestSkills(ctx context.Context, manifest *models.AgentManifest) ([]platformtypes.AgentSkillRef, error)
	ResolveAgentManifestPrompts(ctx context.Context, manifest *models.AgentManifest) ([]platformtypes.ResolvedPrompt, error)
}

// PlatformRuntimeView provides the focused service surface used by runtime materialization.
type PlatformRuntimeView struct {
	providers platformRuntimeProviderView
	servers   platformRuntimeServerView
	agents    platformRuntimeAgentView
}

func NewPlatformRuntimeView(provider platformRuntimeProviderView, server platformRuntimeServerView, agent platformRuntimeAgentView) *PlatformRuntimeView {
	return &PlatformRuntimeView{
		providers: provider,
		servers:   server,
		agents:    agent,
	}
}

func NewPlatformRuntimeViewFromSet(set *Set) *PlatformRuntimeView {
	return NewPlatformRuntimeView(set.Provider(), set.Server(), set.Agent())
}

func (v *PlatformRuntimeView) GetProviderByID(ctx context.Context, providerID string) (*models.Provider, error) {
	return v.providers.GetProviderByID(ctx, providerID)
}

func (v *PlatformRuntimeView) GetServerByNameAndVersion(ctx context.Context, serverName string, version string) (*apiv0.ServerResponse, error) {
	return v.servers.GetServerByNameAndVersion(ctx, serverName, version)
}

func (v *PlatformRuntimeView) GetAgentByNameAndVersion(ctx context.Context, agentName string, version string) (*models.AgentResponse, error) {
	return v.agents.GetAgentByNameAndVersion(ctx, agentName, version)
}

func (v *PlatformRuntimeView) ResolveAgentManifestSkills(ctx context.Context, manifest *models.AgentManifest) ([]platformtypes.AgentSkillRef, error) {
	return v.agents.ResolveAgentManifestSkills(ctx, manifest)
}

func (v *PlatformRuntimeView) ResolveAgentManifestPrompts(ctx context.Context, manifest *models.AgentManifest) ([]platformtypes.ResolvedPrompt, error) {
	return v.agents.ResolveAgentManifestPrompts(ctx, manifest)
}

type indexerServerView interface {
	ListServers(ctx context.Context, filter *database.ServerFilter, cursor string, limit int) ([]*apiv0.ServerResponse, string, error)
	GetServerEmbeddingMetadata(ctx context.Context, serverName, version string) (*database.SemanticEmbeddingMetadata, error)
	UpsertServerEmbedding(ctx context.Context, serverName, version string, embedding *database.SemanticEmbedding) error
}

type indexerAgentView interface {
	ListAgents(ctx context.Context, filter *database.AgentFilter, cursor string, limit int) ([]*models.AgentResponse, string, error)
	GetAgentEmbeddingMetadata(ctx context.Context, agentName, version string) (*database.SemanticEmbeddingMetadata, error)
	UpsertAgentEmbedding(ctx context.Context, agentName, version string, embedding *database.SemanticEmbedding) error
}

// IndexerRegistryView provides the focused service surface used by the embeddings indexer.
type IndexerRegistryView struct {
	servers indexerServerView
	agents  indexerAgentView
}

func NewIndexerRegistryView(server indexerServerView, agent indexerAgentView) *IndexerRegistryView {
	return &IndexerRegistryView{
		servers: server,
		agents:  agent,
	}
}

func NewIndexerRegistryViewFromSet(set *Set) *IndexerRegistryView {
	return NewIndexerRegistryView(set.Server(), set.Agent())
}

func (v *IndexerRegistryView) ListServers(ctx context.Context, filter *database.ServerFilter, cursor string, limit int) ([]*apiv0.ServerResponse, string, error) {
	return v.servers.ListServers(ctx, filter, cursor, limit)
}

func (v *IndexerRegistryView) GetServerEmbeddingMetadata(ctx context.Context, serverName, version string) (*database.SemanticEmbeddingMetadata, error) {
	return v.servers.GetServerEmbeddingMetadata(ctx, serverName, version)
}

func (v *IndexerRegistryView) UpsertServerEmbedding(ctx context.Context, serverName, version string, embedding *database.SemanticEmbedding) error {
	return v.servers.UpsertServerEmbedding(ctx, serverName, version, embedding)
}

func (v *IndexerRegistryView) ListAgents(ctx context.Context, filter *database.AgentFilter, cursor string, limit int) ([]*models.AgentResponse, string, error) {
	return v.agents.ListAgents(ctx, filter, cursor, limit)
}

func (v *IndexerRegistryView) GetAgentEmbeddingMetadata(ctx context.Context, agentName, version string) (*database.SemanticEmbeddingMetadata, error) {
	return v.agents.GetAgentEmbeddingMetadata(ctx, agentName, version)
}

func (v *IndexerRegistryView) UpsertAgentEmbedding(ctx context.Context, agentName, version string, embedding *database.SemanticEmbedding) error {
	return v.agents.UpsertAgentEmbedding(ctx, agentName, version, embedding)
}
