package errorcodes

// Error codes constants
const (
	// Request errors
	HMS_REQ_001 = "HMS_REQ_001" // Invalid JSON format

	// Validation errors
	HMS_VAL_001 = "HMS_VAL_001" // Validation failed 
	HMS_VAL_002 = "HMS_VAL_002" // Bad request 
	HMS_VAL_003 = "HMS_VAL_003" // Unprocessable entity 

	// Record errors
	HMS_REC_404 = "HMS_REC_404" // Record not found

	// Conflict errors
	HMS_CONFLICT_409 = "HMS_CONFLICT_409" // Conflict 

	// Database errors
	HMS_DB_001 = "HMS_DB_001" // Internal server / database error
	HMS_DB_002 = "HMS_DB_002" // Database not reachable
)

// ErrorCodeToHTTPStatus maps internal error codes to HTTP status codes
var ErrorCodeToHTTPStatus = map[string]int{
	HMS_REC_404:      404,
	HMS_VAL_001:      400,
	HMS_VAL_002:      400,
	HMS_VAL_003:      422,
	HMS_CONFLICT_409: 409,
	HMS_REQ_001:      400,
	HMS_DB_001:       500,
	HMS_DB_002:       500,
}