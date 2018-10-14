package rand

import (
	"math/rand"
	"sync"
	"time"
)

const (
	DefaultInt64GeneratorBufferSize = 256
)

type Int64Generator struct {
	option Int64GeneratorOption
	idCh   chan int64
	stop   chan struct{}
	wg     *sync.WaitGroup
}

type Int64GeneratorOption struct {
	BufferSize int `json:"buffer_size"` // pre generate id number
}

func NewInt64Generator(opt Int64GeneratorOption) *Int64Generator {
	if opt.BufferSize <= 0 {
		opt.BufferSize = DefaultInt64GeneratorBufferSize
	}
	g := &Int64Generator{
		option: opt,
		idCh:   make(chan int64, opt.BufferSize),
		stop:   make(chan struct{}),
		wg:     &sync.WaitGroup{},
	}
	go runGenerator(g)
	return g
}

func runGenerator(g *Int64Generator) {
	g.wg.Add(1)
	defer g.wg.Done()

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for {
		v := r.Int63()
		select {
		case <-g.stop:
			return
		case g.idCh <- v:
		}
	}
}

func (g *Int64Generator) Rand() int64 {
	return <-g.idCh
}

func (g *Int64Generator) Close() error {
	close(g.stop)
	g.wg.Wait()
	return nil
}
