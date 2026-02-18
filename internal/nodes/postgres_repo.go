package nodes

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PgRepository implements Repository using pgx.
type PgRepository struct {
	pool *pgxpool.Pool
}

// NewPgRepository creates a new PostgreSQL-backed repository for MCP servers.
func NewPgRepository(pool *pgxpool.Pool) *PgRepository {
	return &PgRepository{pool: pool}
}

func (r *PgRepository) ListServers(ctx context.Context, ownerID uuid.UUID) ([]MCPServer, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, owner_id, name, image, status, config, container_id, created_at, updated_at
		 FROM mcp_servers WHERE owner_id = $1
		 ORDER BY created_at DESC`,
		ownerID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var servers []MCPServer
	for rows.Next() {
		var s MCPServer
		var configBytes []byte
		if err := rows.Scan(&s.ID, &s.OwnerID, &s.Name, &s.Image, &s.Status, &configBytes, &s.ContainerID, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, err
		}
		if configBytes != nil {
			if err := json.Unmarshal(configBytes, &s.Config); err != nil {
				return nil, err
			}
		}
		servers = append(servers, s)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Return empty slice instead of nil for consistent JSON marshaling.
	if servers == nil {
		servers = []MCPServer{}
	}
	return servers, nil
}

func (r *PgRepository) GetServer(ctx context.Context, id uuid.UUID) (*MCPServer, error) {
	var s MCPServer
	var configBytes []byte
	err := r.pool.QueryRow(ctx,
		`SELECT id, owner_id, name, image, status, config, container_id, created_at, updated_at
		 FROM mcp_servers WHERE id = $1`,
		id,
	).Scan(&s.ID, &s.OwnerID, &s.Name, &s.Image, &s.Status, &configBytes, &s.ContainerID, &s.CreatedAt, &s.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	if configBytes != nil {
		if err := json.Unmarshal(configBytes, &s.Config); err != nil {
			return nil, err
		}
	}
	return &s, nil
}

func (r *PgRepository) CreateServer(ctx context.Context, server *MCPServer) error {
	configBytes, err := json.Marshal(server.Config)
	if err != nil {
		return err
	}

	_, err = r.pool.Exec(ctx,
		`INSERT INTO mcp_servers (id, owner_id, name, image, status, config, container_id, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		server.ID, server.OwnerID, server.Name, server.Image, server.Status, configBytes, server.ContainerID, server.CreatedAt, server.UpdatedAt,
	)
	return err
}

func (r *PgRepository) UpdateServer(ctx context.Context, server *MCPServer) error {
	configBytes, err := json.Marshal(server.Config)
	if err != nil {
		return err
	}

	tag, err := r.pool.Exec(ctx,
		`UPDATE mcp_servers SET name = $2, image = $3, status = $4, config = $5, container_id = $6, updated_at = $7
		 WHERE id = $1`,
		server.ID, server.Name, server.Image, server.Status, configBytes, server.ContainerID, server.UpdatedAt,
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *PgRepository) DeleteServer(ctx context.Context, id uuid.UUID) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM mcp_servers WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *PgRepository) SearchServers(ctx context.Context, query string) ([]MCPServer, error) {
	var rows pgx.Rows
	var err error

	if query == "" {
		rows, err = r.pool.Query(ctx,
			`SELECT id, owner_id, name, image, status, config, container_id, created_at, updated_at
			 FROM mcp_servers ORDER BY created_at DESC`)
	} else {
		pattern := "%" + query + "%"
		rows, err = r.pool.Query(ctx,
			`SELECT id, owner_id, name, image, status, config, container_id, created_at, updated_at
			 FROM mcp_servers WHERE name ILIKE $1 OR image ILIKE $1
			 ORDER BY created_at DESC`, pattern)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var servers []MCPServer
	for rows.Next() {
		var s MCPServer
		var configBytes []byte
		if err := rows.Scan(&s.ID, &s.OwnerID, &s.Name, &s.Image, &s.Status, &configBytes, &s.ContainerID, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, err
		}
		if configBytes != nil {
			if err := json.Unmarshal(configBytes, &s.Config); err != nil {
				return nil, err
			}
		}
		servers = append(servers, s)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	if servers == nil {
		servers = []MCPServer{}
	}
	return servers, nil
}

func (r *PgRepository) CreateOAuthGrant(ctx context.Context, grant *OAuthGrant) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO oauth_grants (id, server_id, provider, access_token_enc, refresh_token_enc, expires_at, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		grant.ID, grant.ServerID, grant.Provider, grant.AccessTokenEnc, grant.RefreshTokenEnc, grant.ExpiresAt, grant.CreatedAt,
	)
	return err
}

func (r *PgRepository) GetOAuthGrant(ctx context.Context, id uuid.UUID) (*OAuthGrant, error) {
	var g OAuthGrant
	err := r.pool.QueryRow(ctx,
		`SELECT id, server_id, provider, access_token_enc, refresh_token_enc, expires_at, created_at
		 FROM oauth_grants WHERE id = $1`,
		id,
	).Scan(&g.ID, &g.ServerID, &g.Provider, &g.AccessTokenEnc, &g.RefreshTokenEnc, &g.ExpiresAt, &g.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &g, nil
}
