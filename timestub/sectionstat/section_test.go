package sectionstat

import (
	"math/rand"
	"testing"
)

var (
	axis = []int{100, 200, 300, 500, 1000, 2000, 3000, 5000, 10000}
)

func testNewsectionSeries(axis []int, t *testing.T) {
	series := NewSectionSeries(axis)
	t.Logf("%v", series)
	sz := len(axis)
	if series.Len() != sz+1 {
		t.FailNow()
	}
}

func TestNewSectionSeries(t *testing.T) {
	testNewsectionSeries([]int{}, t)
	testNewsectionSeries(axis, t)
}

func testSectionSeriesSearch(axis []int, t *testing.T, target ...int64) {
	series := NewSectionSeries(axis)
	t.Logf("sections:%v", series)
	for _, n := range target {
		i := series.Search(n)
		r := series.Get(i)
		t.Logf("search %d, ret:%d, section:%d-%d", n, i, r.Low(), r.High())
		if n < r.Low() || n >= r.High() {
			t.FailNow()
		}
	}
}

func TestSectionSeries_Search(t *testing.T) {
	testSectionSeriesSearch([]int{}, t, int64(0), int64(1000))
	testSectionSeriesSearch(axis, t, int64(-1), int64(100), int64(1024))
}

func testSectionStatCollect(axis []int, times int, t *testing.T) {
	series := NewSectionSeries(axis)
	t.Logf("sections:%v", series)
	stat := NewSectionStat(series)
	sum := int64(0)
	for i := 0; i < times; i++ {
		n := int64(rand.Int31n(int32(100000)))
		stat.Collect(n)
		sum += n
	}
	if sum != stat.Sum() || int64(times) != stat.Count() {
		t.FailNow()
	}
	t.Logf("sum:%d,count:%d,mean:%d", stat.Sum(), stat.Count(), stat.Mean())
	t.Logf("section mean:%v", stat.SectionMean())

	ratio := float32(0.99)
	mean := stat.PercentMean(ratio)
	t.Logf("percent 0.99 mean:%d", mean)
	num := int64(0)
	threshold := int64(float32(times) * ratio)
	for i := range stat.sectionCounter {
		c := stat.sectionCounter[i]
		num += c.getCount()
		t.Logf("%d num:%d", i, num)
		if stat.sectionCounter[i].mean() == mean {
			if num < threshold {
				t.FailNow()
			}
		}
	}
}

func TestSectionStat_Collect(t *testing.T) {
	testSectionStatCollect(axis, 1000, t)
}

func BenchmarkSectionStat_Collect(b *testing.B) {
	series := NewSectionSeries(axis)
	stat := NewSectionStat(series)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			n := int64(rand.Int31n(int32(100000)))
			stat.Collect(n)
		}
	})
	b.ReportAllocs()
}
