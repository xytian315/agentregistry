package service

import (
	"context"

	"github.com/agentregistry-dev/agentregistry/pkg/registry/database"
	apiv0 "github.com/modelcontextprotocol/registry/pkg/api/v0"
)

// ServerService defines server catalog and mutation operations.
type ServerService interface {
	// ListServers retrieve all servers with optional filtering
	ListServers(ctx context.Context, filter *database.ServerFilter, cursor string, limit int) ([]*apiv0.ServerResponse, string, error)
	// GetServerByName retrieve latest version of a server by server name
	GetServerByName(ctx context.Context, serverName string) (*apiv0.ServerResponse, error)
	// GetServerByNameAndVersion retrieve specific version of a server by server name and version
	GetServerByNameAndVersion(ctx context.Context, serverName string, version string) (*apiv0.ServerResponse, error)
	// GetAllVersionsByServerName retrieve all versions of a server by server name
	GetAllVersionsByServerName(ctx context.Context, serverName string) ([]*apiv0.ServerResponse, error)
	// CreateServer creates a new server version
	CreateServer(ctx context.Context, req *apiv0.ServerJSON) (*apiv0.ServerResponse, error)
	// UpdateServer updates an existing server and optionally its status
	UpdateServer(ctx context.Context, serverName, version string, req *apiv0.ServerJSON, newStatus *string) (*apiv0.ServerResponse, error)
	// StoreServerReadme stores or updates the README for a server version
	StoreServerReadme(ctx context.Context, serverName, version string, content []byte, contentType string) error
	// GetServerReadmeLatest retrieves the README for the latest server version
	GetServerReadmeLatest(ctx context.Context, serverName string) (*database.ServerReadme, error)
	// GetServerReadmeByVersion retrieves the README for a specific server version
	GetServerReadmeByVersion(ctx context.Context, serverName, version string) (*database.ServerReadme, error)
	// DeleteServer permanently removes a server version from the registry
	DeleteServer(ctx context.Context, serverName, version string) error
	// UpsertServerEmbedding stores semantic embedding metadata for a server version
	UpsertServerEmbedding(ctx context.Context, serverName, version string, embedding *database.SemanticEmbedding) error
	// GetServerEmbeddingMetadata retrieves the embedding metadata for a server version
	GetServerEmbeddingMetadata(ctx context.Context, serverName, version string) (*database.SemanticEmbeddingMetadata, error)
}

// ServerRouteService defines the subset of server operations used by the HTTP routing layer.
type ServerRouteService interface {
	ListServers(ctx context.Context, filter *database.ServerFilter, cursor string, limit int) ([]*apiv0.ServerResponse, string, error)
	GetServerByName(ctx context.Context, serverName string) (*apiv0.ServerResponse, error)
	GetServerByNameAndVersion(ctx context.Context, serverName string, version string) (*apiv0.ServerResponse, error)
	GetAllVersionsByServerName(ctx context.Context, serverName string) ([]*apiv0.ServerResponse, error)
	CreateServer(ctx context.Context, req *apiv0.ServerJSON) (*apiv0.ServerResponse, error)
	UpdateServer(ctx context.Context, serverName, version string, req *apiv0.ServerJSON, newStatus *string) (*apiv0.ServerResponse, error)
	StoreServerReadme(ctx context.Context, serverName, version string, content []byte, contentType string) error
	GetServerReadmeLatest(ctx context.Context, serverName string) (*database.ServerReadme, error)
	GetServerReadmeByVersion(ctx context.Context, serverName, version string) (*database.ServerReadme, error)
	DeleteServer(ctx context.Context, serverName, version string) error
}
