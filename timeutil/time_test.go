package timeutil

import (
	"sync"
	"testing"
	"time"
)

func TestFirstDay(t *testing.T) {
	now := time.Now()
	dt := BeginOfMonth(now)
	if dt.Month() != now.Month() || dt.Day() != 1 {
		t.FailNow()
	}
	t.Log(now, dt)
	nextMon := BeginOfNextMonth(now)
	if nextMon.AddDate(0, -1, 0) != dt {
		t.FailNow()
	}
	t.Log(nextMon)
}

func TestTimeRecord(t *testing.T) {
	tr := timeRecord{}
	wg := &sync.WaitGroup{}
	for i := 0; i < 10; i++ {
		ch := make(chan struct{})
		wg.Add(1)
		go func() {
			defer wg.Done()
			close(ch)
			for j := 0; j < 100; j++ {
				tr.record(time.Duration(10+j) * time.Microsecond)
			}
		}()
		<-ch
	}
	wg.Wait()
	t.Log(tr.String())
}

func TestTimeStat(t *testing.T) {
	ts := NewTimeStat()
	key1, key2 := "url1", "url2"
	ts.RegistStat(key1, key2)
	wg := &sync.WaitGroup{}
	for i := 0; i < 10; i++ {
		ch := make(chan struct{})
		wg.Add(1)
		go func() {
			defer wg.Done()
			close(ch)
			for j := 0; j < 100; j++ {
				if j&0x01 != 0 {
					ts.Record(key1, time.Duration(10+j)*time.Microsecond)
				} else {
					ts.Record(key2, time.Duration(10+j)*time.Microsecond)
				}
			}
		}()
		<-ch
	}
	wg.Wait()
	t.Log(ts.Status())
}

func BenchmarkTimeStat(b *testing.B) {
	ts := NewTimeStat()
	key1, key2 := "url1", "url2"
	ts.RegistStat(key1, key2)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			t := time.Now()
			time.Sleep(time.Microsecond*500)
			cost := time.Since(t)
			ts.Record(key1, cost)
		}
	})
	b.Log(ts.Status())
}