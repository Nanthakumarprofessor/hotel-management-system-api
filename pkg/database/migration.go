package database

import (
	"go.uber.org/zap"

	"hotel-updated/internal/loggers"
	"hotel-updated/internal/models"
)

// AutoMigrate runs database migrations for all models.
func AutoMigrate(db *Db, logger *loggers.Logger) error {
	logger.Info("Running database migrations")

	// db.Gorm.DisableForeignKeyConstraintWhenMigrating = true

	err := db.Gorm.AutoMigrate(
		&models.RoomCategory{},
		&models.Guest{},
		&models.Room{},
		&models.Booking{},
	)
	if err != nil {
		logger.Error("Auto migration failed", zap.Error(err))
		return err
	}

	logger.Info("Auto migration completed successfully")
	return nil
}
