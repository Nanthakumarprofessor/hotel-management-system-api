package database

import (
	"database/sql"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Db holds database connections
type Db struct {
	Gorm  *gorm.DB
	SqlDb *sql.DB
}

// DBConnector interface for database connection
type DBConnector interface {
	EstablishConnection(dbURL string) (*Db, error)
}

// DBService implements DBConnector
type DBService struct{}

// EstablishConnection establishes a database connection
func (d *DBService) EstablishConnection(dbURL string) (*Db, error) {
	var dialector gorm.Dialector

	
	dialector = postgres.Open(dbURL)

	// Open database connection
	db, err := gorm.Open(dialector, &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// Get underlying sql.DB
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	// Ping the database
	if err := sqlDB.Ping(); err != nil {
		return nil, err
	}

	return &Db{
		Gorm:  db,
		SqlDb: sqlDB,
	}, nil
}
