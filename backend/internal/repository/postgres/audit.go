package postgres

import (
	"context"
	"encoding/json"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/you/inkvault/internal/domain"
	"github.com/you/inkvault/internal/repository"
)

type auditRepo struct{ pool *pgxpool.Pool }

func (r *auditRepo) CreateLog(ctx context.Context, e *domain.AuditLog) error {
	meta, _ := json.Marshal(e.Meta)
	_, err := r.pool.Exec(ctx, `
		INSERT INTO audit_log
			(id,actor_id,actor_role,action,resource_type,resource_id,ip_address,user_agent,meta,created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`,
		e.ID, nilIfEmpty(e.ActorID), nilIfEmpty(string(e.ActorRole)),
		e.Action, nilIfEmpty(e.ResourceType), nilIfEmpty(e.ResourceID),
		nilIfEmpty(e.IPAddress), nilIfEmpty(e.UserAgent), meta, e.CreatedAt,
	)
	return err
}

func (r *auditRepo) ListLogs(ctx context.Context, f repository.AuditFilters) ([]*domain.AuditLog, error) {
	if f.Limit == 0 {
		f.Limit = 50
	}

	rows, err := r.pool.Query(ctx, `
		SELECT id,actor_id,actor_role,action,resource_type,resource_id,
		       ip_address,user_agent,meta,created_at
		FROM audit_log
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2`,
		f.Limit, f.Offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []*domain.AuditLog
	for rows.Next() {
		e := &domain.AuditLog{}
		var actorID, actorRole, resType, resID, ip, ua *string
		var metaBytes []byte

		if err := rows.Scan(
			&e.ID, &actorID, &actorRole, &e.Action,
			&resType, &resID, &ip, &ua, &metaBytes, &e.CreatedAt,
		); err != nil {
			return nil, err
		}

		if actorID != nil {
			e.ActorID = *actorID
		}
		if actorRole != nil {
			e.ActorRole = domain.Role(*actorRole)
		}
		if resType != nil {
			e.ResourceType = *resType
		}
		if resID != nil {
			e.ResourceID = *resID
		}
		if ip != nil {
			e.IPAddress = *ip
		}
		if ua != nil {
			e.UserAgent = *ua
		}
		if len(metaBytes) > 0 {
			_ = json.Unmarshal(metaBytes, &e.Meta)
		}
		logs = append(logs, e)
	}
	return logs, rows.Err()
}
