package service

import (
	"context"
	"testing"

	"github.com/agentregistry-dev/agentregistry/pkg/models"
	"github.com/agentregistry-dev/agentregistry/pkg/registry/database"
	apiv0 "github.com/modelcontextprotocol/registry/pkg/api/v0"
	"github.com/stretchr/testify/require"
)

type storeTestTx struct {
	database.Transaction
	token *int
}

type storeTestStore struct {
	database.Store
	db *storeTestDB
	tx database.Transaction
}

type storeTestDB struct {
	database.ServiceDatabase
	testingT         *testing.T
	inTransaction    bool
	listServersFn    func(ctx context.Context, tx database.Transaction, filter *database.ServerFilter, cursor string, limit int) ([]*apiv0.ServerResponse, string, error)
	listAgentsFn     func(ctx context.Context, tx database.Transaction, filter *database.AgentFilter, cursor string, limit int) ([]*models.AgentResponse, string, error)
	listSkillsFn     func(ctx context.Context, tx database.Transaction, filter *database.SkillFilter, cursor string, limit int) ([]*models.SkillResponse, string, error)
	listPromptsFn    func(ctx context.Context, tx database.Transaction, filter *database.PromptFilter, cursor string, limit int) ([]*models.PromptResponse, string, error)
	listProvidersFn  func(ctx context.Context, tx database.Transaction, platform *string) ([]*models.Provider, error)
	getDeploymentsFn func(ctx context.Context, tx database.Transaction, filter *models.DeploymentFilter) ([]*models.Deployment, error)
	deleteServerFn   func(ctx context.Context, tx database.Transaction, serverName, version string) error
	deleteAgentFn    func(ctx context.Context, tx database.Transaction, agentName, version string) error
}

var _ database.ServiceDatabase = (*storeTestDB)(nil)

func (m *storeTestDB) asStore(tx database.Transaction) storeTestStore {
	return storeTestStore{db: m, tx: tx}
}

func (m *storeTestDB) InTransaction(ctx context.Context, fn func(context.Context, database.Store) error) error {
	m.inTransaction = true
	defer func() {
		m.inTransaction = false
	}()

	token := 1
	return fn(ctx, m.asStore(storeTestTx{token: &token}))
}

func (m *storeTestDB) Close() error {
	return nil
}

func (m *storeTestDB) ListServers(ctx context.Context, filter *database.ServerFilter, cursor string, limit int) ([]*apiv0.ServerResponse, string, error) {
	return m.asStore(nil).ListServers(ctx, filter, cursor, limit)
}

func (m *storeTestDB) ListAgents(ctx context.Context, filter *database.AgentFilter, cursor string, limit int) ([]*models.AgentResponse, string, error) {
	return m.asStore(nil).ListAgents(ctx, filter, cursor, limit)
}

func (m *storeTestDB) ListSkills(ctx context.Context, filter *database.SkillFilter, cursor string, limit int) ([]*models.SkillResponse, string, error) {
	return m.asStore(nil).ListSkills(ctx, filter, cursor, limit)
}

func (m *storeTestDB) ListPrompts(ctx context.Context, filter *database.PromptFilter, cursor string, limit int) ([]*models.PromptResponse, string, error) {
	return m.asStore(nil).ListPrompts(ctx, filter, cursor, limit)
}

func (m *storeTestDB) ListProviders(ctx context.Context, platform *string) ([]*models.Provider, error) {
	return m.asStore(nil).ListProviders(ctx, platform)
}

func (m *storeTestDB) GetDeployments(ctx context.Context, filter *models.DeploymentFilter) ([]*models.Deployment, error) {
	return m.asStore(nil).GetDeployments(ctx, filter)
}

func (m *storeTestDB) DeleteServer(ctx context.Context, serverName, version string) error {
	return m.asStore(nil).DeleteServer(ctx, serverName, version)
}

func (m *storeTestDB) DeleteAgent(ctx context.Context, agentName, version string) error {
	return m.asStore(nil).DeleteAgent(ctx, agentName, version)
}

func (s storeTestStore) ListServers(ctx context.Context, filter *database.ServerFilter, cursor string, limit int) ([]*apiv0.ServerResponse, string, error) {
	require.NotNil(s.db.testingT, s.db.testingT, "testingT must be set")
	require.NotNil(s.db.testingT, s.db.listServersFn, "listServersFn must be set")
	return s.db.listServersFn(ctx, s.tx, filter, cursor, limit)
}

