package snowflake

import (
	"time"
)

type IdWorker struct {
	machine       int64
	datacenter    int64
	epoch         int64
	sequence      int64
	lasttimestamp int64
}

func NewIdWorker(machine, datacenter, epoch int64) *IdWorker {
	return &IdWorker{machine & 0x1F, datacenter & 0x1F, epoch, 0, -1}
}

func (this *IdWorker) Generate() (value int64) {
	timeGen := func() int64 {
		return time.Now().UnixNano() / int64(time.Millisecond)
	}

	t := timeGen()
	if t != this.lasttimestamp {
		this.sequence = 0
		goto Generate
	}

	this.sequence = (this.sequence + 1) & 0xFFF
	if this.sequence == 0 {
		for {
			t = timeGen()
			if t > this.lasttimestamp {
				break
			}
		}
	}
Generate:
	this.lasttimestamp = t - this.epoch
	// 时间左移 12+5+5
	value = this.lasttimestamp << 22
	// 数据中心ID左移17位
	value |= this.datacenter << 17
	// 机器码ID左移12位
	value |= this.machine << 12
	// 最后12位
	value |= this.sequence
	return
}
