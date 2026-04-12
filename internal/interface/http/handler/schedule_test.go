package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"room-booking/internal/domain"
	"room-booking/internal/service"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockScheduleCreator struct {
	mock.Mock
}

func (m *MockScheduleCreator) Create(ctx context.Context, input service.CreateScheduleInput) (*domain.Schedule, error) {
	args := m.Called(ctx, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Schedule), args.Error(1)
}

func TestScheduleCreateHandler_ServeHTTP(t *testing.T) {
	t.Parallel()

	validRoomID := uuid.New().String()

	tests := []struct {
		name           string
		setupMock      func(*MockScheduleCreator)
		requestBody    string
		pathRoomID     string
		expectedStatus int
		expectedCode   string
		expectedMsg    string
	}{
		{
			name: "201 created - success",
			setupMock: func(m *MockScheduleCreator) {
				m.On("Create", mock.Anything, mock.MatchedBy(func(input service.CreateScheduleInput) bool {
					return input.RoomID == validRoomID && len(input.DaysOfWeek) == 2
				})).Return(&domain.Schedule{}, nil)
			},
			requestBody:    `{"daysOfWeek":[1,3],"startTime":"09:00","endTime":"10:30"}`,
			pathRoomID:     validRoomID,
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "400 bad request - missing roomId",
			setupMock:      func(m *MockScheduleCreator) {},
			requestBody:    `{"daysOfWeek":[1],"startTime":"09:00","endTime":"10:00"}`,
			pathRoomID:     "",
			expectedStatus: http.StatusBadRequest,
			expectedCode:   "InvalidRequest",
		},
		{
			name:           "400 bad request - malformed JSON",
			setupMock:      func(m *MockScheduleCreator) {},
			requestBody:    `{"daysOfWeek":[1,"bad"],"startTime":"09:00"`,
			pathRoomID:     validRoomID,
			expectedStatus: http.StatusBadRequest,
			expectedCode:   "InvalidRequest",
		},
		{
			name: "400 bad request - domain validation (invalid day)",
			setupMock: func(m *MockScheduleCreator) {
				m.On("Create", mock.Anything, mock.Anything).Return(nil, domain.ErrInvalidDayOfWeek)
			},
			requestBody:    `{"daysOfWeek":[0],"startTime":"09:00","endTime":"10:00"}`,
			pathRoomID:     validRoomID,
			expectedStatus: http.StatusBadRequest,
			expectedCode:   "InvalidRequest",
		},
		{
			name: "404 not found - room does not exist",
			setupMock: func(m *MockScheduleCreator) {
				m.On("Create", mock.Anything, mock.Anything).Return(nil, domain.ErrRoomNotFound)
			},
			requestBody:    `{"daysOfWeek":[1],"startTime":"09:00","endTime":"10:00"}`,
			pathRoomID:     validRoomID,
			expectedStatus: http.StatusNotFound,
			expectedCode:   "NotFound",
		},
		{
			name: "409 conflict - schedule already exists",
			setupMock: func(m *MockScheduleCreator) {
				m.On("Create", mock.Anything, mock.Anything).Return(nil, domain.ErrScheduleAlreadyExists)
			},
			requestBody:    `{"daysOfWeek":[1],"startTime":"09:00","endTime":"10:00"}`,
			pathRoomID:     validRoomID,
			expectedStatus: http.StatusConflict,
			expectedCode:   "SCHEDULE_EXISTS",
		},
		{
			name: "500 internal error - infrastructure failure",
			setupMock: func(m *MockScheduleCreator) {
				m.On("Create", mock.Anything, mock.Anything).Return(nil, assert.AnError)
			},
			requestBody:    `{"daysOfWeek":[1],"startTime":"09:00","endTime":"10:00"}`,
			pathRoomID:     validRoomID,
			expectedStatus: http.StatusInternalServerError,
			expectedCode:   "InternalError",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockSvc := new(MockScheduleCreator)
			if tt.setupMock != nil {
				tt.setupMock(mockSvc)
			}

			handler := NewScheduleHandler(mockSvc)

			body := bytes.NewReader([]byte(tt.requestBody))
			req := httptest.NewRequest(http.MethodPost, "/rooms/"+tt.pathRoomID+"/schedule/create", body)
			req.Header.Set("Content-Type", "service/json")

			// Инжект roomId в контекст (эмуляция router-middleware)
			ctx := context.WithValue(req.Context(), CtxRoomID, tt.pathRoomID)
			req = req.WithContext(ctx)

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code, "status code mismatch")

			if tt.expectedCode != "" {
				var resp map[string]any
				require.NoError(t, json.NewDecoder(rr.Body).Decode(&resp), "response should be valid JSON")

				errObj, ok := resp["error"].(map[string]any)
				require.True(t, ok, "response must contain 'error' object")
				assert.Equal(t, tt.expectedCode, errObj["code"], "error code mismatch")
				if tt.expectedMsg != "" {
					assert.Contains(t, errObj["message"], tt.expectedMsg, "error message mismatch")
				}
			}

			mockSvc.AssertExpectations(t)
		})
	}
}
