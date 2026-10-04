package dtos

// CreateBookingRequest represents the request body for creating a booking
type CreateBookingRequest struct {
	Guest    GuestRequest `json:"guest" validate:"required"`
	RoomUUID string       `json:"room_uuid" validate:"required,uuid4"`
	CheckIn  string       `json:"check_in" validate:"required,datetime=2006-01-02T15:04:05Z07:00"`
	CheckOut string       `json:"check_out" validate:"required,datetime=2006-01-02T15:04:05Z07:00"`
}

// GuestRequest represents guest information in booking request
type GuestRequest struct {
	GuestName   string `json:"guest_name" validate:"required,min=1,max=255"`
	Age         int    `json:"age" validate:"required,min=1,max=150"`
	Address     string `json:"address" validate:"required"`
	PhoneNumber string `json:"phone_number" validate:"required,max=15"`
}

// ExtendBookingRequest represents the request body for extending a booking
type ExtendBookingRequest struct {
	BookingUUID string `json:"booking_uuid" validate:"required,uuid4"`
	CheckOut    string `json:"check_out" validate:"required,datetime=2006-01-02T15:04:05Z07:00"`
}

// CancelBookingRequest represents the request body for cancelling a booking
type CancelBookingRequest struct {
	BookingUUID string `json:"booking_uuid" validate:"required,uuid4"`
}

// RestoreBookingRequest represents the request body for restoring a booking
type RestoreBookingRequest struct {
	BookingUUID string `json:"booking_uuid" validate:"required,uuid4"`
}

// GetAvailableRoomsRequest represents query parameters for getting available rooms
type GetAvailableRoomsRequest struct {
	CheckIn          string `json:"check_in"`
	CheckOut         string `json:"check_out"`
	RoomCategoryUUID string `json:"room_category_uuid"`
	CapacityStr      string `json:"-"`
	PageStr          string `json:"-"`
	LimitStr         string `json:"-"`
	// Parsed values set after validation
	Capacity int `json:"-"`
	Page     int `json:"-"`
	Limit    int `json:"-"`
}

// ========== Response Data DTOs ==========

// RoomCategoryItem represents a room category in response
type RoomCategoryItem struct {
	RoomCategoryUUID string `json:"room_category_uuid"`
	RoomCategoryName string `json:"room_category_name"`
}

// RoomCategoryListData represents the data field for room categories response
type RoomCategoryListData struct {
	Categories []RoomCategoryItem `json:"categories"`
	Total      int                `json:"total"`
}

// RoomAvailabilityItem represents a room in response
type RoomAvailabilityItem struct {
	RoomUUID string  `json:"room_uuid"`
	RoomNo   string  `json:"room_no"`
	Price    float64 `json:"price"`
	IsActive bool    `json:"is_active"`
}

// PaginationMeta represents pagination metadata
type PaginationMeta struct {
	Page            int   `json:"page"`
	Limit           int   `json:"limit"`
	TotalRecords    int64 `json:"total_records"`
	TotalPages      int64 `json:"total_pages"`
	HasNextPage     bool  `json:"has_next_page"`
	HasPreviousPage bool  `json:"has_previous_page"`
}

// RoomListData represents the data field for available rooms response
type RoomListData struct {
	Rooms      []RoomAvailabilityItem `json:"rooms"`
	Pagination PaginationMeta         `json:"pagination"`
}

// BookingData represents booking details in response
type BookingData struct {
	BookingUUID string  `json:"booking_uuid"`
	GuestUUID   string  `json:"guest_uuid"`
	RoomUUID    string  `json:"room_uuid"`
	RoomNo      string  `json:"room_no"`
	CheckIn     string  `json:"check_in"`
	CheckOut    string  `json:"check_out"`
	TotalPrice  float64 `json:"total_price"`
	TotalDays   int     `json:"total_days"`
}

// CancelData represents cancellation response
type CancelData struct {
	Message string `json:"message"`
}

// RestoreData represents restore response
type RestoreData struct {
	Message string `json:"message"`
}
