package handlers

import (
	"net/http"
	"time"

	"hotel-updated/internal/dtos"
	"hotel-updated/internal/loggers"
	"hotel-updated/internal/services"
	"hotel-updated/internal/utils"
)

// RoomHandler handles room-related HTTP requests
type RoomHandler struct {
	roomService services.RoomServiceInterface
	logger      *loggers.Logger
}

// NewRoomHandler creates a new room handler
func NewRoomHandler(roomService services.RoomServiceInterface, logger *loggers.Logger) *RoomHandler {
	return &RoomHandler{
		roomService: roomService,
		logger:      logger,
	}
}

// GetRoomCategories handles GET /api/v1/rooms/categories
func (h *RoomHandler) GetRoomCategories(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("GetRoomCategories: fetching all room categories")

	apiResponse := h.roomService.GetRoomCategories(r.Context())
	response := utils.MapErrorCode(apiResponse)
	response.RequestID = utils.GenerateUUID()
	response.Timestamp = time.Now().UTC().Format(time.RFC3339)
	utils.WriteResponse(w, response.Code, response)
}

// GetAvailableRooms handles GET /api/v1/rooms/available
func (h *RoomHandler) GetAvailableRooms(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("GetAvailableRooms: parsing query parameters")

	query := r.URL.Query()

	req := dtos.GetAvailableRoomsRequest{
		CheckIn:          query.Get("check_in"),
		CheckOut:         query.Get("check_out"),
		RoomCategoryUUID: query.Get("room_category_uuid"),
		CapacityStr:      query.Get("capacity"),
		PageStr:          query.Get("page"),
		LimitStr:         query.Get("limit"),
	}

	errs := utils.ValidateRoomQueryRequest(&req)
	if len(errs) > 0 {
		response := utils.ErrorResponse("Validation failed for the request", errs)
		response = utils.MapErrorCode(response)
		response.RequestID = utils.GenerateUUID()
		response.Timestamp = time.Now().UTC().Format(time.RFC3339)
		utils.WriteResponse(w, response.Code, response)
		return
	}

	apiResponse := h.roomService.GetAvailableRooms(r.Context(), req)
	response := utils.MapErrorCode(apiResponse)
	response.RequestID = utils.GenerateUUID()
	response.Timestamp = time.Now().UTC().Format(time.RFC3339)
	utils.WriteResponse(w, response.Code, response)
}
