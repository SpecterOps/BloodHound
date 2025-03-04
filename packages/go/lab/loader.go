package lab

import (
	"context"
	"fmt"
	"io/fs"
	"path/filepath"

	"github.com/specterops/bloodhound/dawgs/graph"
	"gorm.io/gorm"
)

type FixtureFile struct {
	FS   fs.FS
	Path string
}

type FileSet map[fs.FS][]string

func LoadSqlFiles(db *gorm.DB, files ...FixtureFile) error {
	var data = make([]string, len(files))
	for idx, file := range files {
		if filepath.Ext(file.Path) != ".sql" {
			return fmt.Errorf("%s is not a .sql file", file.Path)
		}
		if b, err := fs.ReadFile(file.FS, file.Path); err != nil {
			return fmt.Errorf("error loading sql file `%s`: %w", file.Path, err)
		} else {
			data[idx] = string(b)
		}
	}
	return LoadSql(db, data...)
}

func LoadSql(db *gorm.DB, sql ...string) error {
	if err := db.Transaction(func(tx *gorm.DB) error {
		for _, statement := range sql {
			if res := tx.Raw(statement); res.Error != nil {
				return res.Error
			}
		}
		return nil
	}); err != nil {
		return fmt.Errorf("error loading sql data: %w", err)
	}
	return nil
}

func LoadCypherFiles(db graph.Database, files ...FixtureFile) error {
	var data = make([]string, len(files))
	for idx, file := range files {
		if filepath.Ext(file.Path) != ".cypher" {
			return fmt.Errorf("%s is not a .cypher file", file.Path)
		}
		if b, err := fs.ReadFile(file.FS, file.Path); err != nil {
			return fmt.Errorf("error loading cypher file `%s`: %w", file.Path, err)
		} else {
			data[idx] = string(b)
		}
	}
	return LoadCypher(db, data...)
}

func LoadCypher(db graph.Database, cypher ...string) error {
	if err := db.WriteTransaction(context.Background(), func(tx graph.Transaction) error {
		for _, statement := range cypher {
			if res := tx.Raw(statement, nil); res.Error() != nil {
				return res.Error()
			}
		}
		return nil
	}); err != nil {
		return fmt.Errorf("error loading cypher data: %w", err)
	}
	return nil
}
