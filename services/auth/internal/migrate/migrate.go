package migrate

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Run applies *.up.sql migrations from dir in filename order, tracking
// applied versions in <schema>._migrations. Idempotent.
func Run(ctx context.Context, pool *pgxpool.Pool, dir, schema string) error {
	if _, err := pool.Exec(ctx, fmt.Sprintf(`CREATE SCHEMA IF NOT EXISTS %s`, schema)); err != nil {
		return fmt.Errorf("create schema: %w", err)
	}
	if _, err := pool.Exec(ctx, fmt.Sprintf(`SET search_path TO %s, public`, schema)); err != nil {
		return fmt.Errorf("set search_path: %w", err)
	}

	if _, err := pool.Exec(ctx, fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s._migrations (
			version TEXT PRIMARY KEY,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`, schema)); err != nil {
		return fmt.Errorf("create _migrations: %w", err)
	}

	applied, err := loadApplied(ctx, pool, schema)
	if err != nil {
		return err
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("read migrations dir %q: %w", dir, err)
	}

	var files []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".up.sql") {
			files = append(files, e.Name())
		}
	}
	sort.Strings(files)

	for _, name := range files {
		version := strings.TrimSuffix(name, ".up.sql")
		if applied[version] {
			continue
		}

		sqlBytes, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			return fmt.Errorf("read %s: %w", name, err)
		}

		tx, err := pool.Begin(ctx)
		if err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, string(sqlBytes)); err != nil {
			_ = tx.Rollback(ctx)
			return fmt.Errorf("apply %s: %w", name, err)
		}
		if _, err := tx.Exec(ctx,
			fmt.Sprintf(`INSERT INTO %s._migrations (version) VALUES ($1)`, schema),
			version,
		); err != nil {
			_ = tx.Rollback(ctx)
			return fmt.Errorf("record %s: %w", name, err)
		}
		if err := tx.Commit(ctx); err != nil {
			return fmt.Errorf("commit %s: %w", name, err)
		}
	}

	return nil
}

func loadApplied(ctx context.Context, pool *pgxpool.Pool, schema string) (map[string]bool, error) {
	rows, err := pool.Query(ctx, fmt.Sprintf(`SELECT version FROM %s._migrations`, schema))
	if err != nil {
		return nil, fmt.Errorf("load applied: %w", err)
	}
	defer rows.Close()

	applied := make(map[string]bool)
	for rows.Next() {
		var v string
		if err := rows.Scan(&v); err != nil {
			return nil, err
		}
		applied[v] = true
	}
	return applied, rows.Err()
}
