package sectionstat_test

import (
	"fmt"
	"math/rand"

	"github.com/RivenZoo/goutil/timestub/sectionstat"
)

func Example_basic() {
	// create section stat by axis
	axis := []int{100, 200, 300, 500, 1000, 2000, 3000, 5000, 10000}
	series := sectionstat.NewSectionSeries(axis)
	stat := sectionstat.NewSectionStat(series)

	times := 1000
	for i := 0; i < times; i++ {
		n := int64(rand.Int31n(int32(100000)))
		// Collect is concurrency safe
		stat.Collect(n)
	}

	fmt.Printf("sum:%d,count:%d,mean:%d\n", stat.Sum(), stat.Count(), stat.Mean())
	fmt.Printf("section mean:%v\n", stat.SectionMean())

	ratio := float32(0.99)
	mean := stat.PercentMean(ratio)
	fmt.Printf("percent 0.99 mean:%d\n", mean)
    // Output: sum:50088051,count:1000,mean:50088
    // section mean:[84 192 0 395 844 1646 2462 4026 7404 54728]
    // percent 0.99 mean:54728
}
