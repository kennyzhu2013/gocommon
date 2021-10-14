package media

import "math"

func AvgEnergy(data []byte) int {
	if len(data) == 0 {
		return 0
	}

	waves := Transform(data)
	var sum int
	for _, wave := range waves {
		sum += int(math.Abs(float64(wave)))
	}
	return sum / len(waves)
}
