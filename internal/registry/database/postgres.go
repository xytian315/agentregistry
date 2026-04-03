package database

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/agentregistry-dev/agentregistry/pkg/models"
	"github.com/agentregistry-dev/agentregistry/pkg/registry/auth"
	"github.com/agentregistry-dev/agentregistry/pkg/registry/database"
)

// PostgreSQL is an implementation of the Database interface using PostgreSQL
type PostgreSQL struct {
	pool  *pgxpool.Pool
	authz auth.Authorizer
}

type commandTagAdapter struct {
	tag pgconn.CommandTag
}

func (c commandTagAdapter) RowsAffected() int64 {
	return c.tag.RowsAffected()
}

type rowsAdapter struct {
	rows pgx.Rows
}

func (r rowsAdapter) Close() {
	r.rows.Close()
}

func (r rowsAdapter) Err() error {
	return r.rows.Err()
}

func (r rowsAdapter) Next() bool {
	return r.rows.Next()
}

func (r rowsAdapter) Scan(dest ...any) error {
	return r.rows.Scan(dest...)
}

type rowAdapter struct {
	row pgx.Row
}

func (r rowAdapter) Scan(dest ...any) error {
	return r.row.Scan(dest...)
}

type transactionAdapter struct {
	tx pgx.Tx
}

func (t transactionAdapter) Exec(ctx context.Context, sql string, arguments ...any) (database.CommandTag, error) {
	result, err := t.tx.Exec(ctx, sql, arguments...)
	if err != nil {
		return nil, err
	}
	return commandTagAdapter{tag: result}, nil
}

func (t transactionAdapter) Query(ctx context.Context, sql string, args ...any) (database.Rows, error) {
	rows, err := t.tx.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	return rowsAdapter{rows: rows}, nil
}

func (t transactionAdapter) QueryRow(ctx context.Context, sql string, args ...any) database.Row {
	return rowAdapter{row: t.tx.QueryRow(ctx, sql, args...)}
}

type poolExecutor struct {
	pool *pgxpool.Pool
}

func (p poolExecutor) Exec(ctx context.Context, sql string, arguments ...any) (database.CommandTag, error) {
	result, err := p.pool.Exec(ctx, sql, arguments...)
	if err != nil {
		return nil, err
	}
	return commandTagAdapter{tag: result}, nil
}

func (p poolExecutor) Query(ctx context.Context, sql string, args ...any) (database.Rows, error) {
	rows, err := p.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	return rowsAdapter{rows: rows}, nil
}

func (p poolExecutor) QueryRow(ctx context.Context, sql string, args ...any) database.Row {
	return rowAdapter{row: p.pool.QueryRow(ctx, sql, args...)}
}

// Executor is an internal query surface used by repository methods.
type Executor interface {
	Exec(ctx context.Context, sql string, arguments ...any) (database.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (database.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) database.Row
}

// getExecutor returns the appropriate executor (transaction or pool)
func (db *PostgreSQL) getExecutor(tx database.Transaction) Executor {
	if tx != nil {
		return tx
	}
	return poolExecutor{pool: db.pool}
}

// NewPostgreSQL creates a new instance of the PostgreSQL database
func NewPostgreSQL(ctx context.Context, connectionURI string, authz auth.Authorizer, vectorEnabled bool) (*PostgreSQL, error) {
	// Parse connection config for pool settings
	config, err := pgxpool.ParseConfig(connectionURI)
	if err != nil {
		return nil, fmt.Errorf("failed to parse PostgreSQL config: %w", err)
	}

	// Configure pool for stability-focused defaults
	config.MaxConns = 30                      // Handle good concurrent load
	config.MinConns = 5                       // Keep connections warm for fast response
	config.MaxConnIdleTime = 30 * time.Minute // Keep connections available for bursts
	config.MaxConnLifetime = 2 * time.Hour    // Refresh connections regularly for stability

	// Create connection pool with configured settings
	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create PostgreSQL pool: %w", err)
	}

	// Test the connection
	if err = pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping PostgreSQL: %w", err)
	}

	// Run migrations using a single connection from the pool
	conn, err := pool.Acquire(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to acquire connection for migrations: %w", err)
	}
	defer conn.Release()

	migrator := database.NewMigrator(conn.Conn(), DefaultMigratorConfig())
	if err := migrator.Migrate(ctx); err != nil {
		return nil, fmt.Errorf("failed to run database migrations: %w", err)
	}

	if vectorEnabled {
		vectorMigrator := database.NewMigrator(conn.Conn(), VectorMigratorConfig())
		if err := vectorMigrator.Migrate(ctx); err != nil {
			return nil, fmt.Errorf("failed to run vector database migrations: %w", err)
		}
	}

	return &PostgreSQL{
		pool:  pool,
		authz: authz,
	}, nil
}

