// Package schema contains the database schema, migrations and seeding data.
package schema

import (
	"context"
	_ "embed" // Calls init function.
	"fmt"

	"github.com/ezratameno/yify/business/sys/database"
	"github.com/jmoiron/sqlx"
)

var (
	//go:embed sql/schema.sql
	schemaDoc string
)

//	Migrate attempts to bring the schema for db up to date with the migrations
//
// defined in this package.
func Migrate(ctx context.Context, db *sqlx.DB) error {
	if err := database.StatusCheck(ctx, db); err != nil {
		return fmt.Errorf("status check database: %w", err)
	}

	// driver, err := darwin.NewGenericDriver(db.DB, darwin.PostgresDialect{})
	// if err != nil {
	// 	return fmt.Errorf("construct darwin driver: %w", err)
	// }
	_, err := db.Exec(schemaDoc)
	// fmt.Println(schemaDoc)
	// d := darwin.New(driver, darwin.ParseMigrations(schemaDoc))
	// return d.Migrate()
	return err
}
