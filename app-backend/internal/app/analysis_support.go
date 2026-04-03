package app

func closestSupportBelow(points []ohlcPoint, target float64) (float64, bool) {
	if len(points) < 7 {
		return 0, false
	}
	best := 0.0
	for i := 3; i < len(points)-3; i++ {
		low := points[i].Low
		if low <= 0 || low >= target {
			continue
		}
		if !isSwingLow(points, i) {
			continue
		}
		if low > best {
			best = low
		}
	}
	if best == 0 {
		return 0, false
	}
	return best, true
}

func isSwingLow(points []ohlcPoint, index int) bool {
	if index < 3 || index+3 >= len(points) {
		return false
	}
	low := points[index].Low
	if low <= 0 {
		return false
	}
	for i := index - 3; i <= index-1; i++ {
		if points[i].Low < low {
			return false
		}
	}
	for i := index + 1; i <= index+3; i++ {
		if points[i].Low < low {
			return false
		}
	}
	return true
}