// InTransaction executes a function within a database transaction
func (db *PostgreSQL) InTransaction(ctx context.Context, fn func(ctx context.Context, tx database.Transaction) error) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}

	tx, err := db.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	//nolint:contextcheck // Intentionally using separate context for rollback to ensure cleanup even if request is cancelled
	defer func() {
		rollbackCtx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()
		if rbErr := tx.Rollback(rollbackCtx); rbErr != nil && !errors.Is(rbErr, pgx.ErrTxClosed) {
			slog.Error("failed to rollback transaction", "error", rbErr)
		}
	}()

	if err := fn(ctx, transactionAdapter{tx: tx}); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// CreateProvider creates a provider record.
func (db *PostgreSQL) CreateProvider(ctx context.Context, tx database.Transaction, in *models.CreateProviderInput) (*models.Provider, error) {
	if in == nil {
		return nil, database.ErrInvalidInput
	}
	if strings.TrimSpace(in.ID) == "" || strings.TrimSpace(in.Name) == "" || strings.TrimSpace(in.Platform) == "" {
		return nil, database.ErrInvalidInput
	}
	executor := db.getExecutor(tx)
	configJSON, err := json.Marshal(in.Config)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal provider config: %w", err)
	}
	query := `
		INSERT INTO providers (id, name, platform, config)
		VALUES ($1, $2, $3, $4)
		RETURNING id, name, platform, COALESCE(config, '{}'::jsonb), created_at, updated_at
	`
	var provider models.Provider
	var configOut []byte
	err = executor.QueryRow(ctx, query, in.ID, in.Name, in.Platform, configJSON).Scan(
		&provider.ID,
		&provider.Name,
		&provider.Platform,
		&configOut,
		&provider.CreatedAt,
		&provider.UpdatedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, database.ErrAlreadyExists
		}
		return nil, fmt.Errorf("failed to create provider: %w", err)
	}
	if len(configOut) > 0 {
		if err := json.Unmarshal(configOut, &provider.Config); err != nil {
			return nil, fmt.Errorf("failed to unmarshal provider config: %w", err)
		}
	}
	if provider.Config == nil {
		provider.Config = map[string]any{}
	}
	return &provider, nil
}

// ListProviders lists providers, optionally filtered by platform.
func (db *PostgreSQL) ListProviders(ctx context.Context, tx database.Transaction, platform *string) ([]*models.Provider, error) {
	executor := db.getExecutor(tx)
	query := `SELECT id, name, platform, COALESCE(config, '{}'::jsonb), created_at, updated_at FROM providers`
	args := []any{}
	if platform != nil && strings.TrimSpace(*platform) != "" {
		query += ` WHERE platform = $1`
		args = append(args, strings.TrimSpace(*platform))
	}
	query += ` ORDER BY created_at ASC`
	rows, err := executor.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list providers: %w", err)
	}
	defer rows.Close()
	var out []*models.Provider
	for rows.Next() {
		var p models.Provider
		var configJSON []byte
		if err := rows.Scan(&p.ID, &p.Name, &p.Platform, &configJSON, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan provider: %w", err)
		}
		if len(configJSON) > 0 {
			if err := json.Unmarshal(configJSON, &p.Config); err != nil {
				return nil, fmt.Errorf("failed to unmarshal provider config: %w", err)
			}
		}
		if p.Config == nil {
			p.Config = map[string]any{}
		}
		out = append(out, &p)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate providers: %w", err)
	}
	return out, nil
}

// GetProviderByID gets a provider by ID.
func (db *PostgreSQL) GetProviderByID(ctx context.Context, tx database.Transaction, providerID string) (*models.Provider, error) {
	executor := db.getExecutor(tx)
	query := `SELECT id, name, platform, COALESCE(config, '{}'::jsonb), created_at, updated_at FROM providers WHERE id = $1`
	var p models.Provider
	var configJSON []byte
	if err := executor.QueryRow(ctx, query, providerID).Scan(&p.ID, &p.Name, &p.Platform, &configJSON, &p.CreatedAt, &p.UpdatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, database.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get provider: %w", err)
	}
	if len(configJSON) > 0 {
		if err := json.Unmarshal(configJSON, &p.Config); err != nil {
			return nil, fmt.Errorf("failed to unmarshal provider config: %w", err)
		}
	}
	if p.Config == nil {
		p.Config = map[string]any{}
	}
	return &p, nil
}

// UpdateProvider updates mutable provider fields.
func (db *PostgreSQL) UpdateProvider(ctx context.Context, tx database.Transaction, providerID string, in *models.UpdateProviderInput) (*models.Provider, error) {
	if in == nil {
		return db.GetProviderByID(ctx, tx, providerID)
	}
	current, err := db.GetProviderByID(ctx, tx, providerID)
	if err != nil {
		return nil, err
	}
	name := current.Name
	if in.Name != nil {
		name = *in.Name
	}
	config := current.Config
	if in.Config != nil {
		config = in.Config
	}
	configJSON, err := json.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal provider config: %w", err)
	}
	executor := db.getExecutor(tx)
	query := `
		UPDATE providers
		SET name = $2, config = $3, updated_at = NOW()
		WHERE id = $1
		RETURNING id, name, platform, COALESCE(config, '{}'::jsonb), created_at, updated_at
	`
	var p models.Provider
	var configOut []byte
	if err := executor.QueryRow(ctx, query, providerID, name, configJSON).Scan(&p.ID, &p.Name, &p.Platform, &configOut, &p.CreatedAt, &p.UpdatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, database.ErrNotFound
		}
		return nil, fmt.Errorf("failed to update provider: %w", err)
	}
	if len(configOut) > 0 {
		if err := json.Unmarshal(configOut, &p.Config); err != nil {
			return nil, fmt.Errorf("failed to unmarshal provider config: %w", err)
		}
	}
	if p.Config == nil {
		p.Config = map[string]any{}
	}
	return &p, nil
}

// DeleteProvider removes a provider by ID.
func (db *PostgreSQL) DeleteProvider(ctx context.Context, tx database.Transaction, providerID string) error {
	executor := db.getExecutor(tx)
	result, err := executor.Exec(ctx, `DELETE FROM providers WHERE id = $1`, providerID)
	if err != nil {
		return fmt.Errorf("failed to delete provider: %w", err)
	}
	if result.RowsAffected() == 0 {
		return database.ErrNotFound
	}
	return nil
}

