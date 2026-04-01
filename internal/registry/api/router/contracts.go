package router

import "github.com/agentregistry-dev/agentregistry/internal/registry/service"

// APIRouteService defines the registry operations consumed by the HTTP routing layer.
type APIRouteService interface {
	service.ServerRouteService
	service.AgentRouteService
	service.SkillService
	service.PromptService
	service.ProviderService
	service.DeploymentService
}
