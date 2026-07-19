package handlers

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// -----------------------------------------------------------------
// 🎭 MANUAL INTERFACE MOCKS
// -----------------------------------------------------------------

type MockLocker struct {
	MockLock   func(ctx context.Context, resourceKey string, ttl time.Duration) (string, error)
	MockUnlock func(ctx context.Context, resourceKey string, token string) error
}

func (m *MockLocker) Lock(ctx context.Context, k string, t time.Duration) (string, error) {
	return m.MockLock(ctx, k, t)
}
func (m *MockLocker) Unlock(ctx context.Context, k string, tok string) error {
	return m.MockUnlock(ctx, k, tok)
}

type MockStorage struct {
	MockSaveBooking func(ctx context.Context, eventID, userID string, seats int) (string, error)
}

func (m *MockStorage) SaveBooking(ctx context.Context, e, u string, s int) (string, error) {
	return m.MockSaveBooking(ctx, e, u, s)
}

// -----------------------------------------------------------------
// ⚡ TABLE-DRIVEN UNIT TEST SUITE
// -----------------------------------------------------------------

func TestCreateBooking_TableDriven(t *testing.T) {
	// Standard structural variables for successful fallbacks
	defaultLockFunc := func(ctx context.Context, k string, d time.Duration) (string, error) { return "mock-token", nil }
	defaultUnlockFunc := func(ctx context.Context, k string, tok string) error { return nil }
	defaultSaveFunc := func(ctx context.Context, e, u string, s int) (string, error) { return "bkg-12345", nil }

	// Define our table architecture
	tests := []struct {
		name           string
		requestBody    string
		setupMockLock  func(ctx context.Context, k string, d time.Duration) (string, error)
		setupMockSave  func(ctx context.Context, e, u string, s int) (string, error)
		expectedStatus int
	}{
		{
			name:           "Success Path",
			requestBody:    `{"eventId":"dev_summit","userId":"usr_1","seats":2}`,
			setupMockLock:  defaultLockFunc,
			setupMockSave:  defaultSaveFunc,
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "Invalid JSON payload syntax",
			requestBody:    `{"eventId": "broken-json"...`,
			setupMockLock:  defaultLockFunc,
			setupMockSave:  defaultSaveFunc,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Missing structural validation requirements",
			requestBody:    `{"eventId":"","userId":"usr_1","seats":0}`,
			setupMockLock:  defaultLockFunc,
			setupMockSave:  defaultSaveFunc,
			expectedStatus: http.StatusUnprocessableEntity,
		},
		{
			name:        "Concurrency lock collision failure",
			requestBody: `{"eventId":"dev_summit","userId":"usr_1","seats":2}`,
			setupMockLock: func(ctx context.Context, k string, d time.Duration) (string, error) {
				return "", errors.New("lock acquisition timeout") // Simulate Redis being busy
			},
			setupMockSave:  defaultSaveFunc,
			expectedStatus: http.StatusConflict,
		},
		{
			name:          "Underlying SQL database write failure",
			requestBody:   `{"eventId":"dev_summit","userId":"usr_1","seats":2}`,
			setupMockLock: defaultLockFunc,
			setupMockSave: func(ctx context.Context, e, u string, s int) (string, error) {
				return "", errors.New("postgres connection dropped") // Simulate DB crash
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	// Run our execution loops matrix
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Instantiate mock wrapper engines dynamically per test case
			mockLocker := &MockLocker{MockLock: tt.setupMockLock, MockUnlock: defaultUnlockFunc}
			mockStorage := &MockStorage{MockSaveBooking: tt.setupMockSave}

			handler := NewBookingHandler(mockStorage, mockLocker)

			// Construct standard Go request recorders
			req, err := http.NewRequest("POST", "/book", bytes.NewBufferString(tt.requestBody))
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}
			req.Header.Set("Content-Type", "application/json")
			rr := httptest.NewRecorder()

			// Execute handler target logic
			handler.CreateBooking(rr, req)

			// Assertions checking matching state matrices
			if rr.Code != tt.expectedStatus {
				t.Errorf("Test Case [%s] Failed -> Expected status code %d, got %d", tt.name, tt.expectedStatus, rr.Code)
			}
		})
	}
}