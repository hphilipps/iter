package examples

import (
	"context"
	"github.com/hphilipps/iter"
	"testing"
)

type exampleIter struct{}

// We use the Generator func from the generator example to construct an iterator.
func (i exampleIter) Next() (interface{}, error) {
	return dbIterNext()
}

func (i exampleIter) Close() {}

func BenchmarkIteratorExample_1_worker(b *testing.B) {
	b.StopTimer()

	// Instantiate a stream, which is a factory function taking an input Iterator, spinning up
	// the worker goroutines and returning the associated output Iterator.
	// Always close an iterator when done to cancel the associated go routines!
	stream := iter.NewStream(context.Background(), detectPrime)
	iterator := stream(exampleIter{})
	defer iterator.Close()

	b.ReportAllocs()
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		val, err := iterator.Next()
		if err != nil {
			b.Error(err)
		}
		result := val.(data)
		if result.isPrime {
			// ... do something
		}
	}
}

// same as above but with 20 workers - twice as fast on my computer.
func BenchmarkIteratorExample_20_workers(b *testing.B) {
	b.StopTimer()

	// Instantiate a stream with 20 workers this time.
	stream := iter.NewStream(context.Background(), detectPrime, iter.WorkersOpt(20))
	iterator := stream(exampleIter{})
	defer iterator.Close()

	b.ReportAllocs()
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		val, err := iterator.Next()
		if err != nil {
			b.Error(err)
		}
		result := val.(data)
		if result.isPrime {
			// ... do something
		}
	}
}