func (s storeTestStore) ListAgents(ctx context.Context, filter *database.AgentFilter, cursor string, limit int) ([]*models.AgentResponse, string, error) {
	require.NotNil(s.db.testingT, s.db.testingT, "testingT must be set")
	require.NotNil(s.db.testingT, s.db.listAgentsFn, "listAgentsFn must be set")
	return s.db.listAgentsFn(ctx, s.tx, filter, cursor, limit)
}

func (s storeTestStore) ListSkills(ctx context.Context, filter *database.SkillFilter, cursor string, limit int) ([]*models.SkillResponse, string, error) {
	require.NotNil(s.db.testingT, s.db.testingT, "testingT must be set")
	require.NotNil(s.db.testingT, s.db.listSkillsFn, "listSkillsFn must be set")
	return s.db.listSkillsFn(ctx, s.tx, filter, cursor, limit)
}

func (s storeTestStore) ListPrompts(ctx context.Context, filter *database.PromptFilter, cursor string, limit int) ([]*models.PromptResponse, string, error) {
	require.NotNil(s.db.testingT, s.db.testingT, "testingT must be set")
	require.NotNil(s.db.testingT, s.db.listPromptsFn, "listPromptsFn must be set")
	return s.db.listPromptsFn(ctx, s.tx, filter, cursor, limit)
}

func (s storeTestStore) ListProviders(ctx context.Context, platform *string) ([]*models.Provider, error) {
	require.NotNil(s.db.testingT, s.db.testingT, "testingT must be set")
	require.NotNil(s.db.testingT, s.db.listProvidersFn, "listProvidersFn must be set")
	return s.db.listProvidersFn(ctx, s.tx, platform)
}

func (s storeTestStore) GetDeployments(ctx context.Context, filter *models.DeploymentFilter) ([]*models.Deployment, error) {
	require.NotNil(s.db.testingT, s.db.testingT, "testingT must be set")
	require.NotNil(s.db.testingT, s.db.getDeploymentsFn, "getDeploymentsFn must be set")
	return s.db.getDeploymentsFn(ctx, s.tx, filter)
}

func (s storeTestStore) DeleteServer(ctx context.Context, serverName, version string) error {
	require.NotNil(s.db.testingT, s.db.testingT, "testingT must be set")
	require.NotNil(s.db.testingT, s.db.deleteServerFn, "deleteServerFn must be set")
	return s.db.deleteServerFn(ctx, s.tx, serverName, version)
}

func (s storeTestStore) DeleteAgent(ctx context.Context, agentName, version string) error {
	require.NotNil(s.db.testingT, s.db.testingT, "testingT must be set")
	require.NotNil(s.db.testingT, s.db.deleteAgentFn, "deleteAgentFn must be set")
	return s.db.deleteAgentFn(ctx, s.tx, agentName, version)
}

func TestReadStoresUsesServiceDatabase(t *testing.T) {
	called := false
	mockDB := &storeTestDB{
		testingT: t,
		listServersFn: func(ctx context.Context, tx database.Transaction, filter *database.ServerFilter, cursor string, limit int) ([]*apiv0.ServerResponse, string, error) {
			require.Nil(t, tx)
			require.Nil(t, filter)
			require.Equal(t, "", cursor)
			require.Equal(t, 25, limit)
			called = true
			return nil, "next-cursor", nil
		},
	}

	svc := &registryServiceImpl{storeDB: mockDB}

	_, nextCursor, err := svc.readStores().servers.ListServers(context.Background(), nil, "", 25)
	require.NoError(t, err)
	require.True(t, called)
	require.Equal(t, "next-cursor", nextCursor)
}

