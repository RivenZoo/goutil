package timeutil

import (
	"fmt"
	"strings"
	"sync/atomic"
	"time"
)

func BeginOfMonth(dt time.Time) time.Time {
	dt = time.Date(dt.Year(), dt.Month(), dt.Day(), 0, 0, 0, 0, time.UTC)
	return dt.AddDate(0, 0, 1-dt.Day())
}

func BeginOfNextMonth(dt time.Time) time.Time {
	dt = dt.AddDate(0, 1, 0)
	return BeginOfMonth(dt)
}

const (
	recordN    = 1 << 6
	recordMask = recordN - 1
)

type timeRecord struct {
	idx      uint32
	timeCost [recordN]int64
}

func (r *timeRecord) record(t time.Duration) {
	idx := atomic.AddUint32(&r.idx, 1)
	idx = idx & uint32(recordMask)

	atomic.StoreInt64(&r.timeCost[idx], int64(t))
}

func (r *timeRecord) String() string {
	total, n, avg := int64(0), int64(0), int64(-1)
	s := make([]string, 0, recordN)
	for _, t := range r.timeCost {
		if t != 0 {
			s = append(s, fmt.Sprintf("%d", t))
			total += t
			n++
		}
	}
	timeUnit := ""
	if n != 0 {
		avg = total / n
		timeUnit = "ns"
	}

	return fmt.Sprintf("{\"uint\":\"%s\",\"avg\":%d,\"recent\":[%s]}",
		timeUnit, avg, strings.Join(s, ","))
}

type TimeStat struct {
	stub map[string]*timeRecord
}

func NewTimeStat() *TimeStat {
	ts := &TimeStat{
		make(map[string]*timeRecord),
	}
	return ts
}

func (ts *TimeStat) regist(key string)  {
	if _, ok := ts.stub[key]; ok {
		return
	}
	ts.stub[key] = &timeRecord{}
}

func (ts *TimeStat) RegistStat(keys ...string) {
	for _, key := range keys {
		ts.regist(key)
	}
}

func (ts *TimeStat) Record(key string, t time.Duration) {
	if c, ok := ts.stub[key]; ok {
		c.record(t)
	}
}

func (ts *TimeStat) Status() string {
	var s []string
	for k, c := range ts.stub {
		s = append(s, fmt.Sprintf("{\"%s\":%s}", k, c))
	}
	return fmt.Sprintf("[%s]", strings.Join(s, ",\n"))
}

