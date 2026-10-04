package mock

import (
	"context"
	"time"

	"hotel-updated/internal/dtos"
	"hotel-updated/internal/models"

	"github.com/stretchr/testify/mock"
)

type RoomRepositoryMock struct{ mock.Mock }

func (m *RoomRepositoryMock) GetRooms(ctx context.Context, checkIn, checkOut time.Time, roomCategoryUUID string, capacity, page, limit int) ([]models.Room, int64, *dtos.PaginationMeta, *dtos.Error) {
	args := m.Called(ctx, checkIn, checkOut, roomCategoryUUID, capacity, page, limit)
	var rooms []models.Room
	if args.Get(0) != nil { rooms = args.Get(0).([]models.Room) }
	var p *dtos.PaginationMeta
	if args.Get(2) != nil { p = args.Get(2).(*dtos.PaginationMeta) }
	var e *dtos.Error
	if args.Get(3) != nil { e = args.Get(3).(*dtos.Error) }
	return rooms, args.Get(1).(int64), p, e
}
func (m *RoomRepositoryMock) GetRoomByUUID(ctx context.Context, roomUUID string) (*models.Room, int, *dtos.Error) {
	args := m.Called(ctx, roomUUID)
	var r *models.Room
	if args.Get(0) != nil { r = args.Get(0).(*models.Room) }
	var e *dtos.Error
	if args.Get(2) != nil { e = args.Get(2).(*dtos.Error) }
	return r, args.Int(1), e
}
func (m *RoomRepositoryMock) GetRoomByID(ctx context.Context, roomID int) (*models.Room, int, *dtos.Error) {
	args := m.Called(ctx, roomID)
	var r *models.Room
	if args.Get(0) != nil { r = args.Get(0).(*models.Room) }
	var e *dtos.Error
	if args.Get(2) != nil { e = args.Get(2).(*dtos.Error) }
	return r, args.Int(1), e
}
func (m *RoomRepositoryMock) IsRoomAvailable(ctx context.Context, roomID int, checkIn, checkOut time.Time, excludeBookingID int) (bool, *dtos.Error) {
	args := m.Called(ctx, roomID, checkIn, checkOut, excludeBookingID)
	var e *dtos.Error
	if args.Get(1) != nil { e = args.Get(1).(*dtos.Error) }
	return args.Bool(0), e
}
func (m *RoomRepositoryMock) GetAllCategories(ctx context.Context) ([]models.RoomCategory, *dtos.Error) {
	args := m.Called(ctx)
	var cats []models.RoomCategory
	if args.Get(0) != nil { cats = args.Get(0).([]models.RoomCategory) }
	var e *dtos.Error
	if args.Get(1) != nil { e = args.Get(1).(*dtos.Error) }
	return cats, e
}
