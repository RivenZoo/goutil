/*
Example:
	// create section stat by axis
	axis := []int{100, 200, 300, 500, 1000, 2000, 3000, 5000, 10000}
	series := NewSectionSeries(axis)
	stat := NewSectionStat(series)

	times := 1000
	for i := 0; i < times; i++ {
		n := int64(rand.Int31n(int32(100000)))
		// Collect is concurrency safe
		stat.Collect(n)
	}

	fmt.Println("sum:%d,count:%d,mean:%d", stat.Sum(), stat.Count(), stat.Mean())
	fmt.Println("section mean:%v", stat.SectionMean())

	ratio := float32(0.99)
	mean := stat.PercentMean(ratio)
	fmt.Println("percent 0.99 mean:%d", mean)
*/
package sectionstat
