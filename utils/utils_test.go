package utils

import (
	"math"
	"testing"
)

// TestLeadAngleStationaryTarget makes sure a target standing still gets shot
// straight in the face.
func TestLeadAngleStationaryTarget(t *testing.T) {
	got := LeadAngle(0, 0, 100, 50, 0, 0, 4.0, 1.0, 140)
	want := AngleBetweenPoints(0, 0, 100, 50)
	if math.Abs(got-want) > 1e-6 {
		t.Fatalf("LeadAngle = %f, want %f", got, want)
	}
}

// TestLeadAngleFullLead aims exactly at the future position of the target.
func TestLeadAngleFullLead(t *testing.T) {
	// target at (100,0) running straight away, time of flight is 25 frames
	got := LeadAngle(0, 0, 100, 0, 2, 0, 4.0, 1.0, 1000)
	if math.Abs(got) > 1e-6 {
		t.Fatalf("LeadAngle = %f, want %f", got, 0.0)
	}

	// target at (100,0) running up, the shot must be aimed above the line
	got = LeadAngle(0, 0, 100, 0, 0, 2, 4.0, 1.0, 1000)
	want := math.Atan2(50, 100)
	if math.Abs(got-want) > 1e-6 {
		t.Fatalf("LeadAngle = %f, want %f", got, want)
	}
}

// TestLeadAngleClamps keeps the lead from growing out of control.
func TestLeadAngleClamps(t *testing.T) {
	// a fast close target would otherwise get an enormous lead
	unclamped := LeadAngle(0, 0, 100, 0, 0, 2, 4.0, 1.0, 1000)
	got := LeadAngle(0, 0, 100, 0, 0, 2, 4.0, 1.0, 40)
	if got >= unclamped || got <= 0 {
		t.Fatalf("clamped angle %f should sit between 0 and %f", got, unclamped)
	}
}