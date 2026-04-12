package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"room-booking/internal/domain"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockSlotListService struct {
	mock.Mock
}

func (m *MockSlotListService) GetAvailableSlots(ctx context.Context, roomID, dateStr string) ([]domain.Slot, error) {
	args := m.Called(ctx, roomID, dateStr)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.Slot), args.Error(1)
}

func TestSlotListHandler_ServeHTTP(t *testing.T) {
	t.Parallel()

	roomID := uuid.New().String()
	validDate := "2024-06-10"

	tests := []struct {
		name           string
		setupMock      func(*MockSlotListService)
		queryDate      string
		pathRoomID     string
		expectedStatus int
		expectedCode   string
		checkResponse  func(t *testing.T, body []byte)
	}{
		{
			name: "200 success - empty slots list",
			setupMock: func(m *MockSlotListService) {
				m.On("GetAvailableSlots", mock.Anything, roomID, validDate).
					Return([]domain.Slot{}, nil)
			},
			queryDate:      validDate,
			pathRoomID:     roomID,
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, body []byte) {
				var resp map[string]any
				require.NoError(t, json.Unmarshal(body, &resp))
				slots, ok := resp["slots"].([]any)
				require.True(t, ok)
				assert.Empty(t, slots)
			},
		},
		{
			name: "200 success - with slots",
			setupMock: func(m *MockSlotListService) {
				slot, _ := domain.NewSlot(
					uuid.New().String(), roomID,
					time.Date(2024, 6, 10, 9, 0, 0, 0, time.UTC),
					time.Date(2024, 6, 10, 9, 30, 0, 0, time.UTC),
				)
				m.On("GetAvailableSlots", mock.Anything, roomID, validDate).
					Return([]domain.Slot{*slot}, nil)
			},
			queryDate:      validDate,
			pathRoomID:     roomID,
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, body []byte) {
				var resp map[string]any
				require.NoError(t, json.Unmarshal(body, &resp))
				slots := resp["slots"].([]any)
				require.Len(t, slots, 1)
				slotObj := slots[0].(map[string]any)
				assert.Equal(t, roomID, slotObj["roomId"])
				assert.Equal(t, "2024-06-10T09:00:00Z", slotObj["start"])
			},
		},
		{
			name:           "400 missing date parameter",
			setupMock:      func(m *MockSlotListService) {},
			queryDate:      "",
			pathRoomID:     roomID,
			expectedStatus: http.StatusBadRequest,
			expectedCode:   "InvalidRequest",
		},
		{
			name:           "400 invalid date format",
			setupMock:      func(m *MockSlotListService) {},
			queryDate:      "10/06/2024",
			pathRoomID:     roomID,
			expectedStatus: http.StatusBadRequest,
			expectedCode:   "InvalidRequest",
		},
		{
			name:           "400 missing roomId",
			setupMock:      func(m *MockSlotListService) {},
			queryDate:      validDate,
			pathRoomID:     "",
			expectedStatus: http.StatusBadRequest,
			expectedCode:   "InvalidRequest",
		},
		{
			name: "404 room not found",
			setupMock: func(m *MockSlotListService) {
				m.On("GetAvailableSlots", mock.Anything, roomID, validDate).
					Return(nil, domain.ErrRoomNotFound)
			},
			queryDate:      validDate,
			pathRoomID:     roomID,
			expectedStatus: http.StatusNotFound,
			expectedCode:   "NotFound",
		},
		{
			name: "500 internal error",
			setupMock: func(m *MockSlotListService) {
				m.On("GetAvailableSlots", mock.Anything, roomID, validDate).
					Return(nil, assert.AnError)
			},
			queryDate:      validDate,
			pathRoomID:     roomID,
			expectedStatus: http.StatusInternalServerError,
			expectedCode:   "InternalError",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockSvc := new(MockSlotListService)
			if tt.setupMock != nil {
				tt.setupMock(mockSvc)
			}
			handler := NewSlotListHandler(mockSvc)

			url := "/rooms/" + tt.pathRoomID + "/slots/list"
			if tt.queryDate != "" {
				url += "?date=" + tt.queryDate
			}
			req := httptest.NewRequest(http.MethodGet, url, nil)

			if tt.pathRoomID != "" {
				ctx := context.WithValue(req.Context(), CtxRoomID, tt.pathRoomID)
				req = req.WithContext(ctx)
			}

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)

			if tt.expectedCode != "" {
				var resp map[string]any
				require.NoError(t, json.NewDecoder(rr.Body).Decode(&resp))
				errObj, ok := resp["error"].(map[string]any)
				require.True(t, ok)
				assert.Equal(t, tt.expectedCode, errObj["code"])
			}

			if tt.checkResponse != nil {
				tt.checkResponse(t, rr.Body.Bytes())
			}

			mockSvc.AssertExpectations(t)
		})
	}
}
