package database

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// OpenSQLite opens (or creates) an SQLite database at a given path.
func OpenSQLite(path string) (*gorm.DB, error) {
	return gorm.Open(sqlite.Open(path), &gorm.Config{})
}
