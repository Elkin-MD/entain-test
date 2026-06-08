package migrations

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

//go:embed sql/*.sql
var files embed.FS

// Run applies migrations in the given direction (up or down).
func Run(args []string) error {
	flags := flag.NewFlagSet("migrate", flag.ContinueOnError)
	direction := flags.String("direction", "up", "migration direction: up or down")

	if err := flags.Parse(args); err != nil {
		return err
	}

	if *direction != "up" && *direction != "down" {
		return fmt.Errorf("invalid direction %q: want up or down", *direction)
	}

	ctx := context.Background()
	cfg := config.Load()

	pool, err := pgxpool.New(ctx, cfg.DSN())
	if err != nil {
		return err
	}

	defer pool.Close()

	names, err := orderedFiles(*direction)
	if err != nil {
		return err
	}

	for _, name := range names {
		query, err := files.ReadFile("sql/" + name)
		if err != nil {
			return err
		}

		if _, err := pool.Exec(ctx, string(query)); err != nil {
			return fmt.Errorf("apply %s: %w", name, err)
		}

		log.Printf("applied %s", name)
	}

	return nil
}

func orderedFiles(direction string) ([]string, error) {
	entries, err := fs.ReadDir(files, "sql")
	if err != nil {
		return nil, err
	}

	suffix := "." + direction + ".sql"

	var names []string

	for _, entry := range entries {
		if strings.HasSuffix(entry.Name(), suffix) {
			names = append(names, entry.Name())
		}
	}

	sort.Strings(names)

	if direction == "down" {
		reverse(names)
	}

	return names, nil
}

func reverse(names []string) {
	for i, j := 0, len(names)-1; i < j; i, j = i+1, j-1 {
		names[i], names[j] = names[j], names[i]
	}
}
