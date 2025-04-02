package database

import (
	"context"
	"fmt"

	"github.com/specterops/bloodhound/src/model"
)

const (
	customNodeKindTable = "custom_node_kinds"
)

type CustomNodeKindData interface {
	CreateCustomNodeKind(ctx context.Context, customNodeKind model.CustomNodeKind) (model.CustomNodeKind, error)
	GetCustomNodeKinds(ctx context.Context) ([]model.CustomNodeKind, error)
	GetCustomNodeKind(ctx context.Context, id int32) (model.CustomNodeKind, error)
	UpdateCustomNodeKind(ctx context.Context, customNodeKind model.CustomNodeKind) (model.CustomNodeKind, error)
	DeleteCustomNodeKind(ctx context.Context, id int32) error
}

func (s *BloodhoundDB) CreateCustomNodeKind(ctx context.Context, customNodeKind model.CustomNodeKind) (model.CustomNodeKind, error) {
	result := s.db.WithContext(ctx).Exec(fmt.Sprintf("INSERT INTO %s (kind_id, config) VALUES (?, ?);", customNodeKindTable), customNodeKind.KindID, customNodeKind.Config)

	return customNodeKind, CheckError(result)
}

func (s *BloodhoundDB) GetCustomNodeKinds(ctx context.Context) ([]model.CustomNodeKind, error) {
	var customNodeKinds []model.CustomNodeKind
	result := s.db.WithContext(ctx).Raw(fmt.Sprintf("SELECT id, kind_id, config FROM %s;", customNodeKindTable)).Scan(&customNodeKinds)

	return customNodeKinds, CheckError(result)
}

func (s *BloodhoundDB) GetCustomNodeKind(ctx context.Context, id int32) (model.CustomNodeKind, error) {
	var customNodeKind model.CustomNodeKind
	result := s.db.WithContext(ctx).Raw(fmt.Sprintf("SELECT id, kind_id, config FROM %s WHERE id = ?;", customNodeKindTable), id).Scan(&customNodeKind)

	return customNodeKind, CheckError(result)
}

func (s *BloodhoundDB) UpdateCustomNodeKind(ctx context.Context, customNodeKind model.CustomNodeKind) (model.CustomNodeKind, error) {
	result := s.db.WithContext(ctx).Exec(fmt.Sprintf("UPDATE %s SET kind_id = ?, config = ? WHERE id = ?;", customNodeKindTable), customNodeKind.KindID, customNodeKind.Config, customNodeKind.ID)
	return customNodeKind, CheckError(result)
}

func (s *BloodhoundDB) DeleteCustomNodeKind(ctx context.Context, id int32) error {
	result := s.db.WithContext(ctx).Exec(fmt.Sprintf("DELETE FROM %s WHERE id = ?;", customNodeKindTable), id)

	return CheckError(result)
}
