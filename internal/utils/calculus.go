package utils

// IsAroundPoint определяет, находится ли точка samplePoint в окрестности
// точки refPoint с допустимым отклонением deviation
func IsAroundPoint(samplePoint float64, refPoint float64, deviation float64) bool {
	return samplePoint >= refPoint-refPoint*deviation &&
		samplePoint <= refPoint+refPoint*deviation
}

// IsBetweenIncl определеяет, находится ли точка samplePoint включительно
// между точками bound1 и bound2
func IsBetweenIncl(samplePoint float64, bound1 float64, bound2 float64) bool {
	return bound1 <= samplePoint && samplePoint <= bound2 ||
		bound2 <= samplePoint && samplePoint <= bound1
}
