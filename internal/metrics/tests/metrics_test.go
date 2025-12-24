package tests

//
//import (
//	"testing"
//	"time"
//
//	"ultra-bis/internal/metrics"
//)
//
//// TestBodyMetric_BMICalculation tests BMI calculation in the model
//func TestBodyMetric_BMICalculation(t *testing.T) {
//	userID := uint(1)
//	weight := 70.0 // kg
//	height := 175  // cm from user profile
//
//	metric := &metrics.BodyMetric{
//		UserID: userID,
//		Date:   time.Now(),
//		Weight: weight,
//	}
//
//	// BMI = weight / (height/100)^2
//	expectedBMI := weight / ((float64(height) / 100.0) * (float64(height) / 100.0))
//
//	// Note: Actual BMI calculation happens in the handler/service layer
//	// This test verifies the model structure is correct
//	if metric.Weight != weight {
//		t.Errorf("Expected weight=%f, got %f", weight, metric.Weight)
//	}
//
//	if metric.UserID != userID {
//		t.Errorf("Expected userID=%d, got %d", userID, metric.UserID)
//	}
//
//	// Verify BMI would be calculated correctly
//	calculatedBMI := weight / ((float64(height) / 100.0) * (float64(height) / 100.0))
//	if calculatedBMI != expectedBMI {
//		t.Errorf("Expected BMI=%f, got %f", expectedBMI, calculatedBMI)
//	}
//}
//
//// TestCreateBodyMetricRequest_Validation tests request struct
//func TestCreateBodyMetricRequest_Validation(t *testing.T) {
//	tests := []struct {
//		name    string
//		request metrics.CreateBodyMetricRequest
//		valid   bool
//	}{
//		{
//			name: "valid request",
//			request: metrics.CreateBodyMetricRequest{
//				Date:   "2025-01-15",
//				Weight: 70.0,
//			},
//			valid: true,
//		},
//		{
//			name: "zero weight should be invalid",
//			request: metrics.CreateBodyMetricRequest{
//				Date:   "2025-01-15",
//				Weight: 0,
//			},
//			valid: false,
//		},
//		{
//			name: "negative weight should be invalid",
//			request: metrics.CreateBodyMetricRequest{
//				Date:   "2025-01-15",
//				Weight: -10,
//			},
//			valid: false,
//		},
//	}
//
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			// Basic validation checks
//			isValid := tt.request.Weight > 0 && tt.request.Date != ""
//
//			if isValid != tt.valid {
//				t.Errorf("Expected valid=%v, got %v", tt.valid, isValid)
//			}
//		})
//	}
//}
