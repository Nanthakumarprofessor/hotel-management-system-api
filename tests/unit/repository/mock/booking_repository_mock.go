package mock

import (
	"context"
	"time"

	"hotel-updated/internal/dtos"
	"hotel-updated/internal/models"

	"github.com/stretchr/testify/mock"
)

type BookingRepositoryMock struct{ mock.Mock }

func (m *BookingRepositoryMock) CreateBooking(ctx context.Context, booking *models.Booking, createdBy string) *dtos.Error {
	args := m.Called(ctx, booking, createdBy)
	if args.Get(0) == nil { return nil }
	return args.Get(0).(*dtos.Error)
}
func (m *BookingRepositoryMock) GetBookingByUUID(ctx context.Context, bookingUUID string) (*models.Booking, int, *dtos.Error) {
	args := m.Called(ctx, bookingUUID)
	var b *models.Booking
	if args.Get(0) != nil { b = args.Get(0).(*models.Booking) }
	var e *dtos.Error
	if args.Get(2) != nil { e = args.Get(2).(*dtos.Error) }
	return b, args.Int(1), e
}
func (m *BookingRepositoryMock) GetGuestByID(ctx context.Context, guestID int) (*models.Guest, *dtos.Error) {
	args := m.Called(ctx, guestID)
	var g *models.Guest
	if args.Get(0) != nil { g = args.Get(0).(*models.Guest) }
	var e *dtos.Error
	if args.Get(1) != nil { e = args.Get(1).(*dtos.Error) }
	return g, e
}
func (m *BookingRepositoryMock) UpdateCheckOutAndTotalPrice(ctx context.Context, bookingID int, checkOut time.Time, totalPrice float64, updatedBy string) *dtos.Error {
	args := m.Called(ctx, bookingID, checkOut, totalPrice, updatedBy)
	if args.Get(0) == nil { return nil }
	return args.Get(0).(*dtos.Error)
}
func (m *BookingRepositoryMock) CancelBooking(ctx context.Context, bookingID int, cancelledBy string) *dtos.Error {
	args := m.Called(ctx, bookingID, cancelledBy)
	if args.Get(0) == nil { return nil }
	return args.Get(0).(*dtos.Error)
}
func (m *BookingRepositoryMock) RestoreBooking(ctx context.Context, bookingID int, restoredBy string) *dtos.Error {
	args := m.Called(ctx, bookingID, restoredBy)
	if args.Get(0) == nil { return nil }
	return args.Get(0).(*dtos.Error)
}
func (m *BookingRepositoryMock) FindGuestByPhoneAndReactivate(ctx context.Context, phoneNumber string, createdBy string) (*models.Guest, *dtos.Error) {
	args := m.Called(ctx, phoneNumber, createdBy)
	var g *models.Guest
	if args.Get(0) != nil { g = args.Get(0).(*models.Guest) }
	var e *dtos.Error
	if args.Get(1) != nil { e = args.Get(1).(*dtos.Error) }
	return g, e
}
func (m *BookingRepositoryMock) CreateGuest(ctx context.Context, guest *models.Guest, createdBy string) *dtos.Error {
	args := m.Called(ctx, guest, createdBy)
	if args.Get(0) == nil { return nil }
	return args.Get(0).(*dtos.Error)
}
