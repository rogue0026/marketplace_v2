package migrations

import (
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func Run(migrationsPath string, databaseURL string) error {
	m, err := migrate.New("file://"+migrationsPath, databaseURL)
	if err != nil {
		return fmt.Errorf("creating migrate instance: %w", err)
	}

	err = m.Up()
	if err != nil {
		return fmt.Errorf("fail while applying migrations: %w", err)
	}

	srcErr, dbErr := m.Close()
	if srcErr != nil {
		return srcErr
	}

	if dbErr != nil {
		return dbErr
	}

	fmt.Println("migrations applied successfully")
	return nil
}
