package iter

import (
	"context"
	"fmt"
)

// Mapper is the signature of a mapper function which is applied to the stream items by the worker go routines.
type Mapper func(ctx context.Context, input interface{}) (outpout interface{}, err error)

// Stream is the signature of a function that creates a new Iterator for processing and streaming of data
// from the given Iterator.
type Stream func(Iterator) Iterator

// NewStream is setting up a Stream func with the given Mapper func and StreamOpts.
func NewStream(ctx context.Context, mapper Mapper, opts ...StreamOpt) Stream {

	return func(inIter Iterator) Iterator {

		// We wrap the given Iterator in a closure with Generator func signature.
		generator := func() (interface{}, error) {
			return inIter.Next()
		}

		// Now we can implement NewStream by calling NewGeneratorStream.
		return NewGeneratorStream(ctx, mapper, opts...)(generator)
	}
}

type streamConf struct {
	BufSize         int
	Workers         int
	ContinueOnError bool
}

// newStreamConf is creating  a default stream config.
func newStreamConf() *streamConf {
	return &streamConf{
		Workers:         1,
		BufSize:         0,
		ContinueOnError: false,
	}
}

// StreamOpt is a functional option type.
type StreamOpt func(conf *streamConf)

// BufSizeOpt is a functional option setting the channel buffer size of the stream (default: 0)
func BufSizeOpt(size int) StreamOpt {
	return func(conf *streamConf) {
		conf.BufSize = size
	}
}

// WorkersOpt is a functional option setting the amount of worker threads (default: 1).
func WorkersOpt(workers int) StreamOpt {
	if workers < 1 {
		panic(fmt.Sprintf("nr of stream Workers: %d - need a least 1 worker", workers))
	}
	return func(conf *streamConf) {
		conf.Workers = workers
	}
}

// ContOnErrOpt is a functional option that lets the stream continue after an error if set to true (default: false).
func ContOnErrOpt(cont bool) StreamOpt {
	return func(conf *streamConf) {
		conf.ContinueOnError = cont
	}
}
