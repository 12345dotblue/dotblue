package dbschema

import (
	"context"
	"fmt"
	"strings"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

type provider interface {
	Ensure(ctx context.Context, db gdb.DB) error
}

type statement struct {
	name string
	sql  string
}

func Ensure(ctx context.Context) error {
	db := g.DB()
	dialect := detectDialect(db)

	schemaProvider, err := providerForDialect(dialect)
	if err != nil {
		return err
	}
	return schemaProvider.Ensure(ctx, db)
}

func providerForDialect(dialect string) (provider, error) {
	switch dialect {
	case "pgsql", "postgres", "postgresql":
		return postgresProvider{}, nil
	default:
		return nil, fmt.Errorf("unsupported database dialect %q for schema initialization", dialect)
	}
}

func detectDialect(db gdb.DB) string {
	if db == nil || db.GetConfig() == nil {
		return ""
	}
	cfg := db.GetConfig()
	if cfg.Type != "" {
		return strings.ToLower(cfg.Type)
	}
	link := strings.ToLower(cfg.Link)
	switch {
	case strings.HasPrefix(link, "pgsql:"):
		return "pgsql"
	case strings.HasPrefix(link, "postgres:"):
		return "postgres"
	case strings.HasPrefix(link, "postgresql:"):
		return "postgresql"
	default:
		return ""
	}
}

func execStatements(ctx context.Context, db gdb.DB, statements []statement) error {
	for _, stmt := range statements {
		if _, err := db.Exec(ctx, stmt.sql); err != nil {
			return fmt.Errorf("%s: %w", stmt.name, err)
		}
	}
	return nil
}
