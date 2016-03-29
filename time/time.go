package time

import "time"

func FirstDayOfMonth(dt time.Time) time.Time {
	return dt.AddDate(0, 0, 1-dt.Day())
}

func FirstDayOfNextMonth(dt time.Time) time.Time {
	dt = dt.AddDate(0, 1, 0)
	return FirstDayOfMonth(dt)
}
