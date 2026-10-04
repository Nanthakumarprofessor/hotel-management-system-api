package handlers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"hotel-updated/internal/dtos"
	"hotel-updated/internal/errorcodes"
	"hotel-updated/internal/handlers"
	"hotel-updated/internal/loggers"
	svcMock "hotel-updated/tests/unit/services/mock"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func newHandlerLogger() *loggers.Logger {
	tmpDir, _ := os.MkdirTemp("", "handler-test-logs-*")
	return loggers.NewLogger(loggers.LogConfig{Level: "info", LogDir: tmpDir, FileName: "handler_test.log", ServiceName: "test"})
}

func successBookingResp() *dtos.APIResponse {
	return &dtos.APIResponse{Status: "Success", Code: http.StatusCreated, Message: "Booking created successfully"}
}

func errorResp(code int, msg, errCode string) *dtos.APIResponse {
	return &dtos.APIResponse{Status: "Error", Code: code, Message: msg, Errors: []dtos.Error{{Code: errCode, Message: msg}}}
}

// ─── CreateBooking ─────────────────────────────────────────────────────────────

func TestCreateBookingHandler_Success(t *testing.T) {
	svc := new(svcMock.BookingServiceMock)
	h := handlers.NewBookingHandler(svc, newHandlerLogger())

	checkIn := time.Now().Add(24 * time.Hour).UTC().Format(time.RFC3339)
	checkOut := time.Now().Add(72 * time.Hour).UTC().Format(time.RFC3339)
	roomUUID := "550e8400-e29b-41d4-a716-446655440000"

	body, _ := json.Marshal(map[string]interface{}{
		"room_uuid": roomUUID, "check_in": checkIn, "check_out": checkOut,
		"guest": map[string]interface{}{"guest_name": "John", "age": 30, "address": "Addr", "phone_number": "+1234567890"},
	})

	svc.On("CreateBooking", mock.Anything, mock.AnythingOfType("*dtos.CreateBookingRequest"),
		mock.AnythingOfType("time.Time"), mock.AnythingOfType("time.Time"), roomUUID).
		Return(successBookingResp()).Once()

	req := httptest.NewRequest(http.MethodPost, "/api/v1/bookings", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.CreateBooking(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code)
	svc.AssertExpectations(t)
}

func TestCreateBookingHandler_InvalidJSON(t *testing.T) {
	h := handlers.NewBookingHandler(new(svcMock.BookingServiceMock), newHandlerLogger())
	req := httptest.NewRequest(http.MethodPost, "/api/v1/bookings", bytes.NewBufferString(`{bad json`))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.CreateBooking(rr, req)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestCreateBookingHandler_MissingRoomUUID(t *testing.T) {
	h := handlers.NewBookingHandler(new(svcMock.BookingServiceMock), newHandlerLogger())
	checkIn := time.Now().Add(24 * time.Hour).UTC().Format(time.RFC3339)
	checkOut := time.Now().Add(72 * time.Hour).UTC().Format(time.RFC3339)
	body, _ := json.Marshal(map[string]interface{}{
		"check_in": checkIn, "check_out": checkOut,
		"guest": map[string]interface{}{"guest_name": "John", "age": 30, "address": "Addr", "phone_number": "+1"},
	})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/bookings", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.CreateBooking(rr, req)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestCreateBookingHandler_InvalidRoomUUID(t *testing.T) {
	h := handlers.NewBookingHandler(new(svcMock.BookingServiceMock), newHandlerLogger())
	checkIn := time.Now().Add(24 * time.Hour).UTC().Format(time.RFC3339)
	checkOut := time.Now().Add(72 * time.Hour).UTC().Format(time.RFC3339)
	body, _ := json.Marshal(map[string]interface{}{
		"room_uuid": "not-a-uuid", "check_in": checkIn, "check_out": checkOut,
		"guest": map[string]interface{}{"guest_name": "John", "age": 30, "address": "Addr", "phone_number": "+1"},
	})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/bookings", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.CreateBooking(rr, req)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestCreateBookingHandler_InvalidCheckInFormat(t *testing.T) {
	h := handlers.NewBookingHandler(new(svcMock.BookingServiceMock), newHandlerLogger())
	checkOut := time.Now().Add(72 * time.Hour).UTC().Format(time.RFC3339)
	body, _ := json.Marshal(map[string]interface{}{
		"room_uuid": "550e8400-e29b-41d4-a716-446655440000",
		"check_in": "2025-06-01", "check_out": checkOut,
		"guest": map[string]interface{}{"guest_name": "John", "age": 30, "address": "Addr", "phone_number": "+1"},
	})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/bookings", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.CreateBooking(rr, req)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestCreateBookingHandler_InvalidCheckOutFormat(t *testing.T) {
	h := handlers.NewBookingHandler(new(svcMock.BookingServiceMock), newHandlerLogger())
	checkIn := time.Now().Add(24 * time.Hour).UTC().Format(time.RFC3339)
	body, _ := json.Marshal(map[string]interface{}{
		"room_uuid": "550e8400-e29b-41d4-a716-446655440000",
		"check_in": checkIn, "check_out": "bad-date",
		"guest": map[string]interface{}{"guest_name": "John", "age": 30, "address": "Addr", "phone_number": "+1"},
	})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/bookings", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.CreateBooking(rr, req)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestCreateBookingHandler_CheckOutBeforeCheckIn(t *testing.T) {
	h := handlers.NewBookingHandler(new(svcMock.BookingServiceMock), newHandlerLogger())
	checkIn := time.Now().Add(72 * time.Hour).UTC().Format(time.RFC3339)
	checkOut := time.Now().Add(24 * time.Hour).UTC().Format(time.RFC3339)
	body, _ := json.Marshal(map[string]interface{}{
		"room_uuid": "550e8400-e29b-41d4-a716-446655440000",
		"check_in": checkIn, "check_out": checkOut,
		"guest": map[string]interface{}{"guest_name": "John", "age": 30, "address": "Addr", "phone_number": "+1234567890"},
	})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/bookings", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.CreateBooking(rr, req)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestCreateBookingHandler_MissingGuestName(t *testing.T) {
	h := handlers.NewBookingHandler(new(svcMock.BookingServiceMock), newHandlerLogger())
	checkIn := time.Now().Add(24 * time.Hour).UTC().Format(time.RFC3339)
	checkOut := time.Now().Add(72 * time.Hour).UTC().Format(time.RFC3339)
	body, _ := json.Marshal(map[string]interface{}{
		"room_uuid": "550e8400-e29b-41d4-a716-446655440000",
		"check_in": checkIn, "check_out": checkOut,
		"guest": map[string]interface{}{"guest_name": "", "age": 30, "address": "Addr", "phone_number": "+1"},
	})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/bookings", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.CreateBooking(rr, req)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestCreateBookingHandler_ServiceError(t *testing.T) {
	svc := new(svcMock.BookingServiceMock)
	h := handlers.NewBookingHandler(svc, newHandlerLogger())
	checkIn := time.Now().Add(24 * time.Hour).UTC().Format(time.RFC3339)
	checkOut := time.Now().Add(72 * time.Hour).UTC().Format(time.RFC3339)
	roomUUID := "550e8400-e29b-41d4-a716-446655440000"
	body, _ := json.Marshal(map[string]interface{}{
		"room_uuid": roomUUID, "check_in": checkIn, "check_out": checkOut,
		"guest": map[string]interface{}{"guest_name": "John", "age": 30, "address": "Addr", "phone_number": "+1234567890"},
	})
	svc.On("CreateBooking", mock.Anything, mock.Anything, mock.Anything, mock.Anything, roomUUID).
		Return(errorResp(http.StatusConflict, "Room is not available", errorcodes.HMS_CONFLICT_409)).Once()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/bookings", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.CreateBooking(rr, req)
	assert.Equal(t, http.StatusConflict, rr.Code)
	svc.AssertExpectations(t)
}

// ─── ExtendBooking ─────────────────────────────────────────────────────────────

func TestExtendBookingHandler_Success(t *testing.T) {
	svc := new(svcMock.BookingServiceMock)
	h := handlers.NewBookingHandler(svc, newHandlerLogger())
	bookingUUID := "550e8400-e29b-41d4-a716-446655440001"
	checkOut := time.Now().Add(96 * time.Hour).UTC().Format(time.RFC3339)
	body, _ := json.Marshal(map[string]interface{}{"booking_uuid": bookingUUID, "check_out": checkOut})
	svc.On("ExtendBooking", mock.Anything, bookingUUID, mock.AnythingOfType("time.Time")).
		Return(&dtos.APIResponse{Status: "Success", Code: http.StatusOK, Message: "Booking extended successfully"}).Once()
	req := httptest.NewRequest(http.MethodPut, "/api/v1/bookings/extend", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.ExtendBooking(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)
	svc.AssertExpectations(t)
}

func TestExtendBookingHandler_InvalidJSON(t *testing.T) {
	h := handlers.NewBookingHandler(new(svcMock.BookingServiceMock), newHandlerLogger())
	req := httptest.NewRequest(http.MethodPut, "/api/v1/bookings/extend", bytes.NewBufferString(`{bad`))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.ExtendBooking(rr, req)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestExtendBookingHandler_MissingBookingUUID(t *testing.T) {
	h := handlers.NewBookingHandler(new(svcMock.BookingServiceMock), newHandlerLogger())
	checkOut := time.Now().Add(96 * time.Hour).UTC().Format(time.RFC3339)
	body, _ := json.Marshal(map[string]interface{}{"check_out": checkOut})
	req := httptest.NewRequest(http.MethodPut, "/api/v1/bookings/extend", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.ExtendBooking(rr, req)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestExtendBookingHandler_InvalidBookingUUID(t *testing.T) {
	h := handlers.NewBookingHandler(new(svcMock.BookingServiceMock), newHandlerLogger())
	checkOut := time.Now().Add(96 * time.Hour).UTC().Format(time.RFC3339)
	body, _ := json.Marshal(map[string]interface{}{"booking_uuid": "not-a-uuid", "check_out": checkOut})
	req := httptest.NewRequest(http.MethodPut, "/api/v1/bookings/extend", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.ExtendBooking(rr, req)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestExtendBookingHandler_InvalidCheckOutFormat(t *testing.T) {
	h := handlers.NewBookingHandler(new(svcMock.BookingServiceMock), newHandlerLogger())
	body, _ := json.Marshal(map[string]interface{}{"booking_uuid": "550e8400-e29b-41d4-a716-446655440001", "check_out": "bad-date"})
	req := httptest.NewRequest(http.MethodPut, "/api/v1/bookings/extend", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.ExtendBooking(rr, req)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestExtendBookingHandler_ServiceError(t *testing.T) {
	svc := new(svcMock.BookingServiceMock)
	h := handlers.NewBookingHandler(svc, newHandlerLogger())
	bookingUUID := "550e8400-e29b-41d4-a716-446655440001"
	checkOut := time.Now().Add(96 * time.Hour).UTC().Format(time.RFC3339)
	body, _ := json.Marshal(map[string]interface{}{"booking_uuid": bookingUUID, "check_out": checkOut})
	svc.On("ExtendBooking", mock.Anything, bookingUUID, mock.AnythingOfType("time.Time")).
		Return(errorResp(http.StatusNotFound, "Booking not found", errorcodes.HMS_REC_404)).Once()
	req := httptest.NewRequest(http.MethodPut, "/api/v1/bookings/extend", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.ExtendBooking(rr, req)
	assert.Equal(t, http.StatusNotFound, rr.Code)
	svc.AssertExpectations(t)
}

// ─── CancelBooking ─────────────────────────────────────────────────────────────

func TestCancelBookingHandler_Success(t *testing.T) {
	svc := new(svcMock.BookingServiceMock)
	h := handlers.NewBookingHandler(svc, newHandlerLogger())
	bookingUUID := "550e8400-e29b-41d4-a716-446655440002"
	body, _ := json.Marshal(map[string]interface{}{"booking_uuid": bookingUUID})
	svc.On("CancelBooking", mock.Anything, bookingUUID).
		Return(&dtos.APIResponse{Status: "Success", Code: http.StatusOK, Message: "Booking cancelled successfully"}).Once()
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/bookings/cancel", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.CancelBooking(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)
	svc.AssertExpectations(t)
}

func TestCancelBookingHandler_InvalidJSON(t *testing.T) {
	h := handlers.NewBookingHandler(new(svcMock.BookingServiceMock), newHandlerLogger())
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/bookings/cancel", bytes.NewBufferString(`{bad`))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.CancelBooking(rr, req)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestCancelBookingHandler_MissingBookingUUID(t *testing.T) {
	h := handlers.NewBookingHandler(new(svcMock.BookingServiceMock), newHandlerLogger())
	body, _ := json.Marshal(map[string]interface{}{"booking_uuid": ""})
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/bookings/cancel", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.CancelBooking(rr, req)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestCancelBookingHandler_InvalidBookingUUID(t *testing.T) {
	h := handlers.NewBookingHandler(new(svcMock.BookingServiceMock), newHandlerLogger())
	body, _ := json.Marshal(map[string]interface{}{"booking_uuid": "not-a-uuid"})
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/bookings/cancel", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.CancelBooking(rr, req)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestCancelBookingHandler_ServiceError(t *testing.T) {
	svc := new(svcMock.BookingServiceMock)
	h := handlers.NewBookingHandler(svc, newHandlerLogger())
	bookingUUID := "550e8400-e29b-41d4-a716-446655440002"
	body, _ := json.Marshal(map[string]interface{}{"booking_uuid": bookingUUID})
	svc.On("CancelBooking", mock.Anything, bookingUUID).
		Return(errorResp(http.StatusConflict, "Booking is already cancelled", errorcodes.HMS_CONFLICT_409)).Once()
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/bookings/cancel", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.CancelBooking(rr, req)
	assert.Equal(t, http.StatusConflict, rr.Code)
	svc.AssertExpectations(t)
}

// ─── RestoreBooking ────────────────────────────────────────────────────────────

func TestRestoreBookingHandler_Success(t *testing.T) {
	svc := new(svcMock.BookingServiceMock)
	h := handlers.NewBookingHandler(svc, newHandlerLogger())
	bookingUUID := "550e8400-e29b-41d4-a716-446655440003"
	body, _ := json.Marshal(map[string]interface{}{"booking_uuid": bookingUUID})
	svc.On("RestoreBooking", mock.Anything, bookingUUID).
		Return(&dtos.APIResponse{Status: "Success", Code: http.StatusOK, Message: "Booking restored successfully"}).Once()
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/bookings/restore", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.RestoreBooking(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)
	svc.AssertExpectations(t)
}

func TestRestoreBookingHandler_InvalidJSON(t *testing.T) {
	h := handlers.NewBookingHandler(new(svcMock.BookingServiceMock), newHandlerLogger())
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/bookings/restore", bytes.NewBufferString(`{bad`))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.RestoreBooking(rr, req)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestRestoreBookingHandler_MissingBookingUUID(t *testing.T) {
	h := handlers.NewBookingHandler(new(svcMock.BookingServiceMock), newHandlerLogger())
	body, _ := json.Marshal(map[string]interface{}{"booking_uuid": ""})
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/bookings/restore", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.RestoreBooking(rr, req)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestRestoreBookingHandler_InvalidBookingUUID(t *testing.T) {
	h := handlers.NewBookingHandler(new(svcMock.BookingServiceMock), newHandlerLogger())
	body, _ := json.Marshal(map[string]interface{}{"booking_uuid": "not-a-uuid"})
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/bookings/restore", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.RestoreBooking(rr, req)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestRestoreBookingHandler_ServiceError(t *testing.T) {
	svc := new(svcMock.BookingServiceMock)
	h := handlers.NewBookingHandler(svc, newHandlerLogger())
	bookingUUID := "550e8400-e29b-41d4-a716-446655440003"
	body, _ := json.Marshal(map[string]interface{}{"booking_uuid": bookingUUID})
	svc.On("RestoreBooking", mock.Anything, bookingUUID).
		Return(errorResp(http.StatusConflict, "Booking is already active", errorcodes.HMS_CONFLICT_409)).Once()
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/bookings/restore", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.RestoreBooking(rr, req)
	assert.Equal(t, http.StatusConflict, rr.Code)
	svc.AssertExpectations(t)
}