func TestReadStoresUsesRepositoryOverrides(t *testing.T) {
	tests := []struct {
		name      string
		configure func(base, override *storeTestDB, baseCalled, overrideCalled *bool)
		setRepo   func(svc *registryServiceImpl, repo database.ServiceDatabase)
		invoke    func(t *testing.T, stores storeBundle)
	}{
		{
			name: "servers",
			configure: func(base, override *storeTestDB, baseCalled, overrideCalled *bool) {
				base.listServersFn = func(ctx context.Context, tx database.Transaction, filter *database.ServerFilter, cursor string, limit int) ([]*apiv0.ServerResponse, string, error) {
					*baseCalled = true
					return nil, "database", nil
				}
				override.listServersFn = func(ctx context.Context, tx database.Transaction, filter *database.ServerFilter, cursor string, limit int) ([]*apiv0.ServerResponse, string, error) {
					*overrideCalled = true
					return nil, "override", nil
				}
			},
			setRepo: func(svc *registryServiceImpl, repo database.ServiceDatabase) {
				svc.serverRepo = repo
			},
			invoke: func(t *testing.T, stores storeBundle) {
				_, nextCursor, err := stores.servers.ListServers(context.Background(), nil, "", 10)
				require.NoError(t, err)
				require.Equal(t, "override", nextCursor)
			},
		},
		{
			name: "agents",
			configure: func(base, override *storeTestDB, baseCalled, overrideCalled *bool) {
				base.listAgentsFn = func(ctx context.Context, tx database.Transaction, filter *database.AgentFilter, cursor string, limit int) ([]*models.AgentResponse, string, error) {
					*baseCalled = true
					return nil, "database", nil
				}
				override.listAgentsFn = func(ctx context.Context, tx database.Transaction, filter *database.AgentFilter, cursor string, limit int) ([]*models.AgentResponse, string, error) {
					*overrideCalled = true
					return nil, "override", nil
				}
			},
			setRepo: func(svc *registryServiceImpl, repo database.ServiceDatabase) {
				svc.agentRepo = repo
			},
			invoke: func(t *testing.T, stores storeBundle) {
				_, nextCursor, err := stores.agents.ListAgents(context.Background(), nil, "", 10)
				require.NoError(t, err)
				require.Equal(t, "override", nextCursor)
			},
		},
		{
			name: "skills",
			configure: func(base, override *storeTestDB, baseCalled, overrideCalled *bool) {
				base.listSkillsFn = func(ctx context.Context, tx database.Transaction, filter *database.SkillFilter, cursor string, limit int) ([]*models.SkillResponse, string, error) {
					*baseCalled = true
					return nil, "database", nil
				}
				override.listSkillsFn = func(ctx context.Context, tx database.Transaction, filter *database.SkillFilter, cursor string, limit int) ([]*models.SkillResponse, string, error) {
					*overrideCalled = true
					return nil, "override", nil
				}
			},
			setRepo: func(svc *registryServiceImpl, repo database.ServiceDatabase) {
				svc.skillRepo = repo
			},
			invoke: func(t *testing.T, stores storeBundle) {
				_, nextCursor, err := stores.skills.ListSkills(context.Background(), nil, "", 10)
				require.NoError(t, err)
				require.Equal(t, "override", nextCursor)
			},
		},
		{
			name: "prompts",
			configure: func(base, override *storeTestDB, baseCalled, overrideCalled *bool) {
				base.listPromptsFn = func(ctx context.Context, tx database.Transaction, filter *database.PromptFilter, cursor string, limit int) ([]*models.PromptResponse, string, error) {
					*baseCalled = true
					return nil, "database", nil
				}
				override.listPromptsFn = func(ctx context.Context, tx database.Transaction, filter *database.PromptFilter, cursor string, limit int) ([]*models.PromptResponse, string, error) {
					*overrideCalled = true
					return nil, "override", nil
				}
			},
			setRepo: func(svc *registryServiceImpl, repo database.ServiceDatabase) {
				svc.promptRepo = repo
			},
			invoke: func(t *testing.T, stores storeBundle) {
				_, nextCursor, err := stores.prompts.ListPrompts(context.Background(), nil, "", 10)
				require.NoError(t, err)
				require.Equal(t, "override", nextCursor)
			},
		},
		{
			name: "providers",
			configure: func(base, override *storeTestDB, baseCalled, overrideCalled *bool) {
				base.listProvidersFn = func(ctx context.Context, tx database.Transaction, platform *string) ([]*models.Provider, error) {
					*baseCalled = true
					return []*models.Provider{{ID: "database"}}, nil
				}
				override.listProvidersFn = func(ctx context.Context, tx database.Transaction, platform *string) ([]*models.Provider, error) {
					*overrideCalled = true
					return []*models.Provider{{ID: "override"}}, nil
				}
			},
			setRepo: func(svc *registryServiceImpl, repo database.ServiceDatabase) {
				svc.providerRepo = repo
			},
			invoke: func(t *testing.T, stores storeBundle) {
				providers, err := stores.providers.ListProviders(context.Background(), nil)
				require.NoError(t, err)
				require.Len(t, providers, 1)
				require.Equal(t, "override", providers[0].ID)
			},
		},
		{
			name: "deployments",
			configure: func(base, override *storeTestDB, baseCalled, overrideCalled *bool) {
				base.getDeploymentsFn = func(ctx context.Context, tx database.Transaction, filter *models.DeploymentFilter) ([]*models.Deployment, error) {
					*baseCalled = true
					return []*models.Deployment{{ID: "database"}}, nil
				}
				override.getDeploymentsFn = func(ctx context.Context, tx database.Transaction, filter *models.DeploymentFilter) ([]*models.Deployment, error) {
					*overrideCalled = true
					return []*models.Deployment{{ID: "override"}}, nil
				}
			},
			setRepo: func(svc *registryServiceImpl, repo database.ServiceDatabase) {
				svc.deploymentRepo = repo
			},
			invoke: func(t *testing.T, stores storeBundle) {
				deployments, err := stores.deployments.GetDeployments(context.Background(), nil)
				require.NoError(t, err)
				require.Len(t, deployments, 1)
				require.Equal(t, "override", deployments[0].ID)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			baseCalled := false
			overrideCalled := false

			base := &storeTestDB{testingT: t}
			override := &storeTestDB{testingT: t}
			tt.configure(base, override, &baseCalled, &overrideCalled)

			svc := &registryServiceImpl{storeDB: base}
			tt.setRepo(svc, override)

			tt.invoke(t, svc.readStores())
			require.True(t, overrideCalled)
			require.False(t, baseCalled)
		})
	}
}

