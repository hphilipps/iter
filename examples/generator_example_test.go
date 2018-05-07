package examples

import (
	"context"
	"github.com/hphilipps/iter"
	"math/big"
	"math/rand"
	"testing"
)

// A Generator func mocking a simple DB providing random data.
// The func has to match the iter.Generator func type. If that's not
// the case you can easily wrap almost anything into a matching function.
func dbIterNext() (interface{}, error) {
	return rand.Int63(), nil
}

type data struct {
	n       int64
	isPrime bool
}

// This is the mapper func representing the core logic of our app - detecting primes.
func detectPrime(ctx context.Context, input interface{}) (interface{}, error) {
	return data{input.(int64), big.NewInt(input.(int64)).ProbablyPrime(1)}, nil
}

func BenchmarkGeneratorExample_1_worker(b *testing.B) {
	b.StopTimer()

	// Instantiate a stream, which is a factory function taking a generator func, spinning up
	// the worker goroutines and returning the associated Iterator.
	// Always close an iterator when done to cancel the associated go routines!
	stream := iter.NewGeneratorStream(context.Background(), detectPrime)
	iterator := stream(dbIterNext)
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
func BenchmarkGeneratorExample_20_workers(b *testing.B) {
	b.StopTimer()

	// Instantiate a stream with 20 workers this time.
	stream := iter.NewGeneratorStream(context.Background(), detectPrime, iter.WorkersOpt(20))
	iterator := stream(dbIterNext)
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
