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
	CreateCustomNodeKinds(ctx context.Context, customNodeKind []model.CustomNodeKind) ([]model.CustomNodeKind, error)
	GetCustomNodeKinds(ctx context.Context) ([]model.CustomNodeKind, error)
	GetCustomNodeKind(ctx context.Context, kindName string) (model.CustomNodeKind, error)
	UpdateCustomNodeKind(ctx context.Context, customNodeKind model.CustomNodeKind) (model.CustomNodeKind, error)
	DeleteCustomNodeKind(ctx context.Context, kindName string) error
}

func (s *BloodhoundDB) CreateCustomNodeKinds(ctx context.Context, customNodeKinds []model.CustomNodeKind) ([]model.CustomNodeKind, error) {
	result := s.db.WithContext(ctx).Create(&customNodeKinds)

	return customNodeKinds, CheckError(result)
}

func (s *BloodhoundDB) GetCustomNodeKinds(ctx context.Context) ([]model.CustomNodeKind, error) {
	var customNodeKinds []model.CustomNodeKind
	result := s.db.WithContext(ctx).Raw(fmt.Sprintf("SELECT id, kind_name, config FROM %s;", customNodeKindTable)).Scan(&customNodeKinds)

	return customNodeKinds, CheckError(result)
}

func (s *BloodhoundDB) GetCustomNodeKind(ctx context.Context, kindName string) (model.CustomNodeKind, error) {
	var customNodeKind model.CustomNodeKind
	result := s.db.WithContext(ctx).Raw(fmt.Sprintf("SELECT id, kind_name, config FROM %s WHERE kind_name = ?;", customNodeKindTable), kindName).Scan(&customNodeKind)
	if result.RowsAffected == 0 {
		return customNodeKind, ErrNotFound
	}

	return customNodeKind, CheckError(result)
}

func (s *BloodhoundDB) UpdateCustomNodeKind(ctx context.Context, customNodeKind model.CustomNodeKind) (model.CustomNodeKind, error) {
	result := s.db.WithContext(ctx).Raw(fmt.Sprintf("UPDATE %s SET config = ? WHERE kind_name = ? RETURNING id;", customNodeKindTable), customNodeKind.Config, customNodeKind.KindName).Scan(&customNodeKind.ID)
	if result.RowsAffected == 0 {
		return customNodeKind, ErrNotFound
	}

	return customNodeKind, CheckError(result)
}

func (s *BloodhoundDB) DeleteCustomNodeKind(ctx context.Context, kindName string) error {
	result := s.db.WithContext(ctx).Exec(fmt.Sprintf("DELETE FROM %s WHERE kind_name = ?;", customNodeKindTable), kindName)
	if result.RowsAffected == 0 {
		return ErrNotFound
	}

	return CheckError(result)
}