func TestInTransactionUsesTransactionalStores(t *testing.T) {
	var mockDB *storeTestDB
	mockDB = &storeTestDB{
		testingT: t,
		deleteServerFn: func(ctx context.Context, tx database.Transaction, serverName, version string) error {
			require.True(t, mockDB.inTransaction)
			_, ok := tx.(storeTestTx)
			require.True(t, ok)
			require.Equal(t, "io.test/server", serverName)
			require.Equal(t, "1.0.0", version)
			return nil
		},
	}

	svc := &registryServiceImpl{storeDB: mockDB}

	err := svc.inTransaction(context.Background(), func(ctx context.Context, stores storeBundle) error {
		return stores.servers.DeleteServer(ctx, "io.test/server", "1.0.0")
	})
	require.NoError(t, err)
}

func TestInTransactionReusesTransactionAcrossStoreTypes(t *testing.T) {
	var serverToken *int
	var agentToken *int

	mockDB := &storeTestDB{
		testingT: t,
		deleteServerFn: func(ctx context.Context, tx database.Transaction, serverName, version string) error {
			typedTx, ok := tx.(storeTestTx)
			require.True(t, ok)
			serverToken = typedTx.token
			return nil
		},
		deleteAgentFn: func(ctx context.Context, tx database.Transaction, agentName, version string) error {
			typedTx, ok := tx.(storeTestTx)
			require.True(t, ok)
			agentToken = typedTx.token
			return nil
		},
	}

	svc := &registryServiceImpl{storeDB: mockDB}

	err := svc.inTransaction(context.Background(), func(ctx context.Context, stores storeBundle) error {
		if err := stores.servers.DeleteServer(ctx, "io.test/server", "1.0.0"); err != nil {
			return err
		}
		return stores.agents.DeleteAgent(ctx, "io.test/agent", "1.0.0")
	})
	require.NoError(t, err)
	require.NotNil(t, serverToken)
	require.Same(t, serverToken, agentToken)
}

func TestInTransactionRequiresServiceDatabase(t *testing.T) {
	svc := &registryServiceImpl{}

	err := svc.inTransaction(context.Background(), func(ctx context.Context, stores storeBundle) error {
		return nil
	})
	require.EqualError(t, err, "service database is not configured")
}
