package custom_validator

import (
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTimeHHMM_Validation(t *testing.T) {
	validate := validator.New()
	err := validate.RegisterValidation("time_hhmm", TimeHHMM)
	require.NoError(t, err, "failed to register time_hhmm validator")
	type testStruct struct {
		Time string `validate:"time_hhmm"`
	}

	tests := []struct {
		name        string
		input       string
		expectError bool
		reason      string
	}{

		{"valid: midnight", "00:00", false, "lower bound"},
		{"valid: with leading zero", "09:30", false, "standard format"},
		{"valid: without leading zero", "9:30", false, "single digit hour allowed by spec"},
		{"valid: noon", "12:00", false, "midday"},
		{"valid: afternoon", "14:05", false, "24-hour format"},
		{"valid: max time", "23:59", false, "upper bound"},

		{"invalid: missing minutes", "9:3", true, "minutes must be two digits"},
		{"invalid: missing colon", "0930", true, "colon separator required"},
		{"invalid: extra colon", "09:30:00", true, "only one colon allowed"},
		{"invalid: letters", "ab:cd", true, "digits only"},
		{"invalid: spaces", "09 : 30", true, "no spaces allowed"},
		{"invalid: empty", "", true, "empty string not allowed"},
		{"invalid: nil-like", "null", true, "non-time string"},

		{"invalid: hour 24", "24:00", true, "max hour is 23"},
		{"invalid: hour 25", "25:00", true, "hour > 23"},
		{"invalid: hour 99", "99:99", true, "both out of range"},
		{"invalid: negative", "-1:00", true, "negative not allowed"},
		{"invalid: minute 60", "12:60", true, "max minute is 59"},
		{"invalid: minute 99", "10:99", true, "minute > 59"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := testStruct{Time: tt.input}
			err := validate.Struct(s)

			if tt.expectError {
				assert.Error(t, err, "expected error for input %q (%s)", tt.input, tt.reason)

				if assert.IsType(t, validator.ValidationErrors{}, err) {
					valErrs := err.(validator.ValidationErrors)
					require.Len(t, valErrs, 1, "expected exactly one validation error")
					assert.Equal(t, "Time", valErrs[0].Field(), "error should be on Time field")
					assert.Equal(t, "time_hhmm", valErrs[0].Tag(), "error should be from time_hhmm validator")
				}
			} else {
				assert.NoError(t, err, "unexpected error for valid input %q (%s)", tt.input, tt.reason)
			}
		})
	}
}

func TestIsUUID_Validation(t *testing.T) {
	validate := validator.New()
	err := validate.RegisterValidation("uuid_custom", IsUUID)
	require.NoError(t, err, "failed to register uuid_custom validator")

	type testStruct struct {
		ID string `validate:"uuid_custom"`
	}

	tests := []struct {
		name        string
		input       string
		expectError bool
		reason      string
	}{
		{"valid: lowercase", "550e8400-e29b-41d4-a716-446655440000", false, "standard lowercase"},
		{"valid: uppercase", "550E8400-E29B-41D4-A716-446655440000", false, "uppercase allowed"},
		{"valid: mixed case", "550e8400-E29b-41D4-a716-446655440000", false, "mixed case allowed"},
		{"valid: different uuid", "f47ac10b-58cc-4372-a567-0e02b2c3d479", false, "another valid UUID"},

		{"invalid: empty", "", true, "empty string not a UUID"},
		{"invalid: wrong segment length", "550e840-e29b-41d4-a716-446655440000", true, "first segment must be 8 chars"},
		{"invalid: extra chars", "550e8400-e29b-41d4-a716-446655440000-extra", true, "no extra suffix"},
		{"invalid: wrong char", "550e8400-e29b-41d4-a716-44665544000g", true, "g is not hex"},
		{"invalid: random string", "not-a-uuid", true, "obviously not UUID"},
		{"invalid: uuid v1 format but invalid", "01234567-89ab-cdef-0123-456789abcdef", false, "valid format regardless of version"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := testStruct{ID: tt.input}
			err := validate.Struct(s)

			if tt.expectError {
				assert.Error(t, err, "expected error for input %q (%s)", tt.input, tt.reason)
			} else {
				assert.NoError(t, err, "unexpected error for valid input %q (%s)", tt.input, tt.reason)
			}
		})
	}
}

func TestRegister_Integration(t *testing.T) {
	validate := validator.New()

	Register(validate)
	type scheduleRequest struct {
		RoomID    string `validate:"required,uuid_custom"`
		StartTime string `validate:"required,time_hhmm"`
		EndTime   string `validate:"required,time_hhmm"`
	}

	tests := []struct {
		name          string
		input         scheduleRequest
		expectError   bool
		expectedField string
	}{
		{
			name: "valid: all fields correct",
			input: scheduleRequest{
				RoomID:    "550e8400-e29b-41d4-a716-446655440000",
				StartTime: "09:00",
				EndTime:   "17:30",
			},
			expectError: false,
		},
		{
			name: "invalid: bad uuid",
			input: scheduleRequest{
				RoomID:    "not-a-uuid",
				StartTime: "09:00",
				EndTime:   "17:30",
			},
			expectError:   true,
			expectedField: "RoomID",
		},
		{
			name: "invalid: bad start time",
			input: scheduleRequest{
				RoomID:    "550e8400-e29b-41d4-a716-446655440000",
				StartTime: "25:00",
				EndTime:   "17:30",
			},
			expectError:   true,
			expectedField: "StartTime",
		},
		{
			name: "invalid: bad end time",
			input: scheduleRequest{
				RoomID:    "550e8400-e29b-41d4-a716-446655440000",
				StartTime: "09:00",
				EndTime:   "12:99",
			},
			expectError:   true,
			expectedField: "EndTime",
		},
		{
			name: "invalid: multiple errors",
			input: scheduleRequest{
				RoomID:    "bad",
				StartTime: "99:99",
				EndTime:   "ab:cd",
			},
			expectError: true,
		},
		{
			name: "invalid: required missing",
			input: scheduleRequest{
				RoomID:    "",
				StartTime: "",
				EndTime:   "",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validate.Struct(tt.input)

			if tt.expectError {
				require.Error(t, err, "expected validation error")
				if tt.expectedField != "" {
					if valErrs, ok := err.(validator.ValidationErrors); ok {
						found := false
						for _, e := range valErrs {
							if e.Field() == tt.expectedField {
								found = true
								break
							}
						}
						assert.True(t, found, "expected error on field %q, got errors: %v", tt.expectedField, valErrs)
					}
				}
			} else {
				require.NoError(t, err, "expected no validation error, got: %v", err)
			}
		})
	}
}

func TestRegister_Idempotent(t *testing.T) {
	validate := validator.New()

	Register(validate)

	Register(validate)

	type testStruct struct {
		Time string `validate:"time_hhmm"`
	}
	err := validate.Struct(testStruct{Time: "12:00"})
	assert.NoError(t, err, "validator should still work after duplicate registration")
}
