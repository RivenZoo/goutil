package time

import (
	"testing"
	"time"
)

func TestFirstDay(t *testing.T) {
	now := time.Now()
	dt := FirstDayOfMonth(now)
	if dt.Month() != now.Month() || dt.Day() != 1 {
		t.FailNow()
	}
	t.Log(now, dt)
	nextMon := FirstDayOfNextMonth(now)
	if nextMon.AddDate(0, -1, 0) != dt {
		t.FailNow()
	}
	t.Log(nextMon)
}
