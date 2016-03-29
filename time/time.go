package time

import "time"

func BeginOfMonth(dt time.Time) time.Time {
	dt = time.Date(dt.Year(), dt.Month(), dt.Day(), 0, 0, 0, 0, time.UTC)
	return dt.AddDate(0, 0, 1-dt.Day())
}

func BeginOfNextMonth(dt time.Time) time.Time {
	dt = dt.AddDate(0, 1, 0)
	return BeginOfMonth(dt)
}
