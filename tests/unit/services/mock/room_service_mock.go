package mock

import (
	"context"

	"hotel-updated/internal/dtos"

	"github.com/stretchr/testify/mock"
)

type RoomServiceMock struct{ mock.Mock }

func (m *RoomServiceMock) GetAvailableRooms(ctx context.Context, req dtos.GetAvailableRoomsRequest) *dtos.APIResponse {
	return m.Called(ctx, req).Get(0).(*dtos.APIResponse)
}
func (m *RoomServiceMock) GetRoomCategories(ctx context.Context) *dtos.APIResponse {
	return m.Called(ctx).Get(0).(*dtos.APIResponse)
}
