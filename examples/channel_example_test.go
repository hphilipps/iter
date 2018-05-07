package examples

import (
	"context"
	"github.com/hphilipps/iter"
	"math/rand"
	"testing"
)

// A func starting a go routine filling the channel with random data similar
// to the generator example. The go routine is listening on the Done channel of the given
// context for a signal to cancel it self.
func fillChannels(ctx context.Context) (chan interface{}, chan error) {
	itemChan := make(chan interface{})
	errChan := make(chan error)

	go func() {
		defer close(itemChan)
		for {
			select {
			case <-ctx.Done():
				return
			case itemChan <- rand.Int63():
			}
		}
	}()

	return itemChan, errChan
}

func BenchmarkChannelExample_1_worker(b *testing.B) {
	b.StopTimer()

	// needed to cancel the fillChannels go routine
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Instantiate a stream, which is a factory function taking an item channel
	// and an error channel, spinning up the worker routines and returning the associated Iterator.
	// Always close an iterator when done to cancel the associated go routines!
	stream := iter.NewChannelStream(context.Background(), detectPrime)
	iterator := stream(fillChannels(ctx))
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
func BenchmarkChannelExample_20_workers(b *testing.B) {
	b.StopTimer()

	// needed to cancel the fillChannels go routine
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Instantiate a stream with 20 workers this time.
	stream := iter.NewChannelStream(context.Background(), detectPrime, iter.WorkersOpt(20))
	iterator := stream(fillChannels(ctx))
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
