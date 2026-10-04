package repository_test

import (
	"context"
	"hotel-updated/internal/repository"
	"testing"
	"time"
	"hotel-updated/internal/errorcodes"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestGetRoomByID_Success(t *testing.T) {
	db, mock, cleanup := setupMockDB(t)
	defer cleanup()

	repo := repository.NewRoomRepository(db, setupLogger())

	rows := sqlmock.NewRows([]string{
		"room_id", "room_uuid", "is_active",
	}).AddRow(1, "uuid-1", true)

	mock.ExpectQuery(`SELECT .* FROM "rooms"`).
		WithArgs(1, true, 1).
		WillReturnRows(rows)

	room, id, err := repo.GetRoomByID(context.Background(), 1)

	assert.Nil(t, err)
	assert.Equal(t, 1, id)
	assert.NotNil(t, room)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetRoomByID_NotFound(t *testing.T) {
	db, mock, cleanup := setupMockDB(t)
	defer cleanup()

	repo := repository.NewRoomRepository(db, setupLogger())

	mock.ExpectQuery(`SELECT .* FROM "rooms"`).
		WithArgs(1, true, 1).
		WillReturnError(gorm.ErrRecordNotFound)

	room, id, err := repo.GetRoomByID(context.Background(), 1)

	assert.Nil(t, room)
	assert.Equal(t, 0, id)
	assert.NotNil(t, err)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetRoomByID_DBError(t *testing.T) {
	db, mock, cleanup := setupMockDB(t)
	defer cleanup()

	repo := repository.NewRoomRepository(db, setupLogger())

	mock.ExpectQuery(`SELECT .* FROM "rooms"`).
		WithArgs(1, true, 1).
		WillReturnError(assert.AnError)

	room, id, err := repo.GetRoomByID(context.Background(), 1)

	assert.Nil(t, room)
	assert.Equal(t, 0, id)
	assert.NotNil(t, err)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetRoomByUUID_Success(t *testing.T) {
	db, mock, cleanup := setupMockDB(t)
	defer cleanup()

	repo := repository.NewRoomRepository(db, setupLogger())

	rows := sqlmock.NewRows([]string{"room_id", "room_uuid", "is_active"}).
		AddRow(5, "uuid-5", true)

	mock.ExpectQuery(`SELECT .* FROM "rooms"`).
		WithArgs("uuid-5", true, 1).
		WillReturnRows(rows)

	_, id, err := repo.GetRoomByUUID(context.Background(), "uuid-5")

	assert.Nil(t, err)
	assert.Equal(t, 5, id)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetAllCategories_Success(t *testing.T) {
	db, mock, cleanup := setupMockDB(t)
	defer cleanup()

	repo := repository.NewRoomRepository(db, setupLogger())

	rows := sqlmock.NewRows([]string{"room_category_id"}).
		AddRow(1).AddRow(2)

	mock.ExpectQuery(`SELECT .* FROM "room_categories"`).
		WithArgs(true).
		WillReturnRows(rows)

	res, err := repo.GetAllCategories(context.Background())

	assert.Nil(t, err)
	assert.Equal(t, 2, len(res))

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetAllCategories_Error(t *testing.T) {
	db, mock, cleanup := setupMockDB(t)
	defer cleanup()

	repo := repository.NewRoomRepository(db, setupLogger())

	mock.ExpectQuery(`SELECT .* FROM "room_categories"`).
		WithArgs(true).
		WillReturnError(assert.AnError)

	res, err := repo.GetAllCategories(context.Background())

	assert.Nil(t, res)
	assert.NotNil(t, err)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestIsRoomAvailable_True(t *testing.T) {
	db, mock, cleanup := setupMockDB(t)
	defer cleanup()

	repo := repository.NewRoomRepository(db, setupLogger())

	rows := sqlmock.NewRows([]string{"count"}).AddRow(0)

	mock.ExpectQuery(`SELECT count`).
		WillReturnRows(rows)

	ok, err := repo.IsRoomAvailable(context.Background(), 1, time.Now(), time.Now(), 0)

	assert.Nil(t, err)
	assert.True(t, ok)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestIsRoomAvailable_False(t *testing.T) {
	db, mock, cleanup := setupMockDB(t)
	defer cleanup()

	repo := repository.NewRoomRepository(db, setupLogger())

	rows := sqlmock.NewRows([]string{"count"}).AddRow(2)

	mock.ExpectQuery(`SELECT count`).
		WillReturnRows(rows)

	ok, err := repo.IsRoomAvailable(context.Background(), 1, time.Now(), time.Now(), 0)

	assert.Nil(t, err)
	assert.False(t, ok)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestIsRoomAvailable_DBError(t *testing.T) {
	db, mock, cleanup := setupMockDB(t)
	defer cleanup()

	repo := repository.NewRoomRepository(db, setupLogger())

	mock.ExpectQuery(`SELECT count`).
		WillReturnError(assert.AnError)

	ok, err := repo.IsRoomAvailable(context.Background(), 1, time.Now(), time.Now(), 0)

	assert.False(t, ok)
	assert.NotNil(t, err)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetRooms_CountError(t *testing.T) {
	db, mock, cleanup := setupMockDB(t)
	defer cleanup()

	repo := repository.NewRoomRepository(db, setupLogger())

	mock.ExpectQuery(`SELECT COUNT`).
		WillReturnError(assert.AnError)

	res, total, meta, err := repo.GetRooms(context.Background(), time.Now(), time.Now(), "", 0, 1, 10)

	assert.Nil(t, res)
	assert.Equal(t, int64(0), total)
	assert.Nil(t, meta)
	assert.NotNil(t, err)
}

func TestGetRooms_FetchError(t *testing.T) {
	db, mock, cleanup := setupMockDB(t)
	defer cleanup()

	repo := repository.NewRoomRepository(db, setupLogger())

	mock.ExpectQuery(`SELECT COUNT`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(5))

	mock.ExpectQuery(`SELECT rooms`).
		WillReturnError(assert.AnError)

	res, total, meta, err := repo.GetRooms(context.Background(), time.Now(), time.Now(), "", 0, 1, 10)

	assert.Nil(t, res)
	assert.Equal(t, int64(0), total)
	assert.Nil(t, meta)
	assert.NotNil(t, err)
}

func TestGetRooms_Success(t *testing.T) {
	db, mock, cleanup := setupMockDB(t)
	defer cleanup()

	repo := repository.NewRoomRepository(db, setupLogger())

	// ✅ FIXED COUNT query
	mock.ExpectQuery(`SELECT .*COUNT\(DISTINCT`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))

	// ✅ FETCH query
	rows := sqlmock.NewRows([]string{
		"room_id",
		"room_uuid",
		"is_active",
	}).
		AddRow(1, "uuid-1", true).
		AddRow(2, "uuid-2", true)

	mock.ExpectQuery(`SELECT .*rooms\.\*`).
		WillReturnRows(rows)

	res, total, meta, err := repo.GetRooms(
		context.Background(),
		time.Time{},
		time.Time{},
		"",
		0,
		1,
		10,
	)

	assert.Nil(t, err)
	assert.Equal(t, int64(2), total)
	assert.Equal(t, 2, len(res))
	assert.NotNil(t, meta)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetRooms_WithCategoryFilter(t *testing.T) {
	db, mock, cleanup := setupMockDB(t)
	defer cleanup()

	repo := repository.NewRoomRepository(db, setupLogger())

	mock.ExpectQuery(`SELECT .*COUNT`).
		WithArgs(true, true, "cat-uuid").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	rows := sqlmock.NewRows([]string{"room_id", "room_uuid"}).
		AddRow(1, "r1")

	mock.ExpectQuery(`SELECT .*rooms`).
		WillReturnRows(rows)

	res, _, _, err := repo.GetRooms(
		context.Background(),
		time.Time{},
		time.Time{},
		"cat-uuid", // ✅ triggers branch
		0,
		1,
		10,
	)

	assert.Nil(t, err)
	assert.Equal(t, 1, len(res))
}

func TestGetRooms_WithCapacityFilter(t *testing.T) {
	db, mock, cleanup := setupMockDB(t)
	defer cleanup()

	repo := repository.NewRoomRepository(db, setupLogger())

	mock.ExpectQuery(`SELECT .*COUNT`).
		WithArgs(true, true, 2).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	rows := sqlmock.NewRows([]string{"room_id", "room_uuid"}).
		AddRow(1, "r1")

	mock.ExpectQuery(`SELECT .*rooms`).
		WillReturnRows(rows)

	res, _, _, err := repo.GetRooms(
		context.Background(),
		time.Time{},
		time.Time{},
		"",
		2, // ✅ triggers branch
		1,
		10,
	)

	assert.Nil(t, err)
	assert.Equal(t, 1, len(res))
}

func TestGetRoomByUUID_NotFound(t *testing.T) {
	db, mock, cleanup := setupMockDB(t)
	defer cleanup()

	repo := repository.NewRoomRepository(db, setupLogger())

	mock.ExpectQuery(`SELECT .* FROM "rooms"`).
		WithArgs("missing", true, 1).
		WillReturnError(gorm.ErrRecordNotFound)

	room, id, err := repo.GetRoomByUUID(context.Background(), "missing")

	assert.Nil(t, room)
	assert.Equal(t, 0, id)
	assert.NotNil(t, err)
}

func TestGetRoomByUUID_DBError(t *testing.T) {
    db, mock, cleanup := setupMockDB(t)
    defer cleanup()

    repo := repository.NewRoomRepository(db, setupLogger())

    mock.ExpectQuery(`SELECT .* FROM "rooms"`).
        WithArgs("uuid-1", true, 1).
        WillReturnError(assert.AnError) // ✅ NOT ErrRecordNotFound

    room, id, err := repo.GetRoomByUUID(context.Background(), "uuid-1")

    assert.Nil(t, room)
    assert.Equal(t, 0, id)
    assert.NotNil(t, err)
    assert.Equal(t, errorcodes.HMS_DB_001, err.Code) // ✅ important

    assert.NoError(t, mock.ExpectationsWereMet())
}

func TestIsRoomAvailable_WithExcludeBookingID(t *testing.T) {
    db, mock, cleanup := setupMockDB(t)
    defer cleanup()

    repo := repository.NewRoomRepository(db, setupLogger())

    // ✅ important: match full query pattern loosely
    mock.ExpectQuery(`SELECT .*count.*FROM .*bookings`).
        WithArgs(
            1,                 // room_id
            sqlmock.AnyArg(),  // checkOut
            sqlmock.AnyArg(),  // checkIn
            10,                // excludeBookingID ✅ triggers branch
        ).
        WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

    ok, err := repo.IsRoomAvailable(
        context.Background(),
        1,
        time.Now(),
        time.Now().Add(24*time.Hour),
        10, // ✅ important
    )

    assert.Nil(t, err)
    assert.True(t, ok)

    assert.NoError(t, mock.ExpectationsWereMet())
}