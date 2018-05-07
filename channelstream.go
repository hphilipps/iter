package iter

import "context"

// ChannelStream is a function that creates a new Iterator for processing and streaming of data
// from the given inputChan, while listening for errors on errChan.
type ChannelStream func(inputChan chan interface{}, errChan chan error) Iterator

// NewChannelStream is setting up a ChannelStream func with the given Mapper func and StreamOpts.
func NewChannelStream(ctx context.Context, mapper Mapper, opts ...StreamOpt) ChannelStream {

	myCtx, cancel := context.WithCancel(ctx)

	return func(inputChan chan interface{}, errChan chan error) Iterator {

		inIter := New(inputChan, errChan, cancel)

		return NewStream(myCtx, mapper, opts...)(inIter)
	}
}
