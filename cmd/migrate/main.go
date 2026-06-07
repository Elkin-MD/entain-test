// Command migrate applies database migrations up or down.
package main

import (
	"context"
	"embed"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"sort"
	"strings"

	"entaintest/internal/config"

	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

func main() {
	direction := flag.String("direction", "up", "migration direction: up or down")
	flag.Parse()

	if err := run(*direction); err != nil {
		log.Fatal(err)
	}
}

func run(direction string) error {
	if direction != "up" && direction != "down" {
		return fmt.Errorf("invalid direction %q: want up or down", direction)
	}

	ctx := context.Background()
	cfg := config.Load()

	pool, err := pgxpool.New(ctx, cfg.DSN())
	if err != nil {
		return err
	}

	defer pool.Close()

	files, err := migrationFiles(direction)
	if err != nil {
		return err
	}

	for _, file := range files {
		query, err := migrationsFS.ReadFile("migrations/" + file)
		if err != nil {
			return err
		}

		if _, err := pool.Exec(ctx, string(query)); err != nil {
			return fmt.Errorf("apply %s: %w", file, err)
		}

		log.Printf("applied %s", file)
	}

	return nil
}

func migrationFiles(direction string) ([]string, error) {
	entries, err := fs.ReadDir(migrationsFS, "migrations")
	if err != nil {
		return nil, err
	}

	suffix := "." + direction + ".sql"

	var files []string

	for _, entry := range entries {
		if strings.HasSuffix(entry.Name(), suffix) {
			files = append(files, entry.Name())
		}
	}

	sort.Strings(files)

	// Down migrations apply in reverse version order.
	if direction == "down" {
		for i, j := 0, len(files)-1; i < j; i, j = i+1, j-1 {
			files[i], files[j] = files[j], files[i]
		}
	}

	return files, nil
}
