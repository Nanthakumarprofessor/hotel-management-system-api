package utils

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"

	"hotel-updated/internal/dtos"
	"hotel-updated/internal/errorcodes"
)

var validate = validator.New()

// WriteResponse writes JSON response to http.ResponseWriter
func WriteResponse(w http.ResponseWriter, code int, data interface{}) {
	if w.Header().Get("Content-Type") == "" {
		w.Header().Set("Content-Type", "application/json")
	}
	w.WriteHeader(code)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// ValidateStruct validates a struct using playground validator tags
func ValidateStruct(s interface{}) []dtos.Error {
	var validationErrors []dtos.Error

	err := validate.Struct(s)
	if err != nil {
		if ve, ok := err.(validator.ValidationErrors); ok {
			for _, e := range ve {
				validationErrors = append(validationErrors, dtos.Error{
					Field:   e.Field(),
					Message: fmt.Sprintf("Field '%s' failed validation on '%s' tag", e.Field(), e.Tag()),
					Code:    errorcodes.HMS_VAL_001,
				})
			}
		}
	}

	return validationErrors
}

// ValidateRoomQueryRequest validates room query params using playground validator where possible.
// It enforces: limit requires page, checkin/checkout are optional but must be valid RFC3339 if provided,
// and numeric fields must be positive integers within allowed ranges.
func ValidateRoomQueryRequest(req *dtos.GetAvailableRoomsRequest) []dtos.Error {
	var errs []dtos.Error

	// Validate checkin/checkout: optional, but must be valid RFC3339 if provided
	if req.CheckIn != "" {
		if err := validate.Var(req.CheckIn, "datetime=2006-01-02T15:04:05Z07:00"); err != nil {
			errs = append(errs, dtos.Error{
				Field:   "check_in",
				Message: "check_in must be a valid RFC3339 datetime (e.g., 2025-06-01T14:00:00Z)",
				Code:    errorcodes.HMS_VAL_001,
			})
		}
	}

	if req.CheckOut != "" {
		if err := validate.Var(req.CheckOut, "datetime=2006-01-02T15:04:05Z07:00"); err != nil {
			errs = append(errs, dtos.Error{
				Field:   "check_out",
				Message: "check_out must be a valid RFC3339 datetime (e.g., 2025-06-01T14:00:00Z)",
				Code:    errorcodes.HMS_VAL_001,
			})
		}
	}

	// Both provided: validate date range
	if req.CheckIn != "" && req.CheckOut != "" && len(errs) == 0 {
		checkIn, _ := time.Parse(time.RFC3339, req.CheckIn)
		checkOut, _ := time.Parse(time.RFC3339, req.CheckOut)
		if !checkOut.After(checkIn) {
			errs = append(errs, dtos.Error{
				Field:   "check_out",
				Message: "check_out must be after check_in",
				Code:    errorcodes.HMS_VAL_001,
			})
		}
	}

	// room_category_uuid: optional, but must be valid UUID if provided
	if req.RoomCategoryUUID != "" {
		if err := validate.Var(req.RoomCategoryUUID, "uuid4"); err != nil {
			errs = append(errs, dtos.Error{
				Field:   "room_category_uuid",
				Message: "room_category_uuid must be a valid UUID",
				Code:    errorcodes.HMS_VAL_001,
			})
		}
	}

	// capacity: optional positive integer
	if req.CapacityStr != "" {
		capInt, err := strconv.Atoi(req.CapacityStr)
		if err != nil || capInt < 1 {
			errs = append(errs, dtos.Error{
				Field:   "capacity",
				Message: "capacity must be a positive integer",
				Code:    errorcodes.HMS_VAL_001,
			})
		} else {
			req.Capacity = capInt
		}
	}

	// limit without page is not allowed
	if req.LimitStr != "" && req.PageStr == "" {
		errs = append(errs, dtos.Error{
			Field:   "limit",
			Message: "limit cannot be used without page",
			Code:    errorcodes.HMS_VAL_001,
		})
	}

	// page: optional positive integer
	req.Page = 1
	if req.PageStr != "" {
		pageInt, err := strconv.Atoi(req.PageStr)
		if err != nil || pageInt < 1 {
			errs = append(errs, dtos.Error{
				Field:   "page",
				Message: "page must be a positive integer greater than 0",
				Code:    errorcodes.HMS_VAL_001,
			})
		} else {
			req.Page = pageInt
		}
	}

	// limit: optional, 1–100 (only parsed if page is also given)
	req.Limit = 10
	if req.LimitStr != "" {
		limitInt, err := strconv.Atoi(req.LimitStr)
		if err != nil || limitInt < 1 || limitInt > 100 {
			errs = append(errs, dtos.Error{
				Field:   "limit",
				Message: "limit must be between 1 and 100",
				Code:    errorcodes.HMS_VAL_001,
			})
		} else {
			req.Limit = limitInt
		}
	}

	return errs
}

// MapErrorCode maps internal error code to HTTP status code
func MapErrorCode(apiResponse *dtos.APIResponse) *dtos.APIResponse {
	if len(apiResponse.Errors) == 0 {
		if apiResponse.Code == 0 {
			apiResponse.Code = http.StatusOK
		}
		return apiResponse
	}

	if code, exists := errorcodes.ErrorCodeToHTTPStatus[apiResponse.Errors[0].Code]; exists {
		apiResponse.Code = code
	} else {
		apiResponse.Code = http.StatusInternalServerError
	}

	return apiResponse
}

// ErrorResponse creates an error API response
func ErrorResponse(msg string, errs []dtos.Error) *dtos.APIResponse {
	return &dtos.APIResponse{
		Status:  "Error",
		Message: msg,
		Errors:  errs,
	}
}

// SuccessResponse creates a success API response
func SuccessResponse(msg string, code int, data interface{}) *dtos.APIResponse {
	return &dtos.APIResponse{
		Status:  "Success",
		Code:    code,
		Message: msg,
		Data:    data,
	}
}

// ParseDate parses a date string in RFC3339 format
func ParseDate(fieldName, value string) (time.Time, *dtos.Error) {
	parsedTime, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return time.Time{}, &dtos.Error{
			Field:   fieldName,
			Message: fmt.Sprintf("%s must be a valid RFC3339 datetime (e.g., 2025-06-01T14:00:00Z)", fieldName),
			Code:    errorcodes.HMS_VAL_001,
		}
	}
	return parsedTime, nil
}