// CreateDeployment creates a new deployment record
func (db *PostgreSQL) CreateDeployment(ctx context.Context, tx database.Transaction, deployment *models.Deployment) error {
	// Authz check (determine resource type)
	artifactType := auth.PermissionArtifactTypeServer
	if deployment.ResourceType == "agent" {
		artifactType = auth.PermissionArtifactTypeAgent
	}
	if err := db.authz.Check(ctx, auth.PermissionActionDeploy, auth.Resource{
		Name: deployment.ServerName,
		Type: artifactType,
	}); err != nil {
		return err
	}

	executor := db.getExecutor(tx)

	envJSON, err := json.Marshal(deployment.Env)
	if err != nil {
		return fmt.Errorf("failed to marshal deployment env: %w", err)
	}

	providerConfigJSON, err := json.Marshal(deployment.ProviderConfig)
	if err != nil {
		return fmt.Errorf("failed to marshal provider config: %w", err)
	}
	providerMetadataJSON, err := json.Marshal(deployment.ProviderMetadata)
	if err != nil {
		return fmt.Errorf("failed to marshal provider metadata: %w", err)
	}

	// Default to 'mcp' if not specified
	resourceType := deployment.ResourceType
	if resourceType == "" {
		resourceType = "mcp"
	}
	providerID := strings.TrimSpace(deployment.ProviderID)
	if providerID == "" {
		return fmt.Errorf("%w: provider id is required", database.ErrInvalidInput)
	}
	origin := deployment.Origin
	if origin == "" {
		origin = "managed"
	}
	deployment.ProviderID = providerID

	if deployment.ID == "" {
		_ = db.getExecutor(tx).QueryRow(ctx, "SELECT uuid_generate_v4()::text").Scan(&deployment.ID)
	}

	query := `
		INSERT INTO deployments (
			id, server_name, version, status, config, prefer_remote, resource_type,
			origin, provider_id, provider_config, provider_metadata, error
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NULLIF($9, ''), $10, $11, $12)
	`

	_, err = executor.Exec(ctx, query,
		deployment.ID,
		deployment.ServerName,
		deployment.Version,
		deployment.Status,
		envJSON,
		deployment.PreferRemote,
		resourceType,
		origin,
		providerID,
		providerConfigJSON,
		providerMetadataJSON,
		deployment.Error,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" { // unique_violation
			return database.ErrAlreadyExists
		}
		return fmt.Errorf("failed to create deployment: %w", err)
	}

	return nil
}

// GetDeployments retrieves all deployed servers
func (db *PostgreSQL) GetDeployments(ctx context.Context, tx database.Transaction, filter *models.DeploymentFilter) ([]*models.Deployment, error) {
	executor := db.getExecutor(tx)

	where, args, needsProviderJoin := buildDeploymentFilters(filter)

	query := `SELECT
			d.id, d.server_name, d.version, d.deployed_at, d.updated_at, d.status, d.config, d.prefer_remote, d.resource_type,
			d.origin, COALESCE(d.provider_id, ''), COALESCE(d.provider_config, '{}'::jsonb), COALESCE(d.provider_metadata, '{}'::jsonb), COALESCE(d.error, '')
		FROM deployments d`
	if needsProviderJoin {
		query += ` LEFT JOIN providers p ON p.id = d.provider_id`
	}
	if len(where) > 0 {
		query += " WHERE " + strings.Join(where, " AND ")
	}
	query += " ORDER BY d.deployed_at DESC"

	rows, err := executor.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query deployments: %w", err)
	}
	defer rows.Close()

	var deployments []*models.Deployment
	for rows.Next() {
		var d models.Deployment
		var envJSON []byte
		var providerConfigJSON []byte
		var providerMetadataJSON []byte

		err := rows.Scan(
			&d.ID,
			&d.ServerName,
			&d.Version,
			&d.DeployedAt,
			&d.UpdatedAt,
			&d.Status,
			&envJSON,
			&d.PreferRemote,
			&d.ResourceType,
			&d.Origin,
			&d.ProviderID,
			&providerConfigJSON,
			&providerMetadataJSON,
			&d.Error,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan deployment: %w", err)
		}

		if len(envJSON) > 0 {
			if err := json.Unmarshal(envJSON, &d.Env); err != nil {
				return nil, fmt.Errorf("failed to unmarshal deployment env: %w", err)
			}
		}
		if d.Env == nil {
			d.Env = make(map[string]string)
		}
		if err := json.Unmarshal(providerConfigJSON, &d.ProviderConfig); err != nil {
			return nil, fmt.Errorf("failed to scan provider config: %w", err)
		}
		if err := json.Unmarshal(providerMetadataJSON, &d.ProviderMetadata); err != nil {
			return nil, fmt.Errorf("failed to scan provider metadata: %w", err)
		}

		deployments = append(deployments, &d)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating deployments: %w", err)
	}

	return deployments, nil
}

