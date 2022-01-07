package migrations

import (
	"github.com/emvi/logbuch"
	"github.com/muety/broilerplate/config"
	"gorm.io/gorm"
)

func init() {
	const name = "20220107-example_migration"

	f := migrationFunc{
		name: name,
		f: func(db *gorm.DB, cfg *config.Config) error {
			if hasRun(name, db) {
				return nil
			}

			tx := db.Begin()
			if err := tx.Exec("SELECT 1").Error; err != nil {
				logbuch.Warn("unable to do stuff")
			}
			tx.Commit()

			setHasRun(name, db)
			return nil
		},
	}

	registerPostMigration(f)
}