// ValidateDateRange validates that check_out is after check_in
func ValidateDateRange(checkIn, checkOut time.Time) *dtos.Error {
	if !checkOut.After(checkIn) {
		return &dtos.Error{
			Field:   "check_out",
			Message: "check_out must be after check_in",
			Code:    errorcodes.HMS_VAL_001,
		}
	}
	return nil
}

// ValidateCheckOutExtension validates that new check_out is after current check_out
func ValidateCheckOutExtension(currentCheckOut, newCheckOut time.Time) *dtos.Error {
	if !newCheckOut.After(currentCheckOut) {
		return &dtos.Error{
			Field:   "check_out",
			Message: "new check_out must be after the current check_out",
			Code:    errorcodes.HMS_VAL_001,
		}
	}
	return nil
}

// SingleError creates a single error slice
func SingleError(field, message, code string) []dtos.Error {
	return []dtos.Error{
		{
			Field:   field,
			Message: message,
			Code:    code,
		},
	}
}

// GenerateUUID generates a new UUID string
func GenerateUUID() string {
	return uuid.New().String()
}

// FormatTime formats time to RFC3339 string
func FormatTime(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format(time.RFC3339)
}

// ParseUUID validates and parses a UUID string
func ParseUUID(fieldName, uuidStr string) (uuid.UUID, *dtos.Error) {
	parsedUUID, err := uuid.Parse(uuidStr)
	if err != nil {
		return uuid.Nil, &dtos.Error{
			Field:   fieldName,
			Message: fmt.Sprintf("%s must be a valid UUID", fieldName),
			Code:    errorcodes.HMS_VAL_001,
		}
	}
	return parsedUUID, nil
}
