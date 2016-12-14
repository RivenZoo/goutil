package sectionstat

import (
	"math"
	"sort"
	"sync/atomic"
)

const binarySearchThreshold = 16

type Range interface {
	Low() int64
	High() int64
	Contain(int64) bool
	Less(Range) bool
}

type SectionSeries interface {
	sort.Interface
	Get(int) Range
	Search(int64) int
}

// NewSectionSeries create section series by individual point.
// Eg. [100,200,500,1000] => [[minInt64,100],[100,200],[200,500],[500,1000],[1000,maxInt64]]
func NewSectionSeries(axis []int) SectionSeries {
	sz := len(axis)
	if sz == 0 {
		return sectionSeries{
			numberRange{
				low:  math.MinInt64,
				high: math.MaxInt64,
			},
		}
	}
	sort.Stable(sort.IntSlice(axis))
	ret := make(sectionSeries, 0, sz+1)
	ret = append(ret, numberRange{
		low:  math.MinInt64,
		high: int64(axis[0]),
	})
	for i := range axis {
		nr := numberRange{
			low: int64(axis[i]),
		}
		next := i + 1
		if next < sz {
			nr.high = int64(axis[next])
		} else {
			nr.high = math.MaxInt64
		}
		ret = append(ret, nr)
	}
	return ret
}

type numberRange struct {
	low, high int64
}

func (nr numberRange) Low() int64 {
	return nr.low
}

func (nr numberRange) High() int64 {
	return nr.high
}

func (nr numberRange) Contain(i int64) bool {
	return i >= nr.low && i < nr.high
}

func (nr numberRange) Less(r Range) bool {
	return nr.low < r.Low()
}

type sectionSeries []Range

func (r sectionSeries) Len() int {
	return len([]Range(r))
}

func (r sectionSeries) Swap(i, j int) {
	r[i], r[j] = r[j], r[i]
}

func (r sectionSeries) Less(i, j int) bool {
	return r[i].Less(r[j])
}

func (r sectionSeries) Get(i int) Range {
	return r[i]
}

func (r sectionSeries) linearSearch(n int64) int {
	for i := range r {
		if r[i].Contain(n) {
			return i
		}
	}
	return -1
}

func (r sectionSeries) binarySearch(n int64) int {
	sz := len(r)
	low, high := 0, sz-1
	for low <= high {
		mid := (low + high) / 2
		if r[mid].Contain(n) {
			return mid
		}
		if r[mid].Low() > n {
			high = mid - 1
		} else {
			low = mid + 1
		}
	}
	return -1
}

func (r sectionSeries) Search(n int64) int {
	if len(r) < binarySearchThreshold {
		return r.linearSearch(n)
	}
	return r.binarySearch(n)
}

type counter struct {
	sum, cnt int64
}

func (c *counter) collect(n int64) {
	atomic.AddInt64(&c.sum, n)
	atomic.AddInt64(&c.cnt, 1)
}

func (c *counter) reset() {
	atomic.StoreInt64(&c.sum, 0)
	atomic.StoreInt64(&c.cnt, 0)
}

func (c *counter) getSum() int64 {
	return atomic.LoadInt64(&c.sum)
}

func (c *counter) getCount() int64 {
	return atomic.LoadInt64(&c.cnt)
}

func (c *counter) mean() int64 {
	cnt := atomic.LoadInt64(&c.cnt)
	sum := atomic.LoadInt64(&c.sum)
	if cnt == 0 {
		return 0
	}
	return sum / cnt
}

// SectionStat stat number sample by section.
type SectionStat struct {
	sections       SectionSeries
	sectionCounter []counter
	sum            int64
	cnt            int64
}

// NewSectionStat create SectionStat struct by SectionSeries.
func NewSectionStat(sections SectionSeries) *SectionStat {
	return &SectionStat{
		sections:       sections,
		sectionCounter: make([]counter, sections.Len()),
	}
}

// Collect number sample. It's concurrency safe.
func (st *SectionStat) Collect(n int64) {
	atomic.AddInt64(&st.sum, n)
	atomic.AddInt64(&st.cnt, 1)
	i := st.sections.Search(n)
	if i != -1 {
		st.sectionCounter[i].collect(n)
	}
}

// Reset clean all statistic.
func (st *SectionStat) Reset() {
	atomic.StoreInt64(&st.cnt, 0)
	atomic.StoreInt64(&st.sum, 0)
	for i := range st.sectionCounter {
		st.sectionCounter[i].reset()
	}
}

// Mean return mean which equal sum/count.
func (st *SectionStat) Mean() int64 {
	cnt := atomic.LoadInt64(&st.cnt)
	sum := atomic.LoadInt64(&st.sum)
	if cnt == 0 {
		return int64(0)
	}
	return sum / cnt
}

// Sum return sum of Collect parameter.
func (st *SectionStat) Sum() int64 {
	return atomic.LoadInt64(&st.sum)
}

// Count return count of Collect call.
func (st *SectionStat) Count() int64 {
	return atomic.LoadInt64(&st.cnt)
}

// SectionMean return mean of all section.
func (st *SectionStat) SectionMean() []int64 {
	ret := make([]int64, len(st.sectionCounter))
	for i := range st.sectionCounter {
		ret[i] = st.sectionCounter[i].mean()
	}
	return ret
}

// PercentMean return section mean which n*percent times Collect call parameter less than high of that section.
// Eg. section{mean,count,[low,high]} [{5,6,[1,10]},{15,3,[10,20]},{22,1,[20,30]}]
// PercentMean(0.9) => (6+3+1)*0.9=9 => {15,3,[10,20]} => 15
func (st *SectionStat) PercentMean(percent float32) int64 {
	if percent <= 0 {
		return 0
	}
	if percent >= 1 {
		return st.sectionCounter[len(st.sectionCounter)-1].mean()
	}
	cnt := atomic.LoadInt64(&st.cnt)
	n := int64(float32(cnt) * percent)
	t := int64(0)
	for i := range st.sectionCounter {
		t += st.sectionCounter[i].getCount()
		if t >= n {
			return st.sectionCounter[i].mean()
		}
	}
	return st.sectionCounter[len(st.sectionCounter)-1].mean()
}
