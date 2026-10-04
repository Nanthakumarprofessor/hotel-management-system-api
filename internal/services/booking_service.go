package services

import (
	"context"
	"math"
	"net/http"
	"time"

	"go.uber.org/zap"

	"hotel-updated/internal/dtos"
	"hotel-updated/internal/errorcodes"
	"hotel-updated/internal/loggers"
	"hotel-updated/internal/models"
	"hotel-updated/internal/repository"
	"hotel-updated/internal/utils"
)

// BookingServiceInterface defines booking service operations
type BookingServiceInterface interface {
	CreateBooking(ctx context.Context, req *dtos.CreateBookingRequest, checkIn, checkOut time.Time, roomUUID string) *dtos.APIResponse
	ExtendBooking(ctx context.Context, bookingUUID string, newCheckOut time.Time) *dtos.APIResponse
	CancelBooking(ctx context.Context, bookingUUID string) *dtos.APIResponse
	RestoreBooking(ctx context.Context, bookingUUID string) *dtos.APIResponse
}

// BookingService implements BookingServiceInterface
type BookingService struct {
	bookingRepo repository.BookingRepositoryInterface
	roomRepo    repository.RoomRepositoryInterface
	logger      *loggers.Logger
}

// NewBookingService creates a new booking service
func NewBookingService(bookingRepo repository.BookingRepositoryInterface, roomRepo repository.RoomRepositoryInterface, logger *loggers.Logger) BookingServiceInterface {
	return &BookingService{
		bookingRepo: bookingRepo,
		roomRepo:    roomRepo,
		logger:      logger,
	}
}

// CreateBooking creates a new booking
func (s *BookingService) CreateBooking(ctx context.Context, req *dtos.CreateBookingRequest, checkIn, checkOut time.Time, roomUUID string) *dtos.APIResponse {
	s.logger.Info("CreateBooking: processing new booking request")

	room, roomID, err := s.roomRepo.GetRoomByUUID(ctx, roomUUID)
	if err != nil {
		return utils.ErrorResponse("Room not found", utils.SingleError("room_uuid", err.Message, err.Code))
	}

	guest, err := s.bookingRepo.FindGuestByPhoneAndReactivate(ctx, req.Guest.PhoneNumber, "system")
	if err != nil {
		return utils.ErrorResponse("Failed to look up guest", utils.SingleError("guest.phone_number", err.Message, err.Code))
	}

	if guest == nil {
		newGuest := &models.Guest{
			GuestUUID:   utils.GenerateUUID(),
			GuestName:   req.Guest.GuestName,
			GuestAge:    req.Guest.Age,
			Address:     req.Guest.Address,
			PhoneNumber: req.Guest.PhoneNumber,
			IsActive:    true,
		}

		err = s.bookingRepo.CreateGuest(ctx, newGuest, "system")
		if err != nil {
			return utils.ErrorResponse("Failed to create guest", utils.SingleError("guest", err.Message, err.Code))
		}
		guest = newGuest
	}

	available, err := s.roomRepo.IsRoomAvailable(ctx, roomID, checkIn, checkOut, 0)
	if err != nil {
		return utils.ErrorResponse("Failed to check room availability", utils.SingleError("room_uuid", err.Message, err.Code))
	}

	if !available {
		return utils.ErrorResponse("Room is not available for the selected dates", utils.SingleError("room_uuid", "Room is already booked for the selected date range", errorcodes.HMS_CONFLICT_409))
	}

	totalDays := int(math.Ceil(checkOut.Sub(checkIn).Hours() / 24))
	totalPrice := float64(totalDays) * room.Price

	booking := &models.Booking{
		BookingUUID: utils.GenerateUUID(),
		GuestID:     guest.GuestID,
		RoomID:      roomID,
		CheckIn:     checkIn,
		CheckOut:    checkOut,
		TotalPrice:  totalPrice,
		IsActive:    true,
	}

	err = s.bookingRepo.CreateBooking(ctx, booking, "system")
	if err != nil {
		return utils.ErrorResponse("Failed to create booking", utils.SingleError("", err.Message, err.Code))
	}

	responseData := dtos.BookingData{
		BookingUUID: booking.BookingUUID,
		GuestUUID:   guest.GuestUUID,
		RoomUUID:    room.RoomUUID,
		RoomNo:      room.RoomNo,
		CheckIn:     booking.CheckIn.Format(time.RFC3339),
		CheckOut:    booking.CheckOut.Format(time.RFC3339),
		TotalPrice:  totalPrice,
		TotalDays:   totalDays,
	}

	s.logger.Info("CreateBooking: booking created successfully", zap.String("booking_uuid", booking.BookingUUID))
	return utils.SuccessResponse("Booking created successfully", http.StatusCreated, responseData)
}

