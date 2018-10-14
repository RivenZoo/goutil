package rand

import (
	"math/rand"
	"testing"
	"time"
)

func randInt64Compare() int64 {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return r.Int63()
}

func BenchmarkInt64Generator_Rand(b *testing.B) {
	g := NewInt64Generator(Int64GeneratorOption{
		BufferSize: 16,
	})
	defer g.Close()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			v := g.Rand()
			v += 1
		}
	})
}

func BenchmarkRandInt64Compare(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			v := randInt64Compare()
			v += 1
		}
	})
}
