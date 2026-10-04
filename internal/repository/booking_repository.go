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

// BookingRepositoryInterface defines booking repository operations
type BookingRepositoryInterface interface {
	CreateBooking(ctx context.Context, booking *models.Booking, createdBy string) *dtos.Error
	GetBookingByUUID(ctx context.Context, bookingUUID string) (*models.Booking, int, *dtos.Error)
	GetGuestByID(ctx context.Context, guestID int) (*models.Guest, *dtos.Error)
	UpdateCheckOutAndTotalPrice(ctx context.Context, bookingID int, checkOut time.Time, totalPrice float64, updatedBy string) *dtos.Error
	CancelBooking(ctx context.Context, bookingID int, cancelledBy string) *dtos.Error
	RestoreBooking(ctx context.Context, bookingID int, restoredBy string) *dtos.Error
	FindGuestByPhoneAndReactivate(ctx context.Context, phoneNumber string, createdBy string) (*models.Guest, *dtos.Error)
	CreateGuest(ctx context.Context, guest *models.Guest, createdBy string) *dtos.Error
}

// BookingRepository implements BookingRepositoryInterface
type BookingRepository struct {
	db     *database.Db
	logger *loggers.Logger
}

// NewBookingRepository creates a new booking repository
func NewBookingRepository(db *database.Db, logger *loggers.Logger) BookingRepositoryInterface {
	return &BookingRepository{
		db:     db,
		logger: logger,
	}
}

func (r *BookingRepository) GetGuestByID(ctx context.Context, guestID int) (*models.Guest, *dtos.Error) {
	r.logger.Info("GetGuestByID: fetching guest", zap.Int("guest_id", guestID))

	var guest models.Guest
	err := r.db.Gorm.WithContext(ctx).Where("guest_id = ?", guestID).First(&guest).Error
	if err != nil {
		r.logger.Error("GetGuestByID: db error", zap.Error(err))
		return nil, &dtos.Error{
			Code:    errorcodes.HMS_DB_001,
			Message: "Failed to fetch guest",
		}
	}
	return &guest, nil
}

func (r *BookingRepository) CreateBooking(ctx context.Context, booking *models.Booking, createdBy string) *dtos.Error {
	r.logger.Info("CreateBooking: inserting booking record")

	booking.CreatedBy = createdBy
	booking.UpdatedBy = createdBy

	err := r.db.Gorm.WithContext(ctx).Create(booking).Error
	if err != nil {
		r.logger.Error("CreateBooking: db error", zap.Error(err))
		return &dtos.Error{
			Code:    errorcodes.HMS_DB_001,
			Message: "Failed to create booking",
		}
	}

	r.logger.Info("CreateBooking: booking record created", zap.String("booking_uuid", booking.BookingUUID))
	return nil
}

func (r *BookingRepository) GetBookingByUUID(ctx context.Context, bookingUUID string) (*models.Booking, int, *dtos.Error) {
	r.logger.Info("GetBookingByUUID: fetching booking")

	var booking models.Booking
	err := r.db.Gorm.WithContext(ctx).Model(&models.Booking{}).
		Where("booking_uuid = ?", bookingUUID).
		First(&booking).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		r.logger.Warn("GetBookingByUUID: booking not found")
		return nil, 0, &dtos.Error{
			Code:    errorcodes.HMS_REC_404,
			Message: "Booking not found",
		}
	}

	if err != nil {
		r.logger.Error("GetBookingByUUID: db error", zap.Error(err))
		return nil, 0, &dtos.Error{
			Code:    errorcodes.HMS_DB_001,
			Message: "Failed to fetch booking",
		}
	}

	return &booking, booking.BookingID, nil
}

func (r *BookingRepository) UpdateCheckOutAndTotalPrice(ctx context.Context, bookingID int, checkOut time.Time, totalprice float64, updatedBy string) *dtos.Error {
	r.logger.Info("UpdateCheckOutAndTotalPrice: updating booking")

	updates := map[string]interface{}{
		"total_price": totalprice,
		"check_out":   checkOut,
		"updated_by":  updatedBy,
	}

	result := r.db.Gorm.WithContext(ctx).
		Model(&models.Booking{}).
		Where("booking_id = ?", bookingID).
		Updates(updates)

	if result.Error != nil {
		r.logger.Error("UpdateCheckOutAndTotalPrice: db error", zap.Error(result.Error))
		return &dtos.Error{
			Code:    errorcodes.HMS_DB_001,
			Message: "Failed to update booking",
		}
	}

	if result.RowsAffected == 0 {
		r.logger.Warn("UpdateCheckOutAndTotalPrice: booking not found")
		return &dtos.Error{
			Code:    errorcodes.HMS_REC_404,
			Message: "Booking not found",
		}
	}

	r.logger.Info("UpdateCheckOutAndTotalPrice: booking updated")
	return nil
}

