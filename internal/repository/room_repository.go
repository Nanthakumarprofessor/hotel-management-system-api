package repository

import (
	"context"
	"errors"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"hotel-updated/internal/dtos"
	"hotel-updated/internal/errorcodes"
	"hotel-updated/internal/loggers"
	"hotel-updated/internal/models"
	"hotel-updated/pkg/database"
)

// RoomRepositoryInterface defines room repository operations
type RoomRepositoryInterface interface {
	GetRooms(ctx context.Context, checkIn, checkOut time.Time, roomCategoryUUID string, capacity, page, limit int) ([]models.Room, int64, *dtos.PaginationMeta, *dtos.Error)
	GetRoomByUUID(ctx context.Context, roomUUID string) (*models.Room, int, *dtos.Error)
	GetRoomByID(ctx context.Context, roomID int) (*models.Room, int, *dtos.Error)
	IsRoomAvailable(ctx context.Context, roomID int, checkIn, checkOut time.Time, excludeBookingID int) (bool, *dtos.Error)
	GetAllCategories(ctx context.Context) ([]models.RoomCategory, *dtos.Error)
}

// RoomRepository implements RoomRepositoryInterface
type RoomRepository struct {
	db     *database.Db
	logger *loggers.Logger
}

// NewRoomRepository creates a new room repository
func NewRoomRepository(db *database.Db, logger *loggers.Logger) RoomRepositoryInterface {
	return &RoomRepository{
		db:     db,
		logger: logger,
	}
}

func (r *RoomRepository) GetRoomByID(ctx context.Context, roomID int) (*models.Room, int, *dtos.Error) {
	r.logger.Info("GetRoomByID: fetching room", zap.Int("room_id", roomID))

	var room models.Room
	err := r.db.Gorm.WithContext(ctx).
		Where("room_id = ? AND is_active = ?", roomID, true).
		First(&room).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		r.logger.Warn("GetRoomByID: room not found", zap.Int("room_id", roomID))
		return nil, 0, &dtos.Error{
			Code:    errorcodes.HMS_REC_404,
			Message: "Room not found",
		}
	}

	if err != nil {
		r.logger.Error("GetRoomByID: db error", zap.Error(err))
		return nil, 0, &dtos.Error{
			Code:    errorcodes.HMS_DB_001,
			Message: "Failed to fetch room",
		}
	}

	return &room, room.RoomID, nil
}

// GetRooms returns rooms with optional date-range filter.
// When checkIn/checkOut are zero, no availability filter is applied and all rooms are returned.
func (r *RoomRepository) GetRooms(
	ctx context.Context,
	checkIn, checkOut time.Time,
	roomCategoryUUID string,
	capacity, page, limit int,
) ([]models.Room, int64, *dtos.PaginationMeta, *dtos.Error) {

	r.logger.Info("GetRooms: querying rooms",
		zap.Int("page", page),
		zap.Int("limit", limit),
	)

	offset := (page - 1) * limit

	var rooms []models.Room
	var total int64

	baseQuery := r.db.Gorm.WithContext(ctx).
		Table("rooms").
		Joins("JOIN room_categories ON room_categories.room_category_id = rooms.room_category_id").
		Where("rooms.is_active = ?", true).
		Where("room_categories.is_active = ?", true)

	if roomCategoryUUID != "" {
		baseQuery = baseQuery.Where("room_categories.room_category_uuid = ?", roomCategoryUUID)
	}

	if capacity > 0 {
		baseQuery = baseQuery.Where("rooms.capacity >= ?", capacity)
	}

	// Apply booking-overlap filter only when both dates are provided
	if !checkIn.IsZero() && !checkOut.IsZero() {
		baseQuery = baseQuery.Where(`
			rooms.room_id NOT IN (
				SELECT room_id FROM bookings
				WHERE is_active = true
				AND check_in < ?
				AND check_out > ?
			)
		`, checkOut, checkIn)
	}

	countQuery := baseQuery.Session(&gorm.Session{})

	err := countQuery.
		Distinct("rooms.room_id").
		Count(&total).Error
	if err != nil {
		r.logger.Error("GetRooms: count query failed", zap.Error(err))
		return nil, 0, nil, &dtos.Error{
			Code:    errorcodes.HMS_DB_001,
			Message: "Failed to count rooms",
		}
	}

	err = baseQuery.
		Select("rooms.*").
		Offset(offset).
		Limit(limit).
		Find(&rooms).Error

	if err != nil {
		r.logger.Error("GetRooms: fetch query failed", zap.Error(err))
		return nil, 0, nil, &dtos.Error{
			Code:    errorcodes.HMS_DB_001,
			Message: "Failed to fetch rooms",
		}
	}

	totalPages := (total + int64(limit) - 1) / int64(limit)

	pagination := &dtos.PaginationMeta{
		Page:            page,
		Limit:           limit,
		TotalRecords:    total,
		TotalPages:      totalPages,
		HasNextPage:     page < int(totalPages),
		HasPreviousPage: page > 1,
	}

	r.logger.Info("GetRooms: completed", zap.Int("count", len(rooms)), zap.Int64("total", total))
	return rooms, total, pagination, nil
}

// GetRoomByUUID fetches a room by its UUID
func (r *RoomRepository) GetRoomByUUID(ctx context.Context, roomUUID string) (*models.Room, int, *dtos.Error) {
	r.logger.Info("GetRoomByUUID: fetching room", zap.String("room_uuid", roomUUID))

	var room models.Room
	err := r.db.Gorm.WithContext(ctx).
		Where("room_uuid = ? AND is_active = ?", roomUUID, true).
		First(&room).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		r.logger.Warn("GetRoomByUUID: room not found", zap.String("room_uuid", roomUUID))
		return nil, 0, &dtos.Error{
			Code:    errorcodes.HMS_REC_404,
			Message: "Room not found",
		}
	}

	if err != nil {
		r.logger.Error("GetRoomByUUID: db error", zap.Error(err))
		return nil, 0, &dtos.Error{
			Code:    errorcodes.HMS_DB_001,
			Message: "Failed to fetch room",
		}
	}

	return &room, room.RoomID, nil
}

// IsRoomAvailable checks if a room is available for the given date range
func (r *RoomRepository) IsRoomAvailable(ctx context.Context, roomID int, checkIn, checkOut time.Time, excludeBookingID int) (bool, *dtos.Error) {
	r.logger.Info("IsRoomAvailable: checking availability", zap.Int("room_id", roomID))

	var count int64
	query := r.db.Gorm.WithContext(ctx).
		Model(&models.Booking{}).
		Where("room_id = ? AND is_active = true AND check_in < ? AND check_out > ?", roomID, checkOut, checkIn)

	if excludeBookingID > 0 {
		query = query.Where("booking_id != ?", excludeBookingID)
	}

	err := query.Count(&count).Error
	if err != nil {
		r.logger.Error("IsRoomAvailable: db error", zap.Error(err))
		return false, &dtos.Error{
			Code:    errorcodes.HMS_DB_001,
			Message: "Failed to check room availability",
		}
	}

	available := count == 0
	r.logger.Info("IsRoomAvailable: result", zap.Int("room_id", roomID), zap.Bool("available", available))
	return available, nil
}

// GetAllCategories fetches all active room categories
func (r *RoomRepository) GetAllCategories(ctx context.Context) ([]models.RoomCategory, *dtos.Error) {
	r.logger.Info("GetAllCategories: fetching active room categories")

	var categories []models.RoomCategory
	err := r.db.Gorm.WithContext(ctx).
		Where("is_active = ?", true).
		Find(&categories).Error

	if err != nil {
		r.logger.Error("GetAllCategories: db error", zap.Error(err))
		return nil, &dtos.Error{
			Code:    errorcodes.HMS_DB_001,
			Message: "Failed to fetch room categories",
		}
	}

	r.logger.Info("GetAllCategories: completed", zap.Int("count", len(categories)))
	return categories, nil
}
