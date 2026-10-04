package handlers_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"hotel-updated/internal/dtos"
	"hotel-updated/internal/errorcodes"
	"hotel-updated/internal/handlers"
	svcMock "hotel-updated/tests/unit/services/mock"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// ─── GetRoomCategories ─────────────────────────────────────────────────────────

func TestGetRoomCategoriesHandler_Success(t *testing.T) {
	svc := new(svcMock.RoomServiceMock)
	h := handlers.NewRoomHandler(svc, newHandlerLogger())
	svc.On("GetRoomCategories", mock.Anything).Return(&dtos.APIResponse{
		Status: "Success", Code: http.StatusOK, Message: "Room categories retrieved successfully",
		Data: dtos.RoomCategoryListData{Categories: []dtos.RoomCategoryItem{{RoomCategoryUUID: "cat-1", RoomCategoryName: "Standard"}}, Total: 1},
	}).Once()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/rooms/categories", nil)
	rr := httptest.NewRecorder()
	h.GetRoomCategories(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	var resp dtos.APIResponse
	assert.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	assert.Equal(t, "Success", resp.Status)
	svc.AssertExpectations(t)
}

func TestGetRoomCategoriesHandler_NotFound(t *testing.T) {
	svc := new(svcMock.RoomServiceMock)
	h := handlers.NewRoomHandler(svc, newHandlerLogger())
	svc.On("GetRoomCategories", mock.Anything).Return(&dtos.APIResponse{
		Status: "Error", Code: http.StatusNotFound, Message: "No room categories found",
		Errors: []dtos.Error{{Code: errorcodes.HMS_REC_404, Message: "No active room categories exist"}},
	}).Once()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/rooms/categories", nil)
	rr := httptest.NewRecorder()
	h.GetRoomCategories(rr, req)
	assert.Equal(t, http.StatusNotFound, rr.Code)
	svc.AssertExpectations(t)
}

func TestGetRoomCategoriesHandler_DBError(t *testing.T) {
	svc := new(svcMock.RoomServiceMock)
	h := handlers.NewRoomHandler(svc, newHandlerLogger())
	svc.On("GetRoomCategories", mock.Anything).Return(&dtos.APIResponse{
		Status: "Error", Code: http.StatusInternalServerError, Message: "Failed to fetch room categories",
		Errors: []dtos.Error{{Code: errorcodes.HMS_DB_001, Message: "db error"}},
	}).Once()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/rooms/categories", nil)
	rr := httptest.NewRecorder()
	h.GetRoomCategories(rr, req)
	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	svc.AssertExpectations(t)
}

// ─── GetAvailableRooms ─────────────────────────────────────────────────────────

func validRoomsResp() *dtos.APIResponse {
	return &dtos.APIResponse{
		Status: "Success", Code: http.StatusOK, Message: "Rooms retrieved successfully",
		Data: dtos.RoomListData{
			Rooms:      []dtos.RoomAvailabilityItem{{RoomUUID: "r-1", RoomNo: "101", Price: 100.0, IsActive: true}},
			Pagination: dtos.PaginationMeta{Page: 1, Limit: 10, TotalRecords: 1, TotalPages: 1},
		},
	}
}

func TestGetAvailableRoomsHandler_Success_WithDates(t *testing.T) {
	svc := new(svcMock.RoomServiceMock)
	h := handlers.NewRoomHandler(svc, newHandlerLogger())
	checkIn := time.Now().Add(24 * time.Hour).UTC().Format(time.RFC3339)
	checkOut := time.Now().Add(72 * time.Hour).UTC().Format(time.RFC3339)
	svc.On("GetAvailableRooms", mock.Anything, mock.AnythingOfType("dtos.GetAvailableRoomsRequest")).Return(validRoomsResp()).Once()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/rooms/available?check_in="+checkIn+"&check_out="+checkOut, nil)
	rr := httptest.NewRecorder()
	h.GetAvailableRooms(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)
	svc.AssertExpectations(t)
}

func TestGetAvailableRoomsHandler_Success_NoDates(t *testing.T) {
	svc := new(svcMock.RoomServiceMock)
	h := handlers.NewRoomHandler(svc, newHandlerLogger())
	svc.On("GetAvailableRooms", mock.Anything, mock.AnythingOfType("dtos.GetAvailableRoomsRequest")).Return(validRoomsResp()).Once()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/rooms/available", nil)
	rr := httptest.NewRecorder()
	h.GetAvailableRooms(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)
	svc.AssertExpectations(t)
}

func TestGetAvailableRoomsHandler_Success_WithAllOptionalParams(t *testing.T) {
	svc := new(svcMock.RoomServiceMock)
	h := handlers.NewRoomHandler(svc, newHandlerLogger())
	catUUID := "550e8400-e29b-41d4-a716-446655440000"
	checkIn := time.Now().Add(24 * time.Hour).UTC().Format(time.RFC3339)
	checkOut := time.Now().Add(72 * time.Hour).UTC().Format(time.RFC3339)
	svc.On("GetAvailableRooms", mock.Anything, mock.AnythingOfType("dtos.GetAvailableRoomsRequest")).Return(validRoomsResp()).Once()

	url := "/api/v1/rooms/available?check_in=" + checkIn + "&check_out=" + checkOut + "&room_category_uuid=" + catUUID + "&capacity=2&page=1&limit=5"
	req := httptest.NewRequest(http.MethodGet, url, nil)
	rr := httptest.NewRecorder()
	h.GetAvailableRooms(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)
	svc.AssertExpectations(t)
}

func TestGetAvailableRoomsHandler_InvalidCheckInFormat(t *testing.T) {
	h := handlers.NewRoomHandler(new(svcMock.RoomServiceMock), newHandlerLogger())
	checkOut := time.Now().Add(72 * time.Hour).UTC().Format(time.RFC3339)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/rooms/available?check_in=bad-date&check_out="+checkOut, nil)
	rr := httptest.NewRecorder()
	h.GetAvailableRooms(rr, req)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestGetAvailableRoomsHandler_InvalidCheckOutFormat(t *testing.T) {
	h := handlers.NewRoomHandler(new(svcMock.RoomServiceMock), newHandlerLogger())
	checkIn := time.Now().Add(24 * time.Hour).UTC().Format(time.RFC3339)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/rooms/available?check_in="+checkIn+"&check_out=bad-date", nil)
	rr := httptest.NewRecorder()
	h.GetAvailableRooms(rr, req)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestGetAvailableRoomsHandler_CheckOutBeforeCheckIn(t *testing.T) {
	h := handlers.NewRoomHandler(new(svcMock.RoomServiceMock), newHandlerLogger())
	checkIn := time.Now().Add(72 * time.Hour).UTC().Format(time.RFC3339)
	checkOut := time.Now().Add(24 * time.Hour).UTC().Format(time.RFC3339)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/rooms/available?check_in="+checkIn+"&check_out="+checkOut, nil)
	rr := httptest.NewRecorder()
	h.GetAvailableRooms(rr, req)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestGetAvailableRoomsHandler_InvalidCategoryUUID(t *testing.T) {
	h := handlers.NewRoomHandler(new(svcMock.RoomServiceMock), newHandlerLogger())
	req := httptest.NewRequest(http.MethodGet, "/api/v1/rooms/available?room_category_uuid=not-a-uuid", nil)
	rr := httptest.NewRecorder()
	h.GetAvailableRooms(rr, req)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestGetAvailableRoomsHandler_InvalidCapacity(t *testing.T) {
	h := handlers.NewRoomHandler(new(svcMock.RoomServiceMock), newHandlerLogger())
	req := httptest.NewRequest(http.MethodGet, "/api/v1/rooms/available?capacity=abc", nil)
	rr := httptest.NewRecorder()
	h.GetAvailableRooms(rr, req)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestGetAvailableRoomsHandler_NegativeCapacity(t *testing.T) {
	h := handlers.NewRoomHandler(new(svcMock.RoomServiceMock), newHandlerLogger())
	req := httptest.NewRequest(http.MethodGet, "/api/v1/rooms/available?capacity=-1", nil)
	rr := httptest.NewRecorder()
	h.GetAvailableRooms(rr, req)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestGetAvailableRoomsHandler_LimitWithoutPage(t *testing.T) {
	h := handlers.NewRoomHandler(new(svcMock.RoomServiceMock), newHandlerLogger())
	req := httptest.NewRequest(http.MethodGet, "/api/v1/rooms/available?limit=5", nil)
	rr := httptest.NewRecorder()
	h.GetAvailableRooms(rr, req)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestGetAvailableRoomsHandler_InvalidPage(t *testing.T) {
	h := handlers.NewRoomHandler(new(svcMock.RoomServiceMock), newHandlerLogger())
	req := httptest.NewRequest(http.MethodGet, "/api/v1/rooms/available?page=0", nil)
	rr := httptest.NewRecorder()
	h.GetAvailableRooms(rr, req)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestGetAvailableRoomsHandler_InvalidPageNonNumeric(t *testing.T) {
	h := handlers.NewRoomHandler(new(svcMock.RoomServiceMock), newHandlerLogger())
	req := httptest.NewRequest(http.MethodGet, "/api/v1/rooms/available?page=abc", nil)
	rr := httptest.NewRecorder()
	h.GetAvailableRooms(rr, req)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestGetAvailableRoomsHandler_LimitOver100(t *testing.T) {
	h := handlers.NewRoomHandler(new(svcMock.RoomServiceMock), newHandlerLogger())
	req := httptest.NewRequest(http.MethodGet, "/api/v1/rooms/available?page=1&limit=200", nil)
	rr := httptest.NewRecorder()
	h.GetAvailableRooms(rr, req)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestGetAvailableRoomsHandler_LimitZero(t *testing.T) {
	h := handlers.NewRoomHandler(new(svcMock.RoomServiceMock), newHandlerLogger())
	req := httptest.NewRequest(http.MethodGet, "/api/v1/rooms/available?page=1&limit=0", nil)
	rr := httptest.NewRecorder()
	h.GetAvailableRooms(rr, req)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestGetAvailableRoomsHandler_ServiceError(t *testing.T) {
	svc := new(svcMock.RoomServiceMock)
	h := handlers.NewRoomHandler(svc, newHandlerLogger())
	svc.On("GetAvailableRooms", mock.Anything, mock.AnythingOfType("dtos.GetAvailableRoomsRequest")).Return(&dtos.APIResponse{
		Status: "Error", Code: http.StatusNotFound, Message: "No rooms found for the given filters",
		Errors: []dtos.Error{{Code: errorcodes.HMS_REC_404, Message: "No rooms"}},
	}).Once()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/rooms/available", nil)
	rr := httptest.NewRecorder()
	h.GetAvailableRooms(rr, req)
	assert.Equal(t, http.StatusNotFound, rr.Code)
	svc.AssertExpectations(t)
}
