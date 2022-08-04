package tistrategy

// IsAroundPoint determines if samplePoint lies within a range
// of refPoint with given relative deviation
func IsAroundPoint(samplePoint float64, refPoint float64, deviation float64) bool {
	return samplePoint >= refPoint-refPoint*deviation &&
		samplePoint <= refPoint+refPoint*deviation
}

// IsBetweenIncl determines if samplePoint lies between bound1 and bound2 both included
func IsBetweenIncl(samplePoint float64, bound1 float64, bound2 float64) bool {
	return bound1 <= samplePoint && samplePoint <= bound2 ||
		bound2 <= samplePoint && samplePoint <= bound1
}