func (r *BookingRepository) CancelBooking(ctx context.Context, bookingID int, cancelledBy string) *dtos.Error {
	r.logger.Info("CancelBooking: cancelling booking record")

	updates := map[string]interface{}{
		"is_active":  false,
		"updated_by": cancelledBy,
	}

	result := r.db.Gorm.WithContext(ctx).
		Model(&models.Booking{}).
		Where("booking_id = ?", bookingID).
		Updates(updates)

	if result.Error != nil {
		r.logger.Error("CancelBooking: db error", zap.Error(result.Error))
		return &dtos.Error{
			Code:    errorcodes.HMS_DB_001,
			Message: "Failed to cancel booking",
		}
	}

	if result.RowsAffected == 0 {
		r.logger.Warn("CancelBooking: booking not found")
		return &dtos.Error{
			Code:    errorcodes.HMS_REC_404,
			Message: "Booking not found",
		}
	}

	r.logger.Info("CancelBooking: booking cancelled successfully")
	return nil
}

func (r *BookingRepository) RestoreBooking(ctx context.Context, bookingID int, restoredBy string) *dtos.Error {
	r.logger.Info("RestoreBooking: restoring booking record")

	updates := map[string]interface{}{
		"is_active":  true,
		"updated_by": restoredBy,
	}

	result := r.db.Gorm.WithContext(ctx).
		Model(&models.Booking{}).
		Where("booking_id = ?", bookingID).
		Updates(updates)

	if result.Error != nil {
		r.logger.Error("RestoreBooking: db error", zap.Error(result.Error))
		return &dtos.Error{
			Code:    errorcodes.HMS_DB_001,
			Message: "Failed to restore booking",
		}
	}

	if result.RowsAffected == 0 {
		r.logger.Warn("RestoreBooking: booking not found")
		return &dtos.Error{
			Code:    errorcodes.HMS_REC_404,
			Message: "Booking not found",
		}
	}

	r.logger.Info("RestoreBooking: booking restored successfully")
	return nil
}

func (r *BookingRepository) FindGuestByPhoneAndReactivate(ctx context.Context, phoneNumber string, createdBy string) (*models.Guest, *dtos.Error) {
	r.logger.Info("FindGuestByPhoneAndReactivate: looking up guest by phone")

	var guest models.Guest
	err := r.db.Gorm.WithContext(ctx).
		Where("phone_number = ?", phoneNumber).
		First(&guest).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		r.logger.Info("FindGuestByPhoneAndReactivate: no existing guest found")
		return nil, nil
	}

	if err != nil {
		r.logger.Error("FindGuestByPhoneAndReactivate: db error", zap.Error(err))
		return nil, &dtos.Error{
			Code:    errorcodes.HMS_DB_001,
			Message: "Failed to look up guest",
		}
	}

	if !guest.IsActive {
		r.logger.Info("FindGuestByPhoneAndReactivate: reactivating inactive guest")

		updates := map[string]interface{}{
			"is_active":  true,
			"updated_by": createdBy,
		}

		err = r.db.Gorm.WithContext(ctx).
			Model(&guest).
			Updates(updates).Error

		if err != nil {
			r.logger.Error("FindGuestByPhoneAndReactivate: reactivation failed", zap.Error(err))
			return nil, &dtos.Error{
				Code:    errorcodes.HMS_DB_001,
				Message: "Failed to reactivate guest",
			}
		}
		guest.IsActive = true
		guest.UpdatedBy = createdBy
	}

	r.logger.Info("FindGuestByPhoneAndReactivate: guest resolved", zap.String("guest_uuid", guest.GuestUUID))
	return &guest, nil
}

func (r *BookingRepository) CreateGuest(ctx context.Context, guest *models.Guest, createdBy string) *dtos.Error {
	r.logger.Info("CreateGuest: inserting new guest record")

	guest.CreatedBy = createdBy
	guest.UpdatedBy = createdBy

	err := r.db.Gorm.WithContext(ctx).Create(guest).Error
	if err != nil {
		r.logger.Error("CreateGuest: db error", zap.Error(err))
		return &dtos.Error{
			Code:    errorcodes.HMS_DB_001,
			Message: "Failed to create guest",
		}
	}

	r.logger.Info("CreateGuest: guest created", zap.String("guest_uuid", guest.GuestUUID))
	return nil
}
