// Package schema contains the database schema, migrations and seeding data.
package schema

import (
	"context"
	_ "embed" // Calls init function.
	"fmt"
	"time"

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
func Migrate(db *sqlx.DB) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := database.StatusCheck(ctx, db); err != nil {
		return fmt.Errorf("status check database: %w", err)
	}
	_, err := db.Exec(schemaDoc)

	return err
}