// ExtendBooking extends the check-out date of an existing booking
func (s *BookingService) ExtendBooking(ctx context.Context, bookingUUID string, newCheckOut time.Time) *dtos.APIResponse {
	s.logger.Info("ExtendBooking: processing extend request")

	booking, bookingID, err := s.bookingRepo.GetBookingByUUID(ctx, bookingUUID)
	if err != nil {
		return utils.ErrorResponse("Booking not found", utils.SingleError("booking_uuid", err.Message, err.Code))
	}

	if !booking.IsActive {
		return utils.ErrorResponse("Cannot extend a cancelled booking", utils.SingleError("booking_uuid", "Booking is not active", errorcodes.HMS_VAL_002))
	}

	if !newCheckOut.After(booking.CheckOut) {
		return utils.ErrorResponse("Validation failed", utils.SingleError("check_out", "new check_out must be after the current check_out", errorcodes.HMS_VAL_001))
	}

	available, err := s.roomRepo.IsRoomAvailable(ctx, booking.RoomID, booking.CheckOut, newCheckOut, bookingID)
	if err != nil {
		return utils.ErrorResponse("Failed to check room availability", utils.SingleError("", err.Message, err.Code))
	}

	if !available {
		return utils.ErrorResponse("Cannot extend booking", utils.SingleError("check_out", "Room is already booked during the extended period", errorcodes.HMS_CONFLICT_409))
	}

	room, _, err := s.roomRepo.GetRoomByID(ctx, booking.RoomID)
	if err != nil {
		return utils.ErrorResponse("Failed to fetch room", utils.SingleError("", err.Message, err.Code))
	}

	guest, err := s.bookingRepo.GetGuestByID(ctx, booking.GuestID)
	if err != nil {
		return utils.ErrorResponse("Failed to fetch guest", utils.SingleError("", err.Message, err.Code))
	}

	totalDays := int(math.Ceil(newCheckOut.Sub(booking.CheckIn).Hours() / 24))
	totalPrice := float64(totalDays) * room.Price

	err = s.bookingRepo.UpdateCheckOutAndTotalPrice(ctx, bookingID, newCheckOut, totalPrice, "system")
	if err != nil {
		return utils.ErrorResponse("Failed to extend booking", utils.SingleError("", err.Message, err.Code))
	}

	responseData := dtos.BookingData{
		BookingUUID: booking.BookingUUID,
		GuestUUID:   guest.GuestUUID,
		RoomUUID:    room.RoomUUID,
		RoomNo:      room.RoomNo,
		CheckIn:     booking.CheckIn.Format(time.RFC3339),
		CheckOut:    newCheckOut.Format(time.RFC3339),
		TotalPrice:  totalPrice,
		TotalDays:   totalDays,
	}

	s.logger.Info("ExtendBooking: booking extended successfully", zap.String("booking_uuid", bookingUUID))
	return utils.SuccessResponse("Booking extended successfully", http.StatusOK, responseData)
}

// CancelBooking cancels an existing booking (soft delete)
func (s *BookingService) CancelBooking(ctx context.Context, bookingUUID string) *dtos.APIResponse {
	s.logger.Info("CancelBooking: processing cancel request")

	booking, bookingID, err := s.bookingRepo.GetBookingByUUID(ctx, bookingUUID)
	if err != nil {
		return utils.ErrorResponse("Booking not found", utils.SingleError("booking_uuid", err.Message, err.Code))
	}

	if !booking.IsActive {
		return utils.ErrorResponse("Booking is already cancelled", utils.SingleError("booking_uuid", "Booking is already cancelled", errorcodes.HMS_CONFLICT_409))
	}

	err = s.bookingRepo.CancelBooking(ctx, bookingID, "system")
	if err != nil {
		return utils.ErrorResponse("Failed to cancel booking", utils.SingleError("", err.Message, err.Code))
	}

	s.logger.Info("CancelBooking: booking cancelled", zap.String("booking_uuid", bookingUUID))
	return utils.SuccessResponse("Booking cancelled successfully", http.StatusOK, dtos.CancelData{Message: "Booking cancelled"})
}

// RestoreBooking restores a cancelled booking
func (s *BookingService) RestoreBooking(ctx context.Context, bookingUUID string) *dtos.APIResponse {
	s.logger.Info("RestoreBooking: processing restore request")

	booking, bookingID, err := s.bookingRepo.GetBookingByUUID(ctx, bookingUUID)
	if err != nil {
		return utils.ErrorResponse("Booking not found", utils.SingleError("booking_uuid", err.Message, err.Code))
	}

	if booking.IsActive {
		return utils.ErrorResponse("Booking is already active", utils.SingleError("booking_uuid", "Only cancelled bookings can be restored", errorcodes.HMS_CONFLICT_409))
	}

	err = s.bookingRepo.RestoreBooking(ctx, bookingID, "system")
	if err != nil {
		return utils.ErrorResponse("Failed to restore booking", utils.SingleError("", err.Message, err.Code))
	}

	s.logger.Info("RestoreBooking: booking restored", zap.String("booking_uuid", bookingUUID))
	return utils.SuccessResponse("Booking restored successfully", http.StatusOK, dtos.RestoreData{Message: "Booking restored"})
}
