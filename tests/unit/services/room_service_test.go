package services_test

import (
	"context"
	"net/http"
	"testing"
	"time"

	"hotel-updated/internal/dtos"
	"hotel-updated/internal/errorcodes"
	"hotel-updated/internal/models"
	"hotel-updated/internal/services"
	repoMock "hotel-updated/tests/unit/repository/mock"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// ─── GetRoomCategories ─────────────────────────────────────────────────────────

func TestGetRoomCategories_Success(t *testing.T) {
	rRepo := new(repoMock.RoomRepositoryMock)
	cats := []models.RoomCategory{
		{RoomCategoryID: 1, RoomCategoryUUID: "cat-1", RoomCategoryName: "Standard"},
		{RoomCategoryID: 2, RoomCategoryUUID: "cat-2", RoomCategoryName: "Deluxe"},
	}
	rRepo.On("GetAllCategories", mock.Anything).Return(cats, nil).Once()

	svc := services.NewRoomService(rRepo, newTestLogger())
	resp := svc.GetRoomCategories(context.Background())

	assert.Equal(t, "Success", resp.Status)
	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Equal(t, "Room categories retrieved successfully", resp.Message)
	data, ok := resp.Data.(dtos.RoomCategoryListData)
	assert.True(t, ok)
	assert.Equal(t, 2, data.Total)
	assert.Len(t, data.Categories, 2)
	assert.Equal(t, "cat-1", data.Categories[0].RoomCategoryUUID)
	assert.Equal(t, "Standard", data.Categories[0].RoomCategoryName)
	rRepo.AssertExpectations(t)
}

func TestGetRoomCategories_Empty(t *testing.T) {
	rRepo := new(repoMock.RoomRepositoryMock)
	rRepo.On("GetAllCategories", mock.Anything).Return([]models.RoomCategory{}, nil).Once()

	svc := services.NewRoomService(rRepo, newTestLogger())
	resp := svc.GetRoomCategories(context.Background())

	assert.Equal(t, "Error", resp.Status)
	assert.Equal(t, "No room categories found", resp.Message)
}

func TestGetRoomCategories_DBError(t *testing.T) {
	rRepo := new(repoMock.RoomRepositoryMock)
	rRepo.On("GetAllCategories", mock.Anything).Return(nil, &dtos.Error{Code: errorcodes.HMS_DB_001, Message: "db error"}).Once()

	svc := services.NewRoomService(rRepo, newTestLogger())
	resp := svc.GetRoomCategories(context.Background())

	assert.Equal(t, "Error", resp.Status)
	assert.Equal(t, "Failed to fetch room categories", resp.Message)
}

// ─── GetAvailableRooms ─────────────────────────────────────────────────────────

func TestGetAvailableRooms_Success(t *testing.T) {
	rRepo := new(repoMock.RoomRepositoryMock)
	checkIn := time.Now().Add(24 * time.Hour)
	checkOut := time.Now().Add(72 * time.Hour)

	rooms := []models.Room{
		{RoomID: 1, RoomUUID: "r-1", RoomNo: "101", Price: 100.0, IsActive: true},
		{RoomID: 2, RoomUUID: "r-2", RoomNo: "102", Price: 150.0, IsActive: true},
	}
	pagination := &dtos.PaginationMeta{Page: 1, Limit: 10, TotalRecords: 2, TotalPages: 1}

	req := dtos.GetAvailableRoomsRequest{
		CheckIn:  checkIn.Format(time.RFC3339),
		CheckOut: checkOut.Format(time.RFC3339),
		Page:     1,
		Limit:    10,
	}

	rRepo.On("GetRooms", mock.Anything,
		mock.AnythingOfType("time.Time"),
		mock.AnythingOfType("time.Time"),
		"", 0, 1, 10,
	).Return(rooms, int64(2), pagination, nil).Once()

	svc := services.NewRoomService(rRepo, newTestLogger())
	resp := svc.GetAvailableRooms(context.Background(), req)

	assert.Equal(t, "Success", resp.Status)
	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Equal(t, "Rooms retrieved successfully", resp.Message)
	data, ok := resp.Data.(dtos.RoomListData)
	assert.True(t, ok)
	assert.Len(t, data.Rooms, 2)
	assert.Equal(t, "r-1", data.Rooms[0].RoomUUID)
	assert.True(t, data.Rooms[0].IsActive)
	rRepo.AssertExpectations(t)
}

func TestGetAvailableRooms_WithFilters(t *testing.T) {
	rRepo := new(repoMock.RoomRepositoryMock)
	catUUID := "550e8400-e29b-41d4-a716-446655440000"
	rooms := []models.Room{
		{RoomID: 1, RoomUUID: "r-1", RoomNo: "101", Price: 200.0, IsActive: true},
	}
	pagination := &dtos.PaginationMeta{Page: 1, Limit: 5, TotalRecords: 1, TotalPages: 1}

	req := dtos.GetAvailableRoomsRequest{
		RoomCategoryUUID: catUUID,
		Capacity:         2,
		Page:             1,
		Limit:            5,
	}

	rRepo.On("GetRooms", mock.Anything,
		mock.AnythingOfType("time.Time"),
		mock.AnythingOfType("time.Time"),
		catUUID, 2, 1, 5,
	).Return(rooms, int64(1), pagination, nil).Once()

	svc := services.NewRoomService(rRepo, newTestLogger())
	resp := svc.GetAvailableRooms(context.Background(), req)

	assert.Equal(t, "Success", resp.Status)
	assert.Equal(t, 1, len(resp.Data.(dtos.RoomListData).Rooms))
}

func TestGetAvailableRooms_NoRoomsFound(t *testing.T) {
	rRepo := new(repoMock.RoomRepositoryMock)
	pagination := &dtos.PaginationMeta{Page: 1, Limit: 10, TotalRecords: 0, TotalPages: 0}

	rRepo.On("GetRooms", mock.Anything,
		mock.AnythingOfType("time.Time"),
		mock.AnythingOfType("time.Time"),
		"", 0, 1, 10,
	).Return([]models.Room{}, int64(0), pagination, nil).Once()

	svc := services.NewRoomService(rRepo, newTestLogger())
	resp := svc.GetAvailableRooms(context.Background(), dtos.GetAvailableRoomsRequest{Page: 1, Limit: 10})

	assert.Equal(t, "Error", resp.Status)
	assert.Equal(t, "No rooms found for the given filters", resp.Message)
}

func TestGetAvailableRooms_DBError(t *testing.T) {
	rRepo := new(repoMock.RoomRepositoryMock)

	rRepo.On("GetRooms", mock.Anything,
		mock.AnythingOfType("time.Time"),
		mock.AnythingOfType("time.Time"),
		"", 0, 1, 10,
	).Return(nil, int64(0), nil, &dtos.Error{Code: errorcodes.HMS_DB_001, Message: "db error"}).Once()

	svc := services.NewRoomService(rRepo, newTestLogger())
	resp := svc.GetAvailableRooms(context.Background(), dtos.GetAvailableRoomsRequest{Page: 1, Limit: 10})

	assert.Equal(t, "Error", resp.Status)
	assert.Equal(t, "Failed to fetch rooms", resp.Message)
}
