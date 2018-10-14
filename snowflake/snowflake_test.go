package snowflake

import (
	"sync"
	"testing"
)

func TestIDGenerator_NextID(t *testing.T) {
	g := NewIDGenerator(IDGeneratorOption{
		BufferSize: 16,
	})
	defer g.Close()

	t.Parallel()
	wg := &sync.WaitGroup{}
	for i := 0; i < 5; i++ {
		ch := make(chan struct{})
		go func() {
			close(ch)
			wg.Add(1)
			for j := 0; j < 5; j++ {
				t.Log(g.NextID())
			}
			wg.Done()
		}()
		<-ch
	}
	wg.Wait()
}

func BenchmarkIDGenerator_NextID(b *testing.B) {
	g := NewIDGenerator(IDGeneratorOption{
		BufferSize: 16,
	})
	defer g.Close()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			id := g.NextID()
			id += 1
		}
	})
}
