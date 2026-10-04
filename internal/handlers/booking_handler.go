package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"hotel-updated/internal/dtos"
	"hotel-updated/internal/errorcodes"
	"hotel-updated/internal/loggers"
	"hotel-updated/internal/services"
	"hotel-updated/internal/utils"
)

// BookingHandler handles booking-related HTTP requests
type BookingHandler struct {
	bookingService services.BookingServiceInterface
	logger         *loggers.Logger
}

// NewBookingHandler creates a new booking handler
func NewBookingHandler(bookingService services.BookingServiceInterface, logger *loggers.Logger) *BookingHandler {
	return &BookingHandler{
		bookingService: bookingService,
		logger:         logger,
	}
}

func writeValidationError(w http.ResponseWriter, msg string, errs []dtos.Error) {
	response := utils.ErrorResponse(msg, errs)
	response = utils.MapErrorCode(response)
	response.RequestID = utils.GenerateUUID()
	response.Timestamp = time.Now().UTC().Format(time.RFC3339)
	utils.WriteResponse(w, response.Code, response)
}

// CreateBooking handles POST /api/v1/bookings
func (h *BookingHandler) CreateBooking(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("CreateBooking: handling create booking request")

	var req dtos.CreateBookingRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeValidationError(w, "Invalid request format", utils.SingleError("body", "Invalid JSON format", errorcodes.HMS_REQ_001))
		return
	}

	// Validate top-level fields 
	if errs := utils.ValidateStruct(&req); len(errs) > 0 {
		writeValidationError(w, "Validation failed for the request", errs)
		return
	}

	// Validate nested guest fields
	if errs := utils.ValidateStruct(&req.Guest); len(errs) > 0 {
		writeValidationError(w, "Validation failed for guest fields", errs)
		return
	}

	// Parse already-validated dates
	checkIn, _ := utils.ParseDate("check_in", req.CheckIn)
	checkOut, _ := utils.ParseDate("check_out", req.CheckOut)

	if err := utils.ValidateDateRange(checkIn, checkOut); err != nil {
		writeValidationError(w, "Validation failed", utils.SingleError(err.Field, err.Message, err.Code))
		return
	}

	apiResponse := h.bookingService.CreateBooking(r.Context(), &req, checkIn, checkOut, req.RoomUUID)
	response := utils.MapErrorCode(apiResponse)
	response.RequestID = utils.GenerateUUID()
	response.Timestamp = time.Now().UTC().Format(time.RFC3339)
	utils.WriteResponse(w, response.Code, response)
}

// ExtendBooking handles PUT /api/v1/bookings/extend
func (h *BookingHandler) ExtendBooking(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("ExtendBooking: handling extend booking request")

	var req dtos.ExtendBookingRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeValidationError(w, "Invalid request format", utils.SingleError("body", "Invalid JSON format", errorcodes.HMS_REQ_001))
		return
	}

	if errs := utils.ValidateStruct(&req); len(errs) > 0 {
		writeValidationError(w, "Validation failed for the request", errs)
		return
	}

	newCheckOut, _ := utils.ParseDate("check_out", req.CheckOut)

	apiResponse := h.bookingService.ExtendBooking(r.Context(), req.BookingUUID, newCheckOut)
	response := utils.MapErrorCode(apiResponse)
	response.RequestID = utils.GenerateUUID()
	response.Timestamp = time.Now().UTC().Format(time.RFC3339)
	utils.WriteResponse(w, response.Code, response)
}

// CancelBooking handles PATCH /api/v1/bookings/cancel
func (h *BookingHandler) CancelBooking(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("CancelBooking: handling cancel booking request")

	var req dtos.CancelBookingRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeValidationError(w, "Invalid request format", utils.SingleError("body", "Invalid JSON format", errorcodes.HMS_REQ_001))
		return
	}

	if errs := utils.ValidateStruct(&req); len(errs) > 0 {
		writeValidationError(w, "Validation failed for the request", errs)
		return
	}

	apiResponse := h.bookingService.CancelBooking(r.Context(), req.BookingUUID)
	response := utils.MapErrorCode(apiResponse)
	response.RequestID = utils.GenerateUUID()
	response.Timestamp = time.Now().UTC().Format(time.RFC3339)
	utils.WriteResponse(w, response.Code, response)
}

// RestoreBooking handles PATCH /api/v1/bookings/restore
func (h *BookingHandler) RestoreBooking(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("RestoreBooking: handling restore booking request")

	var req dtos.RestoreBookingRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeValidationError(w, "Invalid request format", utils.SingleError("body", "Invalid JSON format", errorcodes.HMS_REQ_001))
		return
	}

	if errs := utils.ValidateStruct(&req); len(errs) > 0 {
		writeValidationError(w, "Validation failed for the request", errs)
		return
	}

	apiResponse := h.bookingService.RestoreBooking(r.Context(), req.BookingUUID)
	response := utils.MapErrorCode(apiResponse)
	response.RequestID = utils.GenerateUUID()
	response.Timestamp = time.Now().UTC().Format(time.RFC3339)
	utils.WriteResponse(w, response.Code, response)
}
