package utils_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"hotel-updated/internal/dtos"
	"hotel-updated/internal/errorcodes"
	"hotel-updated/internal/utils"

	"github.com/stretchr/testify/assert"
)

// ─── WriteResponse ─────────────────────────────────────────────────────────────

func TestWriteResponse_200(t *testing.T) {
	w := httptest.NewRecorder()
	utils.WriteResponse(w, http.StatusOK, map[string]string{"k": "v"})
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
}

func TestWriteResponse_400(t *testing.T) {
	w := httptest.NewRecorder()
	utils.WriteResponse(w, http.StatusBadRequest, map[string]string{"error": "bad"})
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestWriteResponse_500(t *testing.T) {
	w := httptest.NewRecorder()
	utils.WriteResponse(w, http.StatusInternalServerError, map[string]string{"error": "internal"})
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestWriteResponse_404(t *testing.T) {
	w := httptest.NewRecorder()
	utils.WriteResponse(w, http.StatusNotFound, map[string]string{"error": "not found"})
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestWriteResponse_JSONEncodeError(t *testing.T) {
	w := httptest.NewRecorder()
	// channels cannot be JSON-encoded
	utils.WriteResponse(w, http.StatusInternalServerError, make(chan int))
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// ─── ValidateStruct ────────────────────────────────────────────────────────────

type sampleStruct struct {
	Name  string `validate:"required"`
	Email string `validate:"required,email"`
}

func TestValidateStruct_Valid(t *testing.T) {
	errs := utils.ValidateStruct(&sampleStruct{Name: "John", Email: "j@x.com"})
	assert.Nil(t, errs)
}

func TestValidateStruct_MissingRequired(t *testing.T) {
	errs := utils.ValidateStruct(&sampleStruct{Email: "j@x.com"})
	assert.NotNil(t, errs)
	assert.Len(t, errs, 1)
	assert.Equal(t, "Name", errs[0].Field)
	assert.Equal(t, errorcodes.HMS_VAL_001, errs[0].Code)
}

func TestValidateStruct_InvalidEmail(t *testing.T) {
	errs := utils.ValidateStruct(&sampleStruct{Name: "John", Email: "not-email"})
	assert.NotNil(t, errs)
	assert.Equal(t, "Email", errs[0].Field)
}

func TestValidateStruct_MultipleErrors(t *testing.T) {
	errs := utils.ValidateStruct(&sampleStruct{})
	assert.True(t, len(errs) >= 2)
}

// ─── ValidateRoomQueryRequest ──────────────────────────────────────────────────

func TestValidateRoomQueryRequest_NoParams(t *testing.T) {
	req := &dtos.GetAvailableRoomsRequest{}
	errs := utils.ValidateRoomQueryRequest(req)
	assert.Empty(t, errs)
	assert.Equal(t, 1, req.Page)
	assert.Equal(t, 10, req.Limit)
}

func TestValidateRoomQueryRequest_ValidDates(t *testing.T) {
	checkIn := time.Now().Add(24 * time.Hour).UTC().Format(time.RFC3339)
	checkOut := time.Now().Add(72 * time.Hour).UTC().Format(time.RFC3339)
	req := &dtos.GetAvailableRoomsRequest{CheckIn: checkIn, CheckOut: checkOut}
	errs := utils.ValidateRoomQueryRequest(req)
	assert.Empty(t, errs)
}

func TestValidateRoomQueryRequest_InvalidCheckIn(t *testing.T) {
	req := &dtos.GetAvailableRoomsRequest{CheckIn: "bad-date"}
	errs := utils.ValidateRoomQueryRequest(req)
	assert.NotEmpty(t, errs)
	assert.Equal(t, "check_in", errs[0].Field)
}

func TestValidateRoomQueryRequest_InvalidCheckOut(t *testing.T) {
	req := &dtos.GetAvailableRoomsRequest{CheckOut: "bad-date"}
	errs := utils.ValidateRoomQueryRequest(req)
	assert.NotEmpty(t, errs)
	assert.Equal(t, "check_out", errs[0].Field)
}

func TestValidateRoomQueryRequest_CheckOutBeforeCheckIn(t *testing.T) {
	checkIn := time.Now().Add(72 * time.Hour).UTC().Format(time.RFC3339)
	checkOut := time.Now().Add(24 * time.Hour).UTC().Format(time.RFC3339)
	req := &dtos.GetAvailableRoomsRequest{CheckIn: checkIn, CheckOut: checkOut}
	errs := utils.ValidateRoomQueryRequest(req)
	assert.NotEmpty(t, errs)
	assert.Equal(t, "check_out", errs[0].Field)
}

func TestValidateRoomQueryRequest_ValidCategoryUUID(t *testing.T) {
	req := &dtos.GetAvailableRoomsRequest{RoomCategoryUUID: "550e8400-e29b-41d4-a716-446655440000"}
	errs := utils.ValidateRoomQueryRequest(req)
	assert.Empty(t, errs)
}

func TestValidateRoomQueryRequest_InvalidCategoryUUID(t *testing.T) {
	req := &dtos.GetAvailableRoomsRequest{RoomCategoryUUID: "not-a-uuid"}
	errs := utils.ValidateRoomQueryRequest(req)
	assert.NotEmpty(t, errs)
	assert.Equal(t, "room_category_uuid", errs[0].Field)
}

func TestValidateRoomQueryRequest_ValidCapacity(t *testing.T) {
	req := &dtos.GetAvailableRoomsRequest{CapacityStr: "2"}
	errs := utils.ValidateRoomQueryRequest(req)
	assert.Empty(t, errs)
	assert.Equal(t, 2, req.Capacity)
}

func TestValidateRoomQueryRequest_InvalidCapacity(t *testing.T) {
	req := &dtos.GetAvailableRoomsRequest{CapacityStr: "abc"}
	errs := utils.ValidateRoomQueryRequest(req)
	assert.NotEmpty(t, errs)
	assert.Equal(t, "capacity", errs[0].Field)
}

func TestValidateRoomQueryRequest_NegativeCapacity(t *testing.T) {
	req := &dtos.GetAvailableRoomsRequest{CapacityStr: "-1"}
	errs := utils.ValidateRoomQueryRequest(req)
	assert.NotEmpty(t, errs)
	assert.Equal(t, "capacity", errs[0].Field)
}

func TestValidateRoomQueryRequest_LimitWithoutPage(t *testing.T) {
	req := &dtos.GetAvailableRoomsRequest{LimitStr: "5"}
	errs := utils.ValidateRoomQueryRequest(req)
	// Should have error for limit without page
	hasLimitErr := false
	for _, e := range errs {
		if e.Field == "limit" {
			hasLimitErr = true
		}
	}
	assert.True(t, hasLimitErr)
}

func TestValidateRoomQueryRequest_ValidPage(t *testing.T) {
	req := &dtos.GetAvailableRoomsRequest{PageStr: "2"}
	errs := utils.ValidateRoomQueryRequest(req)
	assert.Empty(t, errs)
	assert.Equal(t, 2, req.Page)
}

func TestValidateRoomQueryRequest_InvalidPage(t *testing.T) {
	req := &dtos.GetAvailableRoomsRequest{PageStr: "0"}
	errs := utils.ValidateRoomQueryRequest(req)
	assert.NotEmpty(t, errs)
	assert.Equal(t, "page", errs[0].Field)
}

func TestValidateRoomQueryRequest_NonNumericPage(t *testing.T) {
	req := &dtos.GetAvailableRoomsRequest{PageStr: "abc"}
	errs := utils.ValidateRoomQueryRequest(req)
	assert.NotEmpty(t, errs)
	assert.Equal(t, "page", errs[0].Field)
}

func TestValidateRoomQueryRequest_ValidLimit(t *testing.T) {
	req := &dtos.GetAvailableRoomsRequest{PageStr: "1", LimitStr: "50"}
	errs := utils.ValidateRoomQueryRequest(req)
	assert.Empty(t, errs)
	assert.Equal(t, 50, req.Limit)
}

func TestValidateRoomQueryRequest_LimitOver100(t *testing.T) {
	req := &dtos.GetAvailableRoomsRequest{PageStr: "1", LimitStr: "200"}
	errs := utils.ValidateRoomQueryRequest(req)
	hasLimitErr := false
	for _, e := range errs {
		if e.Field == "limit" { hasLimitErr = true }
	}
	assert.True(t, hasLimitErr)
}

func TestValidateRoomQueryRequest_LimitZero(t *testing.T) {
	req := &dtos.GetAvailableRoomsRequest{PageStr: "1", LimitStr: "0"}
	errs := utils.ValidateRoomQueryRequest(req)
	hasLimitErr := false
	for _, e := range errs {
		if e.Field == "limit" { hasLimitErr = true }
	}
	assert.True(t, hasLimitErr)
}

// ─── MapErrorCode ──────────────────────────────────────────────────────────────

func TestMapErrorCode_AllCodes(t *testing.T) {
	cases := []struct{ code string; expected int }{
		{errorcodes.HMS_REC_404, http.StatusNotFound},
		{errorcodes.HMS_VAL_001, http.StatusBadRequest},
		{errorcodes.HMS_VAL_002, http.StatusBadRequest},
		{errorcodes.HMS_VAL_003, http.StatusUnprocessableEntity},
		{errorcodes.HMS_CONFLICT_409, http.StatusConflict},
		{errorcodes.HMS_REQ_001, http.StatusBadRequest},
		{errorcodes.HMS_DB_001, http.StatusInternalServerError},
		{errorcodes.HMS_DB_002, http.StatusInternalServerError},
		{"HMS_UNKNOWN", http.StatusInternalServerError},
	}
	for _, c := range cases {
		t.Run(c.code, func(t *testing.T) {
			input := &dtos.APIResponse{Errors: []dtos.Error{{Code: c.code}}}
			assert.Equal(t, c.expected, utils.MapErrorCode(input).Code)
		})
	}
}

func TestMapErrorCode_NoErrors_ExistingCode(t *testing.T) {
	input := &dtos.APIResponse{Code: http.StatusOK, Errors: []dtos.Error{}}
	assert.Equal(t, http.StatusOK, utils.MapErrorCode(input).Code)
}

func TestMapErrorCode_NoErrors_ZeroCode(t *testing.T) {
	input := &dtos.APIResponse{Code: 0, Errors: []dtos.Error{}}
	assert.Equal(t, http.StatusOK, utils.MapErrorCode(input).Code)
}

// ─── ErrorResponse ─────────────────────────────────────────────────────────────

func TestErrorResponse(t *testing.T) {
	errs := []dtos.Error{{Field: "name", Message: "required", Code: errorcodes.HMS_VAL_001}}
	resp := utils.ErrorResponse("Validation failed", errs)
	assert.Equal(t, "Error", resp.Status)
	assert.Equal(t, "Validation failed", resp.Message)
	assert.Len(t, resp.Errors, 1)
}

// ─── SuccessResponse ───────────────────────────────────────────────────────────

func TestSuccessResponse(t *testing.T) {
	resp := utils.SuccessResponse("Created", http.StatusCreated, map[string]string{"k": "v"})
	assert.Equal(t, "Success", resp.Status)
	assert.Equal(t, http.StatusCreated, resp.Code)
	assert.Equal(t, "Created", resp.Message)
	assert.NotNil(t, resp.Data)
}

// ─── ParseDate ─────────────────────────────────────────────────────────────────

func TestParseDate_Valid(t *testing.T) {
	result, err := utils.ParseDate("check_in", "2025-06-01T14:00:00Z")
	assert.Nil(t, err)
	assert.Equal(t, 2025, result.Year())
	assert.Equal(t, time.June, result.Month())
}

func TestParseDate_InvalidFormat(t *testing.T) {
	_, err := utils.ParseDate("check_in", "2025-06-01")
	assert.NotNil(t, err)
	assert.Equal(t, "check_in", err.Field)
	assert.Equal(t, errorcodes.HMS_VAL_001, err.Code)
}

func TestParseDate_Empty(t *testing.T) {
	_, err := utils.ParseDate("check_in", "")
	assert.NotNil(t, err)
}

// ─── ValidateDateRange ─────────────────────────────────────────────────────────

func TestValidateDateRange_Valid(t *testing.T) {
	now := time.Now()
	assert.Nil(t, utils.ValidateDateRange(now, now.Add(24*time.Hour)))
}

func TestValidateDateRange_Equal(t *testing.T) {
	now := time.Now()
	err := utils.ValidateDateRange(now, now)
	assert.NotNil(t, err)
	assert.Equal(t, "check_out", err.Field)
	assert.Equal(t, errorcodes.HMS_VAL_001, err.Code)
}

func TestValidateDateRange_Before(t *testing.T) {
	now := time.Now()
	assert.NotNil(t, utils.ValidateDateRange(now, now.Add(-time.Hour)))
}

// ─── ValidateCheckOutExtension ─────────────────────────────────────────────────

func TestValidateCheckOutExtension_Valid(t *testing.T) {
	now := time.Now()
	assert.Nil(t, utils.ValidateCheckOutExtension(now, now.Add(24*time.Hour)))
}

func TestValidateCheckOutExtension_Equal(t *testing.T) {
	now := time.Now()
	err := utils.ValidateCheckOutExtension(now, now)
	assert.NotNil(t, err)
	assert.Equal(t, "check_out", err.Field)
}

func TestValidateCheckOutExtension_Before(t *testing.T) {
	now := time.Now()
	assert.NotNil(t, utils.ValidateCheckOutExtension(now, now.Add(-time.Hour)))
}

// ─── SingleError ───────────────────────────────────────────────────────────────

func TestSingleError(t *testing.T) {
	errs := utils.SingleError("field", "msg", errorcodes.HMS_VAL_001)
	assert.Len(t, errs, 1)
	assert.Equal(t, "field", errs[0].Field)
	assert.Equal(t, "msg", errs[0].Message)
	assert.Equal(t, errorcodes.HMS_VAL_001, errs[0].Code)
}

// ─── GenerateUUID ──────────────────────────────────────────────────────────────

func TestGenerateUUID(t *testing.T) {
	u1, u2 := utils.GenerateUUID(), utils.GenerateUUID()
	assert.NotEmpty(t, u1)
	assert.NotEqual(t, u1, u2)
	assert.Len(t, u1, 36)
}

// ─── FormatTime ────────────────────────────────────────────────────────────────

func TestFormatTime_Valid(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	assert.Equal(t, now.Format(time.RFC3339), utils.FormatTime(&now))
}

func TestFormatTime_Nil(t *testing.T) {
	assert.Equal(t, "", utils.FormatTime(nil))
}

// ─── ParseUUID ─────────────────────────────────────────────────────────────────

func TestParseUUID_Valid(t *testing.T) {
	parsed, err := utils.ParseUUID("room_uuid", "550e8400-e29b-41d4-a716-446655440000")
	assert.Nil(t, err)
	assert.NotEmpty(t, parsed.String())
}

func TestParseUUID_Invalid(t *testing.T) {
	_, err := utils.ParseUUID("room_uuid", "not-a-uuid")
	assert.NotNil(t, err)
	assert.Equal(t, "room_uuid", err.Field)
	assert.Equal(t, errorcodes.HMS_VAL_001, err.Code)
}

func TestParseUUID_Empty(t *testing.T) {
	_, err := utils.ParseUUID("room_uuid", "")
	assert.NotNil(t, err)
}
