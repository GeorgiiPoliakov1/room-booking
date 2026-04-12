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

type MockBookingCreator struct {
	mock.Mock
}

func (m *MockBookingCreator) Create(ctx context.Context, input service.BookingCreateInput) (*domain.Booking, error) {
	args := m.Called(ctx, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Booking), args.Error(1)
}

func makeTestBooking(t *testing.T, slotID, userID string, confLink *string) *domain.Booking {
	t.Helper()
	b, err := domain.NewBooking(domain.BookingCreateParams{
		ID:             uuid.New().String(),
		SlotID:         slotID,
		UserID:         userID,
		ConferenceLink: confLink,
	})
	require.NoError(t, err, "domain.NewBooking failed in test setup")
	return b
}

func TestBookingCreateHandler_ServeHTTP(t *testing.T) {
	t.Parallel()

	validSlotID := uuid.New().String()
	userID := uuid.New().String()

	tests := []struct {
		name           string
		setupMock      func(*MockBookingCreator)
		requestBody    string
		contextUserID  string
		expectedStatus int
		expectedCode   string
		checkResponse  func(t *testing.T, body []byte)
	}{
		{
			name: "201 created - without conference link",
			setupMock: func(m *MockBookingCreator) {
				booking := makeTestBooking(t, validSlotID, userID, nil)
				m.On("Create", mock.Anything, mock.Anything).Return(booking, nil)
			},
			requestBody:    `{"slotId": "` + validSlotID + `"}`,
			contextUserID:  userID,
			expectedStatus: http.StatusCreated,
			checkResponse: func(t *testing.T, body []byte) {
				var resp map[string]any
				require.NoError(t, json.Unmarshal(body, &resp))
				booking, ok := resp["booking"].(map[string]any)
				require.True(t, ok)
				assert.Equal(t, validSlotID, booking["slotId"])
				assert.Equal(t, "active", booking["status"])
			},
		},
		{
			name: "201 created - with conference link",
			setupMock: func(m *MockBookingCreator) {
				link := "https://meet.room-booking.com/abc-123"
				booking := makeTestBooking(t, validSlotID, userID, &link)
				m.On("Create", mock.Anything, mock.Anything).Return(booking, nil)
			},
			requestBody:    `{"slotId": "` + validSlotID + `", "createConferenceLink": true}`,
			contextUserID:  userID,
			expectedStatus: http.StatusCreated,
			checkResponse: func(t *testing.T, body []byte) {
				var resp map[string]any
				require.NoError(t, json.Unmarshal(body, &resp))
				booking := resp["booking"].(map[string]any)
				confLink, ok := booking["conferenceLink"].(string)
				require.True(t, ok, "conferenceLink should be a string")
				assert.Contains(t, confLink, "https://meet.room-booking.com/")
			},
		},
		{
			name:           "401 unauthorized - missing user in context",
			setupMock:      func(m *MockBookingCreator) {},
			requestBody:    `{"slotId": "` + validSlotID + `"}`,
			contextUserID:  "",
			expectedStatus: http.StatusUnauthorized,
			expectedCode:   "Unauthorized",
		},
		{
			name:           "400 bad request - malformed JSON",
			setupMock:      func(m *MockBookingCreator) {},
			requestBody:    `{"slotId": "invalid-json`,
			contextUserID:  userID,
			expectedStatus: http.StatusBadRequest,
			expectedCode:   "InvalidRequest",
		},
		{
			name: "400 bad request - slot in past",
			setupMock: func(m *MockBookingCreator) {
				m.On("Create", mock.Anything, mock.Anything).Return(nil, domain.ErrSlotInPast)
			},
			requestBody:    `{"slotId": "` + validSlotID + `"}`,
			contextUserID:  userID,
			expectedStatus: http.StatusBadRequest,
			expectedCode:   "INVALID_REQUEST",
		},
		{
			name: "404 not found - slot does not exist",
			setupMock: func(m *MockBookingCreator) {
				m.On("Create", mock.Anything, mock.Anything).Return(nil, domain.ErrSlotNotFound)
			},
			requestBody:    `{"slotId": "` + validSlotID + `"}`,
			contextUserID:  userID,
			expectedStatus: http.StatusNotFound,
			expectedCode:   "NotFound",
		},
		{
			name: "409 conflict - slot already booked",
			setupMock: func(m *MockBookingCreator) {
				m.On("Create", mock.Anything, mock.Anything).Return(nil, domain.ErrSlotAlreadyBooked)
			},
			requestBody:    `{"slotId": "` + validSlotID + `"}`,
			contextUserID:  userID,
			expectedStatus: http.StatusConflict,
			expectedCode:   "SLOT_ALREADY_BOOKED",
		},
		{
			name: "500 internal error",
			setupMock: func(m *MockBookingCreator) {
				m.On("Create", mock.Anything, mock.Anything).Return(nil, assert.AnError)
			},
			requestBody:    `{"slotId": "` + validSlotID + `"}`,
			contextUserID:  userID,
			expectedStatus: http.StatusInternalServerError,
			expectedCode:   "InternalError",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockSvc := new(MockBookingCreator)
			if tt.setupMock != nil {
				tt.setupMock(mockSvc)
			}

			handler := NewBookingCreateHandler(mockSvc)

			body := bytes.NewReader([]byte(tt.requestBody))
			req := httptest.NewRequest(http.MethodPost, "/bookings/create", body)
			req.Header.Set("Content-Type", "service/json")

			if tt.contextUserID != "" {
				ctx := context.WithValue(req.Context(), CtxUserID, tt.contextUserID)
				req = req.WithContext(ctx)
			}

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code, "status code mismatch")

			if tt.expectedCode != "" {
				var resp map[string]any
				require.NoError(t, json.NewDecoder(rr.Body).Decode(&resp), "response should be valid JSON")

				errObj, ok := resp["error"].(map[string]any)
				if ok {
					assert.Equal(t, tt.expectedCode, errObj["code"], "error code mismatch")
				}
			}

			if tt.checkResponse != nil {
				tt.checkResponse(t, rr.Body.Bytes())
			}

			mockSvc.AssertExpectations(t)
		})
	}
}
