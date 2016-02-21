package httputil

import log "github.com/cihub/seelog"

import (
	"fmt"
	"strings"
	"sync/atomic"
	"time"
)

type counter struct {
	n         uint64
	startTime time.Time
}

func (c *counter) incr(i uint64) {
	atomic.AddUint64(&c.n, i)
}

func (c *counter) reset() {
	atomic.StoreUint64(&c.n, 0)
	c.startTime = time.Now()
}

func (c *counter) String() string {
	return fmt.Sprintf("{start-time:%s, count:%d}", c.startTime, c.n)
}

type StatStub struct {
	stub map[string]*counter
}

func NewStatStub() *StatStub {
	ss := &StatStub{
		make(map[string]*counter),
	}
	go ss.dailyReset()
	return ss
}

func (ss *StatStub) RegistStat(key string) {
	if _, ok := ss.stub[key]; ok {
		return
	}
	ss.stub[key] = &counter{0, time.Now()}
}

func (ss *StatStub) IncrCounter(key string) {
	if c, ok := ss.stub[key]; ok {
		c.incr(1)
	}
}

func (ss *StatStub) Status() string {
	var s []string
	for k, c := range ss.stub {
		s = append(s, fmt.Sprintf("%s:%s", k, c))
	}
	return strings.Join(s, "\n")
}

func (ss *StatStub) resetStatStub() {
	log.Info(ss.Status())
	for _, c := range ss.stub {
		c.reset()
	}
}

func (ss *StatStub) dailyReset() {
	ticker := time.Tick(1 * time.Minute)
	for now := range ticker {
		if now.Hour() == 23 && now.Minute() == 59 {
			ss.resetStatStub()
		}
	}
}
