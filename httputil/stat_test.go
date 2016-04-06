package httputil

import (
	"sync"
	"testing"
)

func TestStatStub(t *testing.T) {
	ss := NewStatStub()
	key, key2 := "test", "url"
	ss.RegistStat(key)
	ss.RegistStat(key2)
	wg := &sync.WaitGroup{}
	for i := 0; i < 5; i++ {
		wg.Add(1)
		ch := make(chan struct{})
		go func() {
			defer wg.Done()
			close(ch)
			for j := 0; j < 5000; j++ {
				if j & 0x01 != 0 {
					ss.IncrCounter(key)
				} else {
					ss.IncrCounter(key2)
				}
			}
		}()
		<- ch
	}
	wg.Wait()
	t.Log(ss.Status())
}
