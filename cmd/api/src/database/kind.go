package database

import (
	"context"
	"fmt"

	"github.com/specterops/bloodhound/cmd/api/src/model"
)

// GetKindById gets a row from the kind table by id.
func (s *BloodhoundDB) GetKindById(ctx context.Context, id int32) (model.Kind, error) {
	var kind model.Kind

	// var kind Kind
	return kind, CheckError(s.db.WithContext(ctx).Raw(fmt.Sprintf(`
		SELECT id, name
		FROM %s WHERE id = ?`, kindTable), id).First(&kind),
	)
}