func buildDeploymentFilters(filter *models.DeploymentFilter) ([]string, []any, bool) {
	where := make([]string, 0)
	args := make([]any, 0)
	nextArg := 1
	needsProviderJoin := false
	if filter == nil {
		return where, args, needsProviderJoin
	}

	if filter.Platform != nil {
		platform := strings.ToLower(strings.TrimSpace(*filter.Platform))
		needsProviderJoin = true
		where = append(where, fmt.Sprintf("p.platform = $%d", nextArg))
		args = append(args, platform)
		nextArg++
	}
	if filter.ResourceType != nil {
		where = append(where, fmt.Sprintf("resource_type = $%d", nextArg))
		args = append(args, *filter.ResourceType)
		nextArg++
	}
	if filter.Status != nil {
		where = append(where, fmt.Sprintf("status = $%d", nextArg))
		args = append(args, *filter.Status)
		nextArg++
	}
	if filter.Origin != nil {
		where = append(where, fmt.Sprintf("origin = $%d", nextArg))
		args = append(args, *filter.Origin)
		nextArg++
	}
	if filter.ResourceName != nil {
		where = append(where, fmt.Sprintf("server_name ILIKE $%d", nextArg))
		args = append(args, "%"+*filter.ResourceName+"%")
		nextArg++
	}
	if filter.ProviderID != nil {
		where = append(where, fmt.Sprintf("d.provider_id = $%d", nextArg))
		args = append(args, *filter.ProviderID)
	}

	return where, args, needsProviderJoin
}

// GetDeploymentByID retrieves a specific deployment by UUID.
func (db *PostgreSQL) GetDeploymentByID(ctx context.Context, tx database.Transaction, id string) (*models.Deployment, error) {
	executor := db.getExecutor(tx)
	query := `SELECT
			id, server_name, version, deployed_at, updated_at, status, config, prefer_remote, resource_type,
			origin, COALESCE(provider_id, ''), COALESCE(provider_config, '{}'::jsonb), COALESCE(provider_metadata, '{}'::jsonb), COALESCE(error, '')
		FROM deployments
		WHERE id = $1`

	var d models.Deployment
	var envJSON []byte
	var providerConfigJSON []byte
	var providerMetadataJSON []byte
	err := executor.QueryRow(ctx, query, id).Scan(
		&d.ID,
		&d.ServerName,
		&d.Version,
		&d.DeployedAt,
		&d.UpdatedAt,
		&d.Status,
		&envJSON,
		&d.PreferRemote,
		&d.ResourceType,
		&d.Origin,
		&d.ProviderID,
		&providerConfigJSON,
		&providerMetadataJSON,
		&d.Error,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, database.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get deployment by id: %w", err)
	}
	if len(envJSON) > 0 {
		if err := json.Unmarshal(envJSON, &d.Env); err != nil {
			return nil, fmt.Errorf("failed to unmarshal deployment env: %w", err)
		}
	}
	if d.Env == nil {
		d.Env = make(map[string]string)
	}
	if err := json.Unmarshal(providerConfigJSON, &d.ProviderConfig); err != nil {
		return nil, fmt.Errorf("failed to scan provider config: %w", err)
	}
	if err := json.Unmarshal(providerMetadataJSON, &d.ProviderMetadata); err != nil {
		return nil, fmt.Errorf("failed to scan provider metadata: %w", err)
	}
	artifactType := auth.PermissionArtifactTypeServer
	if d.ResourceType == "agent" {
		artifactType = auth.PermissionArtifactTypeAgent
	}
	if err := db.authz.Check(ctx, auth.PermissionActionRead, auth.Resource{
		Name: d.ServerName,
		Type: artifactType,
	}); err != nil {
		return nil, err
	}
	return &d, nil
}

