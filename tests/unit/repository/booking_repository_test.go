package repository_test

import (
	"context"
	"os"
	"testing"
	"time"

	"hotel-updated/internal/loggers"
	"hotel-updated/internal/models"
	"hotel-updated/internal/repository"
	"hotel-updated/pkg/database"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func setupMockDB(t *testing.T) (*database.Db, sqlmock.Sqlmock, func()) {
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}

	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: sqlDB,
	}), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open gorm db: %v", err)
	}

	cleanup := func() {
		sqlDB.Close()
	}

	return &database.Db{Gorm: gormDB}, mock, cleanup
}

func setupLogger() *loggers.Logger {
	tmpDir, _ := os.MkdirTemp("", "booking-repo-test-logs-*")
	return loggers.NewLogger(loggers.LogConfig{
		Level:       "info",
		LogDir:      tmpDir,
		FileName:    "test.log",
		ServiceName: "test",
	})
}

func TestCreateBooking(t *testing.T) {
    db, mock, cleanup := setupMockDB(t)
    defer cleanup()

    repo := repository.NewBookingRepository(db, setupLogger())

    mock.ExpectBegin()

    mock.ExpectQuery(`INSERT INTO "bookings"`).
        WithArgs(
            sqlmock.AnyArg(), // guest_id
            sqlmock.AnyArg(), // room_id
            sqlmock.AnyArg(), // check_in
            sqlmock.AnyArg(), // check_out
            sqlmock.AnyArg(), // total_price
            sqlmock.AnyArg(), // created_at
            "admin",          // created_by
            sqlmock.AnyArg(), // updated_at
            "admin",          // updated_by
            true,             // is_active
            "uuid-1",         // booking_uuid
        ).
        WillReturnRows(
            sqlmock.NewRows([]string{"booking_uuid", "booking_id"}).
                AddRow("uuid-1", 1),
        )

    mock.ExpectCommit()

    booking := &models.Booking{
        BookingUUID: "uuid-1",
        IsActive:    true, // ✅ important
    }

    err := repo.CreateBooking(context.Background(), booking, "admin")

    assert.Nil(t, err)
    assert.Equal(t, 1, booking.BookingID)
    assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateBooking_Success(t *testing.T) {
    db, mock, cleanup := setupMockDB(t)
    defer cleanup()

    repo := repository.NewBookingRepository(db, setupLogger())

    mock.ExpectBegin()
    mock.ExpectQuery(`INSERT INTO "bookings"`).
        WillReturnRows(sqlmock.NewRows([]string{"booking_id"}).AddRow(1))
    mock.ExpectCommit()

    booking := &models.Booking{BookingUUID: "uuid-1"}

    err := repo.CreateBooking(context.Background(), booking, "admin")

    assert.Nil(t, err)
    assert.Equal(t, "admin", booking.CreatedBy) // ✅ ensures full success path hit
    assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateBooking_DBError(t *testing.T) {
    db, mock, cleanup := setupMockDB(t)
    defer cleanup()

    repo := repository.NewBookingRepository(db, setupLogger())

    mock.ExpectBegin()

    mock.ExpectQuery(`INSERT INTO "bookings"`).
        WillReturnError(assert.AnError) // ✅ force DB error

    mock.ExpectRollback()

    booking := &models.Booking{
        BookingUUID: "uuid-1",
        IsActive:    true,
    }

    err := repo.CreateBooking(context.Background(), booking, "admin")

    assert.NotNil(t, err) // ✅ triggers your uncovered block
    assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetBookingByUUID(t *testing.T) {
	db, mock, cleanup := setupMockDB(t)
	defer cleanup()

	repo := repository.NewBookingRepository(db, setupLogger())

	rows := sqlmock.NewRows([]string{"booking_id", "booking_uuid"}).
		AddRow(1, "uuid-1")

	mock.ExpectQuery(`SELECT .* FROM "bookings"`).
		WillReturnRows(rows)

	result, id, err := repo.GetBookingByUUID(context.Background(), "uuid-1")

	assert.Nil(t, err)
	assert.Equal(t, 1, id)
	assert.Equal(t, "uuid-1", result.BookingUUID)
}


func TestGetBookingByUUID_NotFound(t *testing.T) {
	db, mock, cleanup := setupMockDB(t)
	defer cleanup()

	repo := repository.NewBookingRepository(db, setupLogger())

	mock.ExpectQuery(`SELECT .* FROM "bookings"`).
		WillReturnError(gorm.ErrRecordNotFound)

	result, id, err := repo.GetBookingByUUID(context.Background(), "missing")

	assert.Nil(t, result)
	assert.Equal(t, 0, id)
	assert.NotNil(t, err)
}

func TestGetBookingByUUID_DBError(t *testing.T) {
    db, mock, cleanup := setupMockDB(t)
    defer cleanup()

    repo := repository.NewBookingRepository(db, setupLogger())

    mock.ExpectQuery(`SELECT .* FROM "bookings"`).
        WillReturnError(assert.AnError)

    result, id, err := repo.GetBookingByUUID(context.Background(), "uuid-1")

    assert.Nil(t, result)
    assert.Equal(t, 0, id)
    assert.NotNil(t, err)

    assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCancelBooking(t *testing.T) {
	db, mock, cleanup := setupMockDB(t)
	defer cleanup()

	repo := repository.NewBookingRepository(db, setupLogger())

	mock.ExpectBegin()

	mock.ExpectExec(`UPDATE "bookings" SET .* WHERE booking_id = \$4`).
		WithArgs(
			false,            // is_active
			"admin",          // updated_by
			sqlmock.AnyArg(), // updated_at (dynamic)
			1,                // booking_id
		).
		WillReturnResult(sqlmock.NewResult(0, 1))

	mock.ExpectCommit()

	err := repo.CancelBooking(context.Background(), 1, "admin")

	assert.Nil(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCancelBooking_DBError(t *testing.T) {
    db, mock, cleanup := setupMockDB(t)
    defer cleanup()

    repo := repository.NewBookingRepository(db, setupLogger())

    mock.ExpectBegin()
    mock.ExpectExec(`UPDATE "bookings"`).
        WillReturnError(assert.AnError)
    mock.ExpectRollback()

    err := repo.CancelBooking(context.Background(), 1, "admin")

    assert.NotNil(t, err)
    assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCancelBooking_NotFound(t *testing.T) {
    db, mock, cleanup := setupMockDB(t)
    defer cleanup()

    repo := repository.NewBookingRepository(db, setupLogger())

    mock.ExpectBegin()
    mock.ExpectExec(`UPDATE "bookings"`).
        WillReturnResult(sqlmock.NewResult(0, 0))
    mock.ExpectCommit()

    err := repo.CancelBooking(context.Background(), 99, "admin")

    assert.NotNil(t, err)
    assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRestoreBooking(t *testing.T) {
	db, mock, cleanup := setupMockDB(t)
	defer cleanup()

	repo := repository.NewBookingRepository(db, setupLogger())

	mock.ExpectBegin()

	mock.ExpectExec(`UPDATE "bookings" SET .* WHERE booking_id = \$4`).
		WithArgs(
			true,
			"admin",
			sqlmock.AnyArg(),
			1,
		).
		WillReturnResult(sqlmock.NewResult(0, 1))

	mock.ExpectCommit()

	err := repo.RestoreBooking(context.Background(), 1, "admin")

	assert.Nil(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRestoreBooking_DBError(t *testing.T) {
    db, mock, cleanup := setupMockDB(t)
    defer cleanup()

    repo := repository.NewBookingRepository(db, setupLogger())

    mock.ExpectBegin()
    mock.ExpectExec(`UPDATE "bookings"`).
        WillReturnError(assert.AnError)
    mock.ExpectRollback()

    err := repo.RestoreBooking(context.Background(), 1, "admin")

    assert.NotNil(t, err)
    assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRestoreBooking_NotFound(t *testing.T) {
    db, mock, cleanup := setupMockDB(t)
    defer cleanup()

    repo := repository.NewBookingRepository(db, setupLogger())

    mock.ExpectBegin()
    mock.ExpectExec(`UPDATE "bookings"`).
        WillReturnResult(sqlmock.NewResult(0, 0))
    mock.ExpectCommit()

    err := repo.RestoreBooking(context.Background(), 99, "admin")

    assert.NotNil(t, err)
    assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateGuest(t *testing.T) {
	db, mock, cleanup := setupMockDB(t)
	defer cleanup()

	repo := repository.NewBookingRepository(db, setupLogger())

	mock.ExpectBegin()

	mock.ExpectQuery(`INSERT INTO "guests"`).
		WithArgs(
			sqlmock.AnyArg(), // guest_name
			sqlmock.AnyArg(), // guest_age
			sqlmock.AnyArg(), // address
			"+123",           // phone_number
			sqlmock.AnyArg(), // created_at
			"admin",          // created_by
			sqlmock.AnyArg(), // updated_at
			"admin",          // updated_by
			true,             // is_active
			"g1",             // guest_uuid
		).
		WillReturnRows(
			sqlmock.NewRows([]string{"guest_uuid", "guest_id"}).
				AddRow("g1", 1),
		)

	mock.ExpectCommit()

	guest := &models.Guest{
		GuestUUID:   "g1",
		PhoneNumber: "+123",
		IsActive:    true,
	}

	err := repo.CreateGuest(context.Background(), guest, "admin")

	assert.Nil(t, err)
	assert.Equal(t, 1, guest.GuestID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateGuest_DBError(t *testing.T) {
    db, mock, cleanup := setupMockDB(t)
    defer cleanup()

    repo := repository.NewBookingRepository(db, setupLogger())

    mock.ExpectBegin()
    mock.ExpectQuery(`INSERT INTO "guests"`).
        WillReturnError(assert.AnError)
    mock.ExpectRollback()

    guest := &models.Guest{GuestUUID: "g1"}

    err := repo.CreateGuest(context.Background(), guest, "admin")

    assert.NotNil(t, err)
    assert.NoError(t, mock.ExpectationsWereMet())
}

func TestFindGuestByPhone_Reactivate(t *testing.T) {
    db, mock, cleanup := setupMockDB(t)
    defer cleanup()

    repo := repository.NewBookingRepository(db, setupLogger())

    // Step 1: SELECT
    rows := sqlmock.NewRows([]string{
        "guest_id",
        "guest_uuid",
        "phone_number",
        "is_active",
    }).
        AddRow(1, "g1", "+123", false)

    mock.ExpectQuery(`SELECT .* FROM "guests" WHERE phone_number = \$1 .*`).
        WithArgs("+123", 1). // ✅ FIX HERE
        WillReturnRows(rows)

    // Step 2: UPDATE
    mock.ExpectBegin()
    mock.ExpectExec(`UPDATE "guests" SET .* WHERE "guest_id" = \$4`).
        WithArgs(
            true,
            "admin",
            sqlmock.AnyArg(),
            1,
        ).
        WillReturnResult(sqlmock.NewResult(0, 1))
    mock.ExpectCommit()

    result, err := repo.FindGuestByPhoneAndReactivate(
        context.Background(),
        "+123",
        "admin",
    )

    assert.Nil(t, err)
    assert.NotNil(t, result) // ✅ prevent panic
    assert.True(t, result.IsActive)
    assert.NoError(t, mock.ExpectationsWereMet())
}

func TestFindGuestByPhone_AlreadyActive(t *testing.T) {
    db, mock, cleanup := setupMockDB(t)
    defer cleanup()

    repo := repository.NewBookingRepository(db, setupLogger())

    rows := sqlmock.NewRows([]string{
        "guest_id",
        "guest_uuid",
        "phone_number",
        "is_active",
    }).
        AddRow(1, "g1", "+123", true)

    mock.ExpectQuery(`SELECT .* FROM "guests"`).
        WithArgs("+123", 1). // ✅ same fix
        WillReturnRows(rows)

    result, err := repo.FindGuestByPhoneAndReactivate(
        context.Background(),
        "+123",
        "admin",
    )

    assert.Nil(t, err)
    assert.NotNil(t, result)
    assert.True(t, result.IsActive)

    assert.NoError(t, mock.ExpectationsWereMet())
}

func TestFindGuestByPhone_DBError(t *testing.T) {
    db, mock, cleanup := setupMockDB(t)
    defer cleanup()

    repo := repository.NewBookingRepository(db, setupLogger())

    mock.ExpectQuery(`SELECT .* FROM "guests"`).
        WithArgs("+123", 1).
        WillReturnError(assert.AnError)

    result, err := repo.FindGuestByPhoneAndReactivate(context.Background(), "+123", "admin")

    assert.Nil(t, result)
    assert.NotNil(t, err)

    assert.NoError(t, mock.ExpectationsWereMet())
}

func TestFindGuestByPhone_NotFound(t *testing.T) {
    db, mock, cleanup := setupMockDB(t)
    defer cleanup()

    repo := repository.NewBookingRepository(db, setupLogger())

    mock.ExpectQuery(`SELECT .* FROM "guests"`).
        WithArgs("+123", 1). // ✅ include LIMIT argument
        WillReturnError(gorm.ErrRecordNotFound)

    result, err := repo.FindGuestByPhoneAndReactivate(
        context.Background(),
        "+123",
        "admin",
    )

    assert.Nil(t, err)     // ✅ important
    assert.Nil(t, result)  // ✅ important

    assert.NoError(t, mock.ExpectationsWereMet())
}

func TestFindGuestByPhone_UpdateFail(t *testing.T) {
    db, mock, cleanup := setupMockDB(t)
    defer cleanup()

    repo := repository.NewBookingRepository(db, setupLogger())

    rows := sqlmock.NewRows([]string{"guest_id", "guest_uuid", "phone_number", "is_active"}).
        AddRow(1, "g1", "+123", false)

    mock.ExpectQuery(`SELECT .* FROM "guests"`).
        WithArgs("+123", 1).
        WillReturnRows(rows)

    mock.ExpectBegin()
    mock.ExpectExec(`UPDATE "guests"`).
        WillReturnError(assert.AnError)
    mock.ExpectRollback()

    result, err := repo.FindGuestByPhoneAndReactivate(context.Background(), "+123", "admin")

    assert.Nil(t, result)
    assert.NotNil(t, err)

    assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdateCheckOutAndTotalPrice(t *testing.T) {
    db, mock, cleanup := setupMockDB(t)
    defer cleanup()

    repo := repository.NewBookingRepository(db, setupLogger())

    mock.ExpectBegin()

    mock.ExpectExec(`UPDATE "bookings" SET .* WHERE booking_id = \$5`).
        WithArgs(
            sqlmock.AnyArg(), // ✅ check_out (time.Time)
            5000.0,           // ✅ total_price
            "admin",          // ✅ updated_by
            sqlmock.AnyArg(), // ✅ updated_at
            1,                // ✅ booking_id
        ).
        WillReturnResult(sqlmock.NewResult(0, 1))

    mock.ExpectCommit()

    err := repo.UpdateCheckOutAndTotalPrice(
        context.Background(),
        1,
        time.Now(),
        5000.0,
        "admin",
    )

    assert.Nil(t, err)
    assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdateCheckOutAndTotalPrice_NotFound(t *testing.T) {
    db, mock, cleanup := setupMockDB(t)
    defer cleanup()

    repo := repository.NewBookingRepository(db, setupLogger())

    mock.ExpectBegin()

    mock.ExpectExec(`UPDATE "bookings" SET .* WHERE booking_id = \$5`).
        WithArgs(
            sqlmock.AnyArg(),
            1000.0,
            "admin",
            sqlmock.AnyArg(),
            99,
        ).
        WillReturnResult(sqlmock.NewResult(0, 0)) // ✅ 0 rows

    mock.ExpectCommit()

    err := repo.UpdateCheckOutAndTotalPrice(
        context.Background(),
        99,
        time.Now(),
        1000.0,
        "admin",
    )

    assert.NotNil(t, err)
    assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdateCheckOutAndTotalPrice_DBError(t *testing.T) {
    db, mock, cleanup := setupMockDB(t)
    defer cleanup()

    repo := repository.NewBookingRepository(db, setupLogger())

    mock.ExpectBegin()

    mock.ExpectExec(`UPDATE "bookings" SET .* WHERE booking_id = \$5`).
        WillReturnError(assert.AnError)

    mock.ExpectRollback()

    err := repo.UpdateCheckOutAndTotalPrice(
        context.Background(),
        1,
        time.Now(),
        5000.0,
        "admin",
    )

    assert.NotNil(t, err)
    assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetGuestByID(t *testing.T) {
    db, mock, cleanup := setupMockDB(t)
    defer cleanup()

    repo := repository.NewBookingRepository(db, setupLogger())

    rows := sqlmock.NewRows([]string{
        "guest_id",
        "guest_uuid",
        "phone_number",
        "is_active",
    }).
        AddRow(1, "g1", "+123", true)

    mock.ExpectQuery(`SELECT .* FROM "guests" WHERE guest_id = \$1 .*`).
        WithArgs(1, 1). // ✅ second arg = LIMIT 1
        WillReturnRows(rows)

    result, err := repo.GetGuestByID(context.Background(), 1)

    assert.Nil(t, err)
    assert.NotNil(t, result)
    assert.Equal(t, 1, result.GuestID)

    assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetGuestByID_NotFound(t *testing.T) {
    db, mock, cleanup := setupMockDB(t)
    defer cleanup()

    repo := repository.NewBookingRepository(db, setupLogger())

    mock.ExpectQuery(`SELECT .* FROM "guests"`).
        WithArgs(99, 1). // ✅ include LIMIT arg
        WillReturnError(gorm.ErrRecordNotFound)

    result, err := repo.GetGuestByID(context.Background(), 99)

    assert.Nil(t, result)
    assert.NotNil(t, err)

    assert.NoError(t, mock.ExpectationsWereMet())
}