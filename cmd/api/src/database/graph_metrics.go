package database

import (
	"context"
)

type GraphMetrics interface {
	GetEdgeCountByKind(ctx context.Context, kind string) (int64, error)
}

func (s *BloodhoundDB) GetEdgeCountByKind(ctx context.Context, kind string) (int64, error) {
	var edgeCount int64

	if result := s.db.WithContext(ctx).Raw(
		`SELECT COUNT(*) FROM edge e
		JOIN kind k ON e.kind_id = k.id
		WHERE k.name = ?`,
		kind,
	).Scan(&edgeCount); result.Error != nil {
		return 0, CheckError(result)
	}

	return int64(edgeCount), nil
}