// UpdateDeploymentState applies partial state updates to a deployment by ID.
func (db *PostgreSQL) UpdateDeploymentState(ctx context.Context, tx database.Transaction, id string, patch *models.DeploymentStatePatch) error {
	if patch == nil {
		return fmt.Errorf("%w: deployment state patch is required", database.ErrInvalidInput)
	}

	deployment, err := db.GetDeploymentByID(ctx, tx, id)
	if err != nil {
		return err
	}
	artifactType := auth.PermissionArtifactTypeServer
	if deployment.ResourceType == "agent" {
		artifactType = auth.PermissionArtifactTypeAgent
	}
	if err := db.authz.Check(ctx, auth.PermissionActionEdit, auth.Resource{
		Name: deployment.ServerName,
		Type: artifactType,
	}); err != nil {
		return err
	}

	executor := db.getExecutor(tx)
	setStatus := patch.Status != nil
	statusValue := deployment.Status
	if patch.Status != nil {
		statusValue = *patch.Status
	}

	setError := patch.Error != nil
	errorValue := deployment.Error
	if patch.Error != nil {
		errorValue = *patch.Error
	}

	setProviderConfig := patch.ProviderConfig != nil
	providerConfigJSON := []byte("{}")
	if patch.ProviderConfig != nil {
		providerConfigJSON, err = json.Marshal(*patch.ProviderConfig)
		if err != nil {
			return fmt.Errorf("failed to marshal provider config patch: %w", err)
		}
	}

	setProviderMetadata := patch.ProviderMetadata != nil
	providerMetadataJSON := []byte("{}")
	if patch.ProviderMetadata != nil {
		providerMetadataJSON, err = json.Marshal(*patch.ProviderMetadata)
		if err != nil {
			return fmt.Errorf("failed to marshal provider metadata patch: %w", err)
		}
	}

	query := `
		UPDATE deployments
		SET
			status = CASE WHEN $2 THEN $3 ELSE status END,
			error = CASE WHEN $4 THEN $5 ELSE error END,
			provider_config = CASE WHEN $6 THEN $7::jsonb ELSE provider_config END,
			provider_metadata = CASE WHEN $8 THEN $9::jsonb ELSE provider_metadata END,
			updated_at = NOW()
		WHERE id = $1
	`

	result, err := executor.Exec(
		ctx,
		query,
		id,
		setStatus,
		statusValue,
		setError,
		errorValue,
		setProviderConfig,
		string(providerConfigJSON),
		setProviderMetadata,
		string(providerMetadataJSON),
	)
	if err != nil {
		return fmt.Errorf("failed to update deployment state: %w", err)
	}

	if result.RowsAffected() == 0 {
		return database.ErrNotFound
	}

	return nil
}

// RemoveDeploymentByID removes a deployment by UUID.
func (db *PostgreSQL) RemoveDeploymentByID(ctx context.Context, tx database.Transaction, id string) error {
	deployment, err := db.GetDeploymentByID(ctx, tx, id)
	if err != nil {
		return err
	}
	artifactType := auth.PermissionArtifactTypeServer
	if deployment.ResourceType == "agent" {
		artifactType = auth.PermissionArtifactTypeAgent
	}
	if err := db.authz.Check(ctx, auth.PermissionActionDeploy, auth.Resource{
		Name: deployment.ServerName,
		Type: artifactType,
	}); err != nil {
		return err
	}

	executor := db.getExecutor(tx)
	query := `DELETE FROM deployments WHERE id = $1`

	result, err := executor.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete deployment by id: %w", err)
	}
	if result.RowsAffected() == 0 {
		return database.ErrNotFound
	}
	return nil
}

// Close closes the database connection
func (db *PostgreSQL) Close() error {
	db.pool.Close()
	return nil
}
