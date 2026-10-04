package services_test

import (
	"context"
	"net/http"
	"os"
	"testing"
	"time"

	"hotel-updated/internal/dtos"
	"hotel-updated/internal/errorcodes"
	"hotel-updated/internal/loggers"
	"hotel-updated/internal/models"
	"hotel-updated/internal/services"
	repoMock "hotel-updated/tests/unit/repository/mock"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func newTestLogger() *loggers.Logger {
	tmpDir, _ := os.MkdirTemp("", "service-test-logs-*")
	return loggers.NewLogger(loggers.LogConfig{Level: "info", LogDir: tmpDir, FileName: "test.log", ServiceName: "test"})
}

func TestCreateBooking_Success_ExistingGuest(t *testing.T) {
	roomUUID := "550e8400-e29b-41d4-a716-446655440000"
	checkIn := time.Now().Add(24 * time.Hour)
	checkOut := time.Now().Add(72 * time.Hour)
	mockRoom := &models.Room{RoomID: 1, RoomUUID: roomUUID, RoomNo: "101", Price: 150.0}
	mockGuest := &models.Guest{GuestID: 1, GuestUUID: "g-uuid", PhoneNumber: "+1234567890", IsActive: true}
	req := &dtos.CreateBookingRequest{RoomUUID: roomUUID, CheckIn: checkIn.Format(time.RFC3339), CheckOut: checkOut.Format(time.RFC3339), Guest: dtos.GuestRequest{GuestName: "John", Age: 30, Address: "Addr", PhoneNumber: "+1234567890"}}

	bRepo, rRepo := new(repoMock.BookingRepositoryMock), new(repoMock.RoomRepositoryMock)
	rRepo.On("GetRoomByUUID", mock.Anything, roomUUID).Return(mockRoom, 1, nil).Once()
	bRepo.On("FindGuestByPhoneAndReactivate", mock.Anything, "+1234567890", "system").Return(mockGuest, nil).Once()
	rRepo.On("IsRoomAvailable", mock.Anything, 1, mock.AnythingOfType("time.Time"), mock.AnythingOfType("time.Time"), 0).Return(true, nil).Once()
	bRepo.On("CreateBooking", mock.Anything, mock.AnythingOfType("*models.Booking"), "system").Return(nil).Once()

	resp := services.NewBookingService(bRepo, rRepo, newTestLogger()).CreateBooking(context.Background(), req, checkIn, checkOut, roomUUID)
	assert.Equal(t, "Success", resp.Status)
	assert.Equal(t, http.StatusCreated, resp.Code)
	assert.Equal(t, "Booking created successfully", resp.Message)
	bRepo.AssertExpectations(t); rRepo.AssertExpectations(t)
}

func TestCreateBooking_Success_NewGuest(t *testing.T) {
	roomUUID := "550e8400-e29b-41d4-a716-446655440000"
	checkIn, checkOut := time.Now().Add(24*time.Hour), time.Now().Add(72*time.Hour)
	mockRoom := &models.Room{RoomID: 1, RoomUUID: roomUUID, RoomNo: "101", Price: 150.0}
	req := &dtos.CreateBookingRequest{RoomUUID: roomUUID, CheckIn: checkIn.Format(time.RFC3339), CheckOut: checkOut.Format(time.RFC3339), Guest: dtos.GuestRequest{GuestName: "Jane", Age: 25, Address: "Addr", PhoneNumber: "+999"}}

	bRepo, rRepo := new(repoMock.BookingRepositoryMock), new(repoMock.RoomRepositoryMock)
	rRepo.On("GetRoomByUUID", mock.Anything, roomUUID).Return(mockRoom, 1, nil).Once()
	bRepo.On("FindGuestByPhoneAndReactivate", mock.Anything, "+999", "system").Return(nil, nil).Once()
	bRepo.On("CreateGuest", mock.Anything, mock.AnythingOfType("*models.Guest"), "system").Return(nil).Once()
	rRepo.On("IsRoomAvailable", mock.Anything, 1, mock.AnythingOfType("time.Time"), mock.AnythingOfType("time.Time"), 0).Return(true, nil).Once()
	bRepo.On("CreateBooking", mock.Anything, mock.AnythingOfType("*models.Booking"), "system").Return(nil).Once()

	resp := services.NewBookingService(bRepo, rRepo, newTestLogger()).CreateBooking(context.Background(), req, checkIn, checkOut, roomUUID)
	assert.Equal(t, "Success", resp.Status)
	assert.Equal(t, http.StatusCreated, resp.Code)
}

func TestCreateBooking_RoomNotFound(t *testing.T) {
	roomUUID := "550e8400-e29b-41d4-a716-446655440000"
	bRepo, rRepo := new(repoMock.BookingRepositoryMock), new(repoMock.RoomRepositoryMock)
	rRepo.On("GetRoomByUUID", mock.Anything, roomUUID).Return(nil, 0, &dtos.Error{Code: errorcodes.HMS_REC_404, Message: "Room not found"}).Once()
	req := &dtos.CreateBookingRequest{RoomUUID: roomUUID, Guest: dtos.GuestRequest{GuestName: "J", Age: 1, Address: "A", PhoneNumber: "+1"}}
	resp := services.NewBookingService(bRepo, rRepo, newTestLogger()).CreateBooking(context.Background(), req, time.Now(), time.Now().Add(time.Hour), roomUUID)
	assert.Equal(t, "Error", resp.Status)
	assert.Equal(t, "Room not found", resp.Message)
}

func TestCreateBooking_FindGuestError(t *testing.T) {
	roomUUID := "550e8400-e29b-41d4-a716-446655440000"
	checkIn, checkOut := time.Now().Add(24*time.Hour), time.Now().Add(72*time.Hour)
	mockRoom := &models.Room{RoomID: 1, RoomUUID: roomUUID, Price: 100.0}
	bRepo, rRepo := new(repoMock.BookingRepositoryMock), new(repoMock.RoomRepositoryMock)
	rRepo.On("GetRoomByUUID", mock.Anything, roomUUID).Return(mockRoom, 1, nil).Once()
	bRepo.On("FindGuestByPhoneAndReactivate", mock.Anything, "+111", "system").Return(nil, &dtos.Error{Code: errorcodes.HMS_DB_001, Message: "db error"}).Once()
	req := &dtos.CreateBookingRequest{RoomUUID: roomUUID, Guest: dtos.GuestRequest{GuestName: "J", Age: 1, Address: "A", PhoneNumber: "+111"}}
	resp := services.NewBookingService(bRepo, rRepo, newTestLogger()).CreateBooking(context.Background(), req, checkIn, checkOut, roomUUID)
	assert.Equal(t, "Error", resp.Status)
	assert.Equal(t, "Failed to look up guest", resp.Message)
}

func TestCreateBooking_CreateGuestError(t *testing.T) {
	roomUUID := "550e8400-e29b-41d4-a716-446655440000"
	checkIn, checkOut := time.Now().Add(24*time.Hour), time.Now().Add(72*time.Hour)
	mockRoom := &models.Room{RoomID: 1, RoomUUID: roomUUID, Price: 100.0}
	bRepo, rRepo := new(repoMock.BookingRepositoryMock), new(repoMock.RoomRepositoryMock)
	rRepo.On("GetRoomByUUID", mock.Anything, roomUUID).Return(mockRoom, 1, nil).Once()
	bRepo.On("FindGuestByPhoneAndReactivate", mock.Anything, "+111", "system").Return(nil, nil).Once()
	bRepo.On("CreateGuest", mock.Anything, mock.AnythingOfType("*models.Guest"), "system").Return(&dtos.Error{Code: errorcodes.HMS_DB_001, Message: "db error"}).Once()
	req := &dtos.CreateBookingRequest{RoomUUID: roomUUID, Guest: dtos.GuestRequest{GuestName: "J", Age: 1, Address: "A", PhoneNumber: "+111"}}
	resp := services.NewBookingService(bRepo, rRepo, newTestLogger()).CreateBooking(context.Background(), req, checkIn, checkOut, roomUUID)
	assert.Equal(t, "Error", resp.Status)
	assert.Equal(t, "Failed to create guest", resp.Message)
}

func TestCreateBooking_IsRoomAvailableError(t *testing.T) {
	roomUUID := "550e8400-e29b-41d4-a716-446655440000"
	checkIn, checkOut := time.Now().Add(24*time.Hour), time.Now().Add(72*time.Hour)
	mockRoom := &models.Room{RoomID: 1, RoomUUID: roomUUID, Price: 100.0}
	mockGuest := &models.Guest{GuestID: 1, PhoneNumber: "+111", IsActive: true}
	bRepo, rRepo := new(repoMock.BookingRepositoryMock), new(repoMock.RoomRepositoryMock)
	rRepo.On("GetRoomByUUID", mock.Anything, roomUUID).Return(mockRoom, 1, nil).Once()
	bRepo.On("FindGuestByPhoneAndReactivate", mock.Anything, "+111", "system").Return(mockGuest, nil).Once()
	rRepo.On("IsRoomAvailable", mock.Anything, 1, mock.AnythingOfType("time.Time"), mock.AnythingOfType("time.Time"), 0).Return(false, &dtos.Error{Code: errorcodes.HMS_DB_001, Message: "db error"}).Once()
	req := &dtos.CreateBookingRequest{RoomUUID: roomUUID, Guest: dtos.GuestRequest{GuestName: "J", Age: 1, Address: "A", PhoneNumber: "+111"}}
	resp := services.NewBookingService(bRepo, rRepo, newTestLogger()).CreateBooking(context.Background(), req, checkIn, checkOut, roomUUID)
	assert.Equal(t, "Error", resp.Status)
	assert.Equal(t, "Failed to check room availability", resp.Message)
}

func TestCreateBooking_RoomNotAvailable(t *testing.T) {
	roomUUID := "550e8400-e29b-41d4-a716-446655440000"
	checkIn, checkOut := time.Now().Add(24*time.Hour), time.Now().Add(72*time.Hour)
	mockRoom := &models.Room{RoomID: 1, RoomUUID: roomUUID, Price: 100.0}
	mockGuest := &models.Guest{GuestID: 1, PhoneNumber: "+111", IsActive: true}
	bRepo, rRepo := new(repoMock.BookingRepositoryMock), new(repoMock.RoomRepositoryMock)
	rRepo.On("GetRoomByUUID", mock.Anything, roomUUID).Return(mockRoom, 1, nil).Once()
	bRepo.On("FindGuestByPhoneAndReactivate", mock.Anything, "+111", "system").Return(mockGuest, nil).Once()
	rRepo.On("IsRoomAvailable", mock.Anything, 1, mock.AnythingOfType("time.Time"), mock.AnythingOfType("time.Time"), 0).Return(false, nil).Once()
	req := &dtos.CreateBookingRequest{RoomUUID: roomUUID, Guest: dtos.GuestRequest{GuestName: "J", Age: 1, Address: "A", PhoneNumber: "+111"}}
	resp := services.NewBookingService(bRepo, rRepo, newTestLogger()).CreateBooking(context.Background(), req, checkIn, checkOut, roomUUID)
	assert.Equal(t, "Error", resp.Status)
	assert.Equal(t, "Room is not available for the selected dates", resp.Message)
}

func TestCreateBooking_CreateBookingError(t *testing.T) {
	roomUUID := "550e8400-e29b-41d4-a716-446655440000"
	checkIn, checkOut := time.Now().Add(24*time.Hour), time.Now().Add(72*time.Hour)
	mockRoom := &models.Room{RoomID: 1, RoomUUID: roomUUID, Price: 100.0}
	mockGuest := &models.Guest{GuestID: 1, PhoneNumber: "+111", IsActive: true}
	bRepo, rRepo := new(repoMock.BookingRepositoryMock), new(repoMock.RoomRepositoryMock)
	rRepo.On("GetRoomByUUID", mock.Anything, roomUUID).Return(mockRoom, 1, nil).Once()
	bRepo.On("FindGuestByPhoneAndReactivate", mock.Anything, "+111", "system").Return(mockGuest, nil).Once()
	rRepo.On("IsRoomAvailable", mock.Anything, 1, mock.AnythingOfType("time.Time"), mock.AnythingOfType("time.Time"), 0).Return(true, nil).Once()
	bRepo.On("CreateBooking", mock.Anything, mock.AnythingOfType("*models.Booking"), "system").Return(&dtos.Error{Code: errorcodes.HMS_DB_001, Message: "db error"}).Once()
	req := &dtos.CreateBookingRequest{RoomUUID: roomUUID, Guest: dtos.GuestRequest{GuestName: "J", Age: 1, Address: "A", PhoneNumber: "+111"}}
	resp := services.NewBookingService(bRepo, rRepo, newTestLogger()).CreateBooking(context.Background(), req, checkIn, checkOut, roomUUID)
	assert.Equal(t, "Error", resp.Status)
	assert.Equal(t, "Failed to create booking", resp.Message)
}

// ─── ExtendBooking ─────────────────────────────────────────────────────────────

func TestExtendBooking_Success(t *testing.T) {
	now := time.Now()
	bookingUUID := "550e8400-e29b-41d4-a716-446655440001"
	newCheckOut := now.Add(96 * time.Hour)
	active := &models.Booking{BookingID: 1, BookingUUID: bookingUUID, RoomID: 1, GuestID: 1, CheckIn: now, CheckOut: now.Add(48 * time.Hour), IsActive: true}
	mockRoom := &models.Room{RoomID: 1, RoomUUID: "r-uuid", RoomNo: "101", Price: 150.0}
	mockGuest := &models.Guest{GuestID: 1, GuestUUID: "g-uuid"}
	bRepo, rRepo := new(repoMock.BookingRepositoryMock), new(repoMock.RoomRepositoryMock)
	bRepo.On("GetBookingByUUID", mock.Anything, bookingUUID).Return(active, 1, nil).Once()
	rRepo.On("IsRoomAvailable", mock.Anything, 1, mock.AnythingOfType("time.Time"), newCheckOut, 1).Return(true, nil).Once()
	rRepo.On("GetRoomByID", mock.Anything, 1).Return(mockRoom, 1, nil).Once()
	bRepo.On("GetGuestByID", mock.Anything, 1).Return(mockGuest, nil).Once()
	bRepo.On("UpdateCheckOutAndTotalPrice", mock.Anything, 1, newCheckOut, mock.AnythingOfType("float64"), "system").Return(nil).Once()
	resp := services.NewBookingService(bRepo, rRepo, newTestLogger()).ExtendBooking(context.Background(), bookingUUID, newCheckOut)
	assert.Equal(t, "Success", resp.Status)
	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Equal(t, "Booking extended successfully", resp.Message)
}

func TestExtendBooking_BookingNotFound(t *testing.T) {
	bRepo, rRepo := new(repoMock.BookingRepositoryMock), new(repoMock.RoomRepositoryMock)
	bRepo.On("GetBookingByUUID", mock.Anything, "no").Return(nil, 0, &dtos.Error{Code: errorcodes.HMS_REC_404, Message: "Booking not found"}).Once()
	resp := services.NewBookingService(bRepo, rRepo, newTestLogger()).ExtendBooking(context.Background(), "no", time.Now())
	assert.Equal(t, "Error", resp.Status)
	assert.Equal(t, "Booking not found", resp.Message)
}

func TestExtendBooking_CancelledBooking(t *testing.T) {
	now := time.Now()
	bRepo, rRepo := new(repoMock.BookingRepositoryMock), new(repoMock.RoomRepositoryMock)
	cancelled := &models.Booking{BookingID: 1, CheckIn: now, CheckOut: now.Add(48 * time.Hour), IsActive: false}
	bRepo.On("GetBookingByUUID", mock.Anything, "uuid-c").Return(cancelled, 1, nil).Once()
	resp := services.NewBookingService(bRepo, rRepo, newTestLogger()).ExtendBooking(context.Background(), "uuid-c", now.Add(96*time.Hour))
	assert.Equal(t, "Error", resp.Status)
	assert.Equal(t, "Cannot extend a cancelled booking", resp.Message)
}

func TestExtendBooking_CheckOutNotAfterCurrent(t *testing.T) {
	now := time.Now()
	bRepo, rRepo := new(repoMock.BookingRepositoryMock), new(repoMock.RoomRepositoryMock)
	active := &models.Booking{BookingID: 1, CheckIn: now, CheckOut: now.Add(48 * time.Hour), IsActive: true}
	bRepo.On("GetBookingByUUID", mock.Anything, "uuid-v").Return(active, 1, nil).Once()
	resp := services.NewBookingService(bRepo, rRepo, newTestLogger()).ExtendBooking(context.Background(), "uuid-v", now.Add(24*time.Hour))
	assert.Equal(t, "Error", resp.Status)
	assert.Equal(t, "Validation failed", resp.Message)
}

func TestExtendBooking_IsRoomAvailableError(t *testing.T) {
	now := time.Now()
	newCheckOut := now.Add(96 * time.Hour)
	bRepo, rRepo := new(repoMock.BookingRepositoryMock), new(repoMock.RoomRepositoryMock)
	active := &models.Booking{BookingID: 1, RoomID: 1, CheckIn: now, CheckOut: now.Add(48 * time.Hour), IsActive: true}
	bRepo.On("GetBookingByUUID", mock.Anything, "uuid-e").Return(active, 1, nil).Once()
	rRepo.On("IsRoomAvailable", mock.Anything, 1, mock.AnythingOfType("time.Time"), newCheckOut, 1).Return(false, &dtos.Error{Code: errorcodes.HMS_DB_001, Message: "db error"}).Once()
	resp := services.NewBookingService(bRepo, rRepo, newTestLogger()).ExtendBooking(context.Background(), "uuid-e", newCheckOut)
	assert.Equal(t, "Error", resp.Status)
	assert.Equal(t, "Failed to check room availability", resp.Message)
}

func TestExtendBooking_RoomNotAvailableForExtension(t *testing.T) {
	now := time.Now()
	newCheckOut := now.Add(96 * time.Hour)
	bRepo, rRepo := new(repoMock.BookingRepositoryMock), new(repoMock.RoomRepositoryMock)
	active := &models.Booking{BookingID: 1, RoomID: 1, CheckIn: now, CheckOut: now.Add(48 * time.Hour), IsActive: true}
	bRepo.On("GetBookingByUUID", mock.Anything, "uuid-na").Return(active, 1, nil).Once()
	rRepo.On("IsRoomAvailable", mock.Anything, 1, mock.AnythingOfType("time.Time"), newCheckOut, 1).Return(false, nil).Once()
	resp := services.NewBookingService(bRepo, rRepo, newTestLogger()).ExtendBooking(context.Background(), "uuid-na", newCheckOut)
	assert.Equal(t, "Error", resp.Status)
	assert.Equal(t, "Cannot extend booking", resp.Message)
}

func TestExtendBooking_GetRoomByIDError(t *testing.T) {
	now := time.Now()
	newCheckOut := now.Add(96 * time.Hour)
	bRepo, rRepo := new(repoMock.BookingRepositoryMock), new(repoMock.RoomRepositoryMock)
	active := &models.Booking{BookingID: 1, RoomID: 1, CheckIn: now, CheckOut: now.Add(48 * time.Hour), IsActive: true}
	bRepo.On("GetBookingByUUID", mock.Anything, "uuid-gr").Return(active, 1, nil).Once()
	rRepo.On("IsRoomAvailable", mock.Anything, 1, mock.AnythingOfType("time.Time"), newCheckOut, 1).Return(true, nil).Once()
	rRepo.On("GetRoomByID", mock.Anything, 1).Return(nil, 0, &dtos.Error{Code: errorcodes.HMS_DB_001, Message: "room error"}).Once()
	resp := services.NewBookingService(bRepo, rRepo, newTestLogger()).ExtendBooking(context.Background(), "uuid-gr", newCheckOut)
	assert.Equal(t, "Error", resp.Status)
	assert.Equal(t, "Failed to fetch room", resp.Message)
}

func TestExtendBooking_GetGuestByIDError(t *testing.T) {
	now := time.Now()
	newCheckOut := now.Add(96 * time.Hour)
	bRepo, rRepo := new(repoMock.BookingRepositoryMock), new(repoMock.RoomRepositoryMock)
	active := &models.Booking{BookingID: 1, RoomID: 1, GuestID: 1, CheckIn: now, CheckOut: now.Add(48 * time.Hour), IsActive: true}
	mockRoom := &models.Room{RoomID: 1, RoomUUID: "r", RoomNo: "101", Price: 100.0}
	bRepo.On("GetBookingByUUID", mock.Anything, "uuid-gg").Return(active, 1, nil).Once()
	rRepo.On("IsRoomAvailable", mock.Anything, 1, mock.AnythingOfType("time.Time"), newCheckOut, 1).Return(true, nil).Once()
	rRepo.On("GetRoomByID", mock.Anything, 1).Return(mockRoom, 1, nil).Once()
	bRepo.On("GetGuestByID", mock.Anything, 1).Return(nil, &dtos.Error{Code: errorcodes.HMS_DB_001, Message: "guest error"}).Once()
	resp := services.NewBookingService(bRepo, rRepo, newTestLogger()).ExtendBooking(context.Background(), "uuid-gg", newCheckOut)
	assert.Equal(t, "Error", resp.Status)
	assert.Equal(t, "Failed to fetch guest", resp.Message)
}

func TestExtendBooking_UpdateCheckOutError(t *testing.T) {
	now := time.Now()
	newCheckOut := now.Add(96 * time.Hour)
	bRepo, rRepo := new(repoMock.BookingRepositoryMock), new(repoMock.RoomRepositoryMock)
	active := &models.Booking{BookingID: 1, RoomID: 1, GuestID: 1, CheckIn: now, CheckOut: now.Add(48 * time.Hour), IsActive: true}
	mockRoom := &models.Room{RoomID: 1, RoomUUID: "r", RoomNo: "101", Price: 100.0}
	mockGuest := &models.Guest{GuestID: 1, GuestUUID: "g"}
	bRepo.On("GetBookingByUUID", mock.Anything, "uuid-uc").Return(active, 1, nil).Once()
	rRepo.On("IsRoomAvailable", mock.Anything, 1, mock.AnythingOfType("time.Time"), newCheckOut, 1).Return(true, nil).Once()
	rRepo.On("GetRoomByID", mock.Anything, 1).Return(mockRoom, 1, nil).Once()
	bRepo.On("GetGuestByID", mock.Anything, 1).Return(mockGuest, nil).Once()
	bRepo.On("UpdateCheckOutAndTotalPrice", mock.Anything, 1, newCheckOut, mock.AnythingOfType("float64"), "system").Return(&dtos.Error{Code: errorcodes.HMS_DB_001, Message: "update error"}).Once()
	resp := services.NewBookingService(bRepo, rRepo, newTestLogger()).ExtendBooking(context.Background(), "uuid-uc", newCheckOut)
	assert.Equal(t, "Error", resp.Status)
	assert.Equal(t, "Failed to extend booking", resp.Message)
}

// ─── CancelBooking ─────────────────────────────────────────────────────────────

func TestCancelBooking_Success(t *testing.T) {
	bRepo, rRepo := new(repoMock.BookingRepositoryMock), new(repoMock.RoomRepositoryMock)
	active := &models.Booking{BookingID: 1, IsActive: true}
	bRepo.On("GetBookingByUUID", mock.Anything, "c-1").Return(active, 1, nil).Once()
	bRepo.On("CancelBooking", mock.Anything, 1, "system").Return(nil).Once()
	resp := services.NewBookingService(bRepo, rRepo, newTestLogger()).CancelBooking(context.Background(), "c-1")
	assert.Equal(t, "Success", resp.Status)
	assert.Equal(t, "Booking cancelled successfully", resp.Message)
}

func TestCancelBooking_BookingNotFound(t *testing.T) {
	bRepo, rRepo := new(repoMock.BookingRepositoryMock), new(repoMock.RoomRepositoryMock)
	bRepo.On("GetBookingByUUID", mock.Anything, "no").Return(nil, 0, &dtos.Error{Code: errorcodes.HMS_REC_404, Message: "Booking not found"}).Once()
	resp := services.NewBookingService(bRepo, rRepo, newTestLogger()).CancelBooking(context.Background(), "no")
	assert.Equal(t, "Error", resp.Status)
	assert.Equal(t, "Booking not found", resp.Message)
}

func TestCancelBooking_AlreadyCancelled(t *testing.T) {
	bRepo, rRepo := new(repoMock.BookingRepositoryMock), new(repoMock.RoomRepositoryMock)
	bRepo.On("GetBookingByUUID", mock.Anything, "c-2").Return(&models.Booking{BookingID: 1, IsActive: false}, 1, nil).Once()
	resp := services.NewBookingService(bRepo, rRepo, newTestLogger()).CancelBooking(context.Background(), "c-2")
	assert.Equal(t, "Error", resp.Status)
	assert.Equal(t, "Booking is already cancelled", resp.Message)
}

func TestCancelBooking_DBError(t *testing.T) {
	bRepo, rRepo := new(repoMock.BookingRepositoryMock), new(repoMock.RoomRepositoryMock)
	bRepo.On("GetBookingByUUID", mock.Anything, "c-3").Return(&models.Booking{BookingID: 1, IsActive: true}, 1, nil).Once()
	bRepo.On("CancelBooking", mock.Anything, 1, "system").Return(&dtos.Error{Code: errorcodes.HMS_DB_001, Message: "db error"}).Once()
	resp := services.NewBookingService(bRepo, rRepo, newTestLogger()).CancelBooking(context.Background(), "c-3")
	assert.Equal(t, "Error", resp.Status)
	assert.Equal(t, "Failed to cancel booking", resp.Message)
}

// ─── RestoreBooking ────────────────────────────────────────────────────────────

func TestRestoreBooking_Success(t *testing.T) {
	bRepo, rRepo := new(repoMock.BookingRepositoryMock), new(repoMock.RoomRepositoryMock)
	bRepo.On("GetBookingByUUID", mock.Anything, "r-1").Return(&models.Booking{BookingID: 1, IsActive: false}, 1, nil).Once()
	bRepo.On("RestoreBooking", mock.Anything, 1, "system").Return(nil).Once()
	resp := services.NewBookingService(bRepo, rRepo, newTestLogger()).RestoreBooking(context.Background(), "r-1")
	assert.Equal(t, "Success", resp.Status)
	assert.Equal(t, "Booking restored successfully", resp.Message)
}

func TestRestoreBooking_BookingNotFound(t *testing.T) {
	bRepo, rRepo := new(repoMock.BookingRepositoryMock), new(repoMock.RoomRepositoryMock)
	bRepo.On("GetBookingByUUID", mock.Anything, "no").Return(nil, 0, &dtos.Error{Code: errorcodes.HMS_REC_404, Message: "Booking not found"}).Once()
	resp := services.NewBookingService(bRepo, rRepo, newTestLogger()).RestoreBooking(context.Background(), "no")
	assert.Equal(t, "Error", resp.Status)
	assert.Equal(t, "Booking not found", resp.Message)
}

func TestRestoreBooking_AlreadyActive(t *testing.T) {
	bRepo, rRepo := new(repoMock.BookingRepositoryMock), new(repoMock.RoomRepositoryMock)
	bRepo.On("GetBookingByUUID", mock.Anything, "r-2").Return(&models.Booking{BookingID: 1, IsActive: true}, 1, nil).Once()
	resp := services.NewBookingService(bRepo, rRepo, newTestLogger()).RestoreBooking(context.Background(), "r-2")
	assert.Equal(t, "Error", resp.Status)
	assert.Equal(t, "Booking is already active", resp.Message)
}

func TestRestoreBooking_DBError(t *testing.T) {
	bRepo, rRepo := new(repoMock.BookingRepositoryMock), new(repoMock.RoomRepositoryMock)
	bRepo.On("GetBookingByUUID", mock.Anything, "r-3").Return(&models.Booking{BookingID: 1, IsActive: false}, 1, nil).Once()
	bRepo.On("RestoreBooking", mock.Anything, 1, "system").Return(&dtos.Error{Code: errorcodes.HMS_DB_001, Message: "db error"}).Once()
	resp := services.NewBookingService(bRepo, rRepo, newTestLogger()).RestoreBooking(context.Background(), "r-3")
	assert.Equal(t, "Error", resp.Status)
	assert.Equal(t, "Failed to restore booking", resp.Message)
}
