package snowflake

import "sync"

const (
	DefaultIDBufferSize = 256
)

type IDGenerator struct {
	option IDGeneratorOption
	idCh   chan int64
	stop   chan struct{}
	wg     *sync.WaitGroup
}

type IDGeneratorOption struct {
	BufferSize int   `json:"buffer_size"` // pre generate id number
	Machine    int64 `json:"machine"`
	Datacenter int64 `json:"datacenter"`
	Epoch      int64 `json:"epoch"`
}

func NewIDGenerator(opt IDGeneratorOption) *IDGenerator {
	if opt.BufferSize <= 0 {
		opt.BufferSize = DefaultIDBufferSize
	}
	g := &IDGenerator{
		option: opt,
		idCh:   make(chan int64, opt.BufferSize),
		stop:   make(chan struct{}),
		wg:     &sync.WaitGroup{},
	}
	go runGenerator(g)
	return g
}

func runGenerator(g *IDGenerator) {
	g.wg.Add(1)
	defer g.wg.Done()

	worker := NewIdWorker(g.option.Machine, g.option.Datacenter, g.option.Epoch)
	for {
		v := worker.Generate()
		select {
		case <-g.stop:
			return
		case g.idCh <- v:
		}
	}
}

func (g *IDGenerator) NextID() int64 {
	return <-g.idCh
}

func (g *IDGenerator) Close() error {
	close(g.stop)
	g.wg.Wait()
	return nil
}
