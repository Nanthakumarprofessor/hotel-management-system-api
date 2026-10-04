package services

import (
	"context"
	"net/http"
	"time"

	"hotel-updated/internal/dtos"
	"hotel-updated/internal/errorcodes"
	"hotel-updated/internal/loggers"
	"hotel-updated/internal/repository"
	"hotel-updated/internal/utils"

	"go.uber.org/zap"
)

// RoomServiceInterface defines room service operations
type RoomServiceInterface interface {
	GetAvailableRooms(ctx context.Context, req dtos.GetAvailableRoomsRequest) *dtos.APIResponse
	GetRoomCategories(ctx context.Context) *dtos.APIResponse
}

// RoomService implements RoomServiceInterface
type RoomService struct {
	roomRepo repository.RoomRepositoryInterface
	logger   *loggers.Logger
}

// NewRoomService creates a new room service
func NewRoomService(roomRepo repository.RoomRepositoryInterface, logger *loggers.Logger) RoomServiceInterface {
	return &RoomService{
		roomRepo: roomRepo,
		logger:   logger,
	}
}

// GetAvailableRooms returns rooms based on filters; checkin/checkout are optional
func (s *RoomService) GetAvailableRooms(ctx context.Context, req dtos.GetAvailableRoomsRequest) *dtos.APIResponse {
	s.logger.Info("GetAvailableRooms: fetching rooms",
		zap.Int("page", req.Page),
		zap.Int("limit", req.Limit),
	)

	var checkIn, checkOut time.Time
	if req.CheckIn != "" {
		checkIn, _ = time.Parse(time.RFC3339, req.CheckIn)
	}
	if req.CheckOut != "" {
		checkOut, _ = time.Parse(time.RFC3339, req.CheckOut)
	}

	rooms, total, pagination, err := s.roomRepo.GetRooms(ctx, checkIn, checkOut, req.RoomCategoryUUID, req.Capacity, req.Page, req.Limit)
	if err != nil {
		return utils.ErrorResponse("Failed to fetch rooms", utils.SingleError("", err.Message, err.Code))
	}

	if len(rooms) == 0 {
		return utils.ErrorResponse("No rooms found for the given filters", utils.SingleError("", "No rooms match the selected filters", errorcodes.HMS_REC_404))
	}

	s.logger.Info("Rooms fetched successfully", zap.Int64("total_count", total))

	// Direct struct mapping using make + index — no intermediate loop variable 
	items := make([]dtos.RoomAvailabilityItem, len(rooms))
	for i := range rooms {
		items[i] = dtos.RoomAvailabilityItem{
			RoomUUID: rooms[i].RoomUUID,
			RoomNo:   rooms[i].RoomNo,
			Price:    rooms[i].Price,
			IsActive: rooms[i].IsActive,
		}
	}

	data := dtos.RoomListData{
		Rooms:      items,
		Pagination: *pagination,
	}

	return utils.SuccessResponse("Rooms retrieved successfully", http.StatusOK, data)
}

// GetRoomCategories returns all active room categories
func (s *RoomService) GetRoomCategories(ctx context.Context) *dtos.APIResponse {
	s.logger.Info("GetRoomCategories: fetching categories")

	categories, err := s.roomRepo.GetAllCategories(ctx)
	if err != nil {
		return utils.ErrorResponse("Failed to fetch room categories", utils.SingleError("", err.Message, err.Code))
	}

	if len(categories) == 0 {
		return utils.ErrorResponse("No room categories found", utils.SingleError("", "No active room categories exist in the system", errorcodes.HMS_REC_404))
	}

	s.logger.Info("Room categories fetched", zap.Int("count", len(categories)))

	items := make([]dtos.RoomCategoryItem, len(categories))
	for i := range categories {
		items[i] = dtos.RoomCategoryItem{
			RoomCategoryUUID: categories[i].RoomCategoryUUID,
			RoomCategoryName: categories[i].RoomCategoryName,
		}
	}

	data := dtos.RoomCategoryListData{
		Categories: items,
		Total:      len(items),
	}

	return utils.SuccessResponse("Room categories retrieved successfully", http.StatusOK, data)
}
