package mock

import (
	"context"
	"time"

	"hotel-updated/internal/dtos"

	"github.com/stretchr/testify/mock"
)

type BookingServiceMock struct{ mock.Mock }

func (m *BookingServiceMock) CreateBooking(ctx context.Context, req *dtos.CreateBookingRequest, checkIn, checkOut time.Time, roomUUID string) *dtos.APIResponse {
	return m.Called(ctx, req, checkIn, checkOut, roomUUID).Get(0).(*dtos.APIResponse)
}
func (m *BookingServiceMock) ExtendBooking(ctx context.Context, bookingUUID string, newCheckOut time.Time) *dtos.APIResponse {
	return m.Called(ctx, bookingUUID, newCheckOut).Get(0).(*dtos.APIResponse)
}
func (m *BookingServiceMock) CancelBooking(ctx context.Context, bookingUUID string) *dtos.APIResponse {
	return m.Called(ctx, bookingUUID).Get(0).(*dtos.APIResponse)
}
func (m *BookingServiceMock) RestoreBooking(ctx context.Context, bookingUUID string) *dtos.APIResponse {
	return m.Called(ctx, bookingUUID).Get(0).(*dtos.APIResponse)
}
