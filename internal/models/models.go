package models

import (
	"time"
)

type Guest struct {
	GuestID     int       `gorm:"column:guest_id;primaryKey;type:serial;autoIncrement" json:"guest_id"`
	GuestUUID   string    `gorm:"column:guest_uuid;type:varchar(36);unique;not null;default:gen_random_uuid()" json:"guest_uuid"`
	GuestName   string    `gorm:"column:guest_name;type:varchar(255);not null" json:"guest_name"`
	GuestAge    int       `gorm:"column:guest_age;type:integer;check:chk_age,guest_age > 1 AND guest_age < 150" json:"guest_age"`
	Address     string    `gorm:"column:address;type:text" json:"address"`
	PhoneNumber string    `gorm:"column:phone_number;type:varchar(15);unique;not null" json:"phone_number"`
	CreatedAt   time.Time `gorm:"column:created_at;type:timestamp;autoCreateTime" json:"created_at"`
	CreatedBy   string    `gorm:"column:created_by;type:varchar(50);default:'Admin'" json:"created_by"`
	UpdatedAt   time.Time `gorm:"column:updated_at;type:timestamp;autoUpdateTime" json:"updated_at"`
	UpdatedBy   string    `gorm:"column:updated_by;type:varchar(50)" json:"updated_by"`
	IsActive    bool      `gorm:"column:is_active;type:boolean;default:true" json:"is_active"`
}

type RoomCategory struct {
	RoomCategoryID   int       `gorm:"column:room_category_id;primaryKey;type:serial;autoIncrement" json:"room_category_id"`
	RoomCategoryUUID string    `gorm:"column:room_category_uuid;type:varchar(36);unique;not null;default:gen_random_uuid()" json:"room_category_uuid"`
	RoomCategoryName string    `gorm:"column:room_category_name;type:varchar(40);not null" json:"room_category_name"`
	CreatedAt        time.Time `gorm:"column:created_at;type:timestamp;autoCreateTime" json:"created_at"`
	CreatedBy        string    `gorm:"column:created_by;type:varchar(50);default:'Admin'" json:"created_by"`
	UpdatedAt        time.Time `gorm:"column:updated_at;type:timestamp;autoUpdateTime" json:"updated_at"`
	UpdatedBy        string    `gorm:"column:updated_by;type:varchar(50)" json:"updated_by"`
	IsActive         bool      `gorm:"column:is_active;type:boolean;default:true" json:"is_active"`
}

type Room struct {
	RoomID         int       `gorm:"column:room_id;primaryKey;type:serial;autoIncrement" json:"room_id"`
	RoomUUID       string    `gorm:"column:room_uuid;type:varchar(36);unique;not null;default:gen_random_uuid()" json:"room_uuid"`
	RoomCategoryID int       `gorm:"column:room_category_id;type:integer;foreignKey;references:RoomCategoryID" json:"room_category_id"`
	RoomNo         string    `gorm:"column:room_no;type:varchar(50);not null;unique" json:"room_no"`
	Price          float64   `gorm:"column:price;type:decimal(8,2);not null" json:"price"`
	Capacity       int       `gorm:"column:capacity;type:integer;not null;check:chk_capacity,capacity > 0" json:"capacity"`
	CreatedAt      time.Time `gorm:"column:created_at;type:timestamp;autoCreateTime" json:"created_at"`
	CreatedBy      string    `gorm:"column:created_by;type:varchar(50);default:'Admin'" json:"created_by"`
	UpdatedAt      time.Time `gorm:"column:updated_at;type:timestamp;autoUpdateTime" json:"updated_at"`
	UpdatedBy      string    `gorm:"column:updated_by;type:varchar(50)" json:"updated_by"`
	IsActive       bool      `gorm:"column:is_active;type:boolean;default:true" json:"is_active"`
}

type Booking struct {
	BookingID   int       `gorm:"column:booking_id;primaryKey;type:serial;autoIncrement" json:"booking_id"`
	BookingUUID string    `gorm:"column:booking_uuid;type:varchar(36);unique;not null;default:gen_random_uuid()" json:"booking_uuid"`
	GuestID     int       `gorm:"column:guest_id;type:integer;foreignKey;references:GuestID" json:"guest_id"`
	RoomID      int       `gorm:"column:room_id;type:integer;foreignKey;references:RoomID" json:"room_id"`
	CheckIn     time.Time `gorm:"column:check_in;type:timestamp;not null" json:"check_in"`
	CheckOut    time.Time `gorm:"column:check_out;type:timestamp;not null" json:"check_out"`
	TotalPrice  float64   `gorm:"column:total_price;type:decimal(10,2);not null" json:"total_price"`
	CreatedAt   time.Time `gorm:"column:created_at;type:timestamp;autoCreateTime" json:"created_at"`
	CreatedBy   string    `gorm:"column:created_by;type:varchar(50);default:'Admin'" json:"created_by"`
	UpdatedAt   time.Time `gorm:"column:updated_at;type:timestamp;autoUpdateTime" json:"updated_at"`
	UpdatedBy   string    `gorm:"column:updated_by;type:varchar(50)" json:"updated_by"`
	IsActive    bool      `gorm:"column:is_active;type:boolean;default:true" json:"is_active"`
}
