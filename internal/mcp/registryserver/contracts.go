package registryserver

import (
	"context"

	"github.com/agentregistry-dev/agentregistry/pkg/models"
	"github.com/agentregistry-dev/agentregistry/pkg/registry/database"
	apiv0 "github.com/modelcontextprotocol/registry/pkg/api/v0"
)

// AgentRegistry defines the agent operations consumed by the MCP bridge.
type AgentRegistry interface {
	ListAgents(ctx context.Context, filter *database.AgentFilter, cursor string, limit int) ([]*models.AgentResponse, string, error)
	GetAgentByName(ctx context.Context, agentName string) (*models.AgentResponse, error)
	GetAgentByNameAndVersion(ctx context.Context, agentName string, version string) (*models.AgentResponse, error)
}

// ServerRegistry defines the server operations consumed by the MCP bridge.
type ServerRegistry interface {
	ListServers(ctx context.Context, filter *database.ServerFilter, cursor string, limit int) ([]*apiv0.ServerResponse, string, error)
	GetServerByNameAndVersion(ctx context.Context, serverName string, version string) (*apiv0.ServerResponse, error)
	GetAllVersionsByServerName(ctx context.Context, serverName string) ([]*apiv0.ServerResponse, error)
	GetServerReadmeLatest(ctx context.Context, serverName string) (*database.ServerReadme, error)
	GetServerReadmeByVersion(ctx context.Context, serverName, version string) (*database.ServerReadme, error)
}

// SkillRegistry defines the skill operations consumed by the MCP bridge.
type SkillRegistry interface {
	ListSkills(ctx context.Context, filter *database.SkillFilter, cursor string, limit int) ([]*models.SkillResponse, string, error)
	GetSkillByName(ctx context.Context, skillName string) (*models.SkillResponse, error)
	GetSkillByNameAndVersion(ctx context.Context, skillName string, version string) (*models.SkillResponse, error)
}

// DeploymentRegistry defines the deployment operations consumed by the MCP bridge.
type DeploymentRegistry interface {
	GetDeployments(ctx context.Context, filter *models.DeploymentFilter) ([]*models.Deployment, error)
	GetDeploymentByID(ctx context.Context, id string) (*models.Deployment, error)
	DeployServer(ctx context.Context, serverName, version string, config map[string]string, preferRemote bool, providerID string) (*models.Deployment, error)
	DeployAgent(ctx context.Context, agentName, version string, config map[string]string, preferRemote bool, providerID string) (*models.Deployment, error)
	UndeployDeployment(ctx context.Context, deployment *models.Deployment) error
}
