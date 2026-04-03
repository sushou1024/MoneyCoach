package app

import "time"

func priceAsOf() time.Time {
	return time.Now().UTC().Truncate(time.Second)
}
