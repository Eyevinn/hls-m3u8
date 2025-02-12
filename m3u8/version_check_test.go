package m3u8

import (
	"testing"
)

func TestDefaultMatchingRule_Validate(t *testing.T) {
	rule := DefaultMatchingRule{}
	valid, err := rule.Validate()
	if !valid {
		t.Errorf("expected true, got false")
	}
	if err != (VersionMismatch{}) {
		t.Errorf("expected empty error, got %v", err)
	}
}

func TestValidIVInEXTXKEY(t *testing.T) {
	tests := []struct {
		name          string
		actualVersion uint8
		iv            string
		expectedValid bool
		expectedError VersionMismatch
	}{
		{
			name:          "Valid case with empty IV",
			actualVersion: 1,
			iv:            "",
			expectedValid: true,
			expectedError: VersionMismatch{},
		},
		{
			name:          "Invalid case with non-empty IV",
			actualVersion: 1,
			iv:            "someIV",
			expectedValid: false,
			expectedError: VersionMismatch{
				ActualVersion:   1,
				ExpectedVersion: 2,
				Description:     "Protocol version needs to be at least 2 if you have IV in EXT-X-KEY.",
			},
		},
		{
			name:          "Valid case with non-empty IV",
			actualVersion: 2,
			iv:            "someIV",
			expectedValid: true,
			expectedError: VersionMismatch{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rule := ValidIVInEXTXKey{
				ActualVersion: tt.actualVersion,
				IV:            tt.iv,
			}
			valid, err := rule.Validate()
			if valid != tt.expectedValid {
				t.Errorf("expected %v, got %v", tt.expectedValid, valid)
			}
			if err != tt.expectedError {
				t.Errorf("expected %v, got %v", tt.expectedError, err)
			}
		})
	}
}
func TestFloatPointDuration(t *testing.T) {
	tests := []struct {
		name          string
		actualVersion uint8
		duration      string
		expectedValid bool
		expectedError VersionMismatch
	}{
		{
			name:          "Valid case with integer duration",
			actualVersion: 1,
			duration:      "10",
			expectedValid: true,
			expectedError: VersionMismatch{},
		},
		{
			name:          "Invalid case with floating point duration and version less than 3",
			actualVersion: 2,
			duration:      "10.5",
			expectedValid: false,
			expectedError: VersionMismatch{
				ActualVersion:   2,
				ExpectedVersion: 3,
				Description:     "Protocol version needs to be at least 3 if you have floating point duration.",
			},
		},
		{
			name:          "Valid case with floating point duration and version 3 or higher",
			actualVersion: 3,
			duration:      "10.5",
			expectedValid: true,
			expectedError: VersionMismatch{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rule := FloatPointDuration{
				ActualVersion: tt.actualVersion,
				duration:      tt.duration,
			}
			valid, err := rule.Validate()
			if valid != tt.expectedValid {
				t.Errorf("expected %v, got %v", tt.expectedValid, valid)
			}
			if err != tt.expectedError {
				t.Errorf("expected %v, got %v", tt.expectedError, err)
			}
		})
	}
}
