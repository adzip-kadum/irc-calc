package postgres

import (
	"context"
	"embed"
	"os"
	"path/filepath"

	"github.com/jackc/tern/migrate"

	"github.com/adzip-kadum/irc-calc/log"
)

const migrationsTable = "migrations"

func MigrateTo(conf ClientConfig, dir embed.FS, target int32) error {
	pool, err := NewPgxPool(conf)
	if err != nil {
		log.Error(err)
		return err
	}

	ctx := context.Background()

	conn, err := pool.Pool().Acquire(ctx)
	if err != nil {
		log.Error(err)
		return err
	}
	defer conn.Release()

	opts := &migrate.MigratorOptions{
		MigratorFS: migratorFS{dir},
	}

	migrator, err := migrate.NewMigratorEx(ctx, conn.Conn(), migrationsTable, opts)
	if err != nil {
		log.Error(err)
		return err
	}

	err = migrator.LoadMigrations(".")
	if err != nil {
		log.Error(err)
		return err
	}

	version, err := migrator.GetCurrentVersion(ctx)
	if err != nil {
		return err
	}

	if target == 0 {
		target = int32(len(migrator.Migrations))
	}

	log.Info("migration started",
		log.Int32("current-version", version),
		log.Int("last-version", len(migrator.Migrations)),
		log.Int32("target-version", target))

	migrator.OnStart = func(v int32, name, direction, sql string) {
		log.Info("migrating", log.Int32("version", v), log.String("name", name), log.String("direction", direction))
	}

	err = migrator.MigrateTo(ctx, target)
	if err != nil {
		log.Error(err)
		return err
	}

	version, err = migrator.GetCurrentVersion(ctx)
	if err != nil {
		log.Error(err)
		return err
	}

	log.Info("migration finished", log.Int32("current-version", version))

	return nil
}

type migratorFS struct {
	embed.FS
}

func (m migratorFS) ReadDir(dirname string) ([]os.FileInfo, error) {
	entries, err := m.FS.ReadDir(dirname)
	if err != nil {
		return nil, err
	}
	infos := []os.FileInfo{}
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			return nil, err
		}
		infos = append(infos, info)
	}
	return infos, nil
}

func (m migratorFS) ReadFile(filename string) ([]byte, error) {
	return m.FS.ReadFile(filename)
}

func (m migratorFS) Glob(pattern string) ([]string, error) {
	pattern = filepath.Base(pattern)
	matches := []string{}
	entries, err := m.FS.ReadDir(".")
	if err != nil {
		return nil, err
	}
	for _, entry := range entries {
		matched, err := filepath.Match(pattern, entry.Name())
		if err != nil {
			return nil, err
		}
		if matched {
			matches = append(matches, entry.Name())
		}
	}
	return matches, nil
}
