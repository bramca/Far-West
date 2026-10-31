package utils

import "math"

func AngleBetweenPoints(x1, y1, x2, y2 float64) float64 {
	return math.Atan2(y2-y1, x2-x1)
}

func DistanceBetweenPoints(x1, y1, x2, y2 float64) float64 {
	return math.Sqrt((x2-x1)*(x2-x1) + (y2-y1)*(y2-y1))
}

// LeadAngle returns the angle to aim at a moving target so a projectile with
// the given speed catches it. leadFactor scales the prediction (1 aims exactly
// where the target will be, less keeps the shots easier to dodge) and the lead
// is clamped to maxLead so the aim never flies too far from the target.
func LeadAngle(fromX, fromY, targetX, targetY, targetVelX, targetVelY, projectileSpeed, leadFactor, maxLead float64) float64 {
	dist := DistanceBetweenPoints(fromX, fromY, targetX, targetY)
	if projectileSpeed <= 0 || dist <= 0 {
		return AngleBetweenPoints(fromX, fromY, targetX, targetY)
	}

	timeOfFlight := dist / projectileSpeed
	leadX := targetVelX * timeOfFlight * leadFactor
	leadY := targetVelY * timeOfFlight * leadFactor
	leadDist := math.Hypot(leadX, leadY)
	if leadDist > maxLead {
		leadX = leadX * maxLead / leadDist
		leadY = leadY * maxLead / leadDist
	}

	return AngleBetweenPoints(fromX, fromY, targetX+leadX, targetY+leadY)
}
