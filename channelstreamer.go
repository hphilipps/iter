package iter

import "context"

type ChannelStreamer func(inputChan chan interface{}, errChan chan error) Iterator

func NewChannelStreamer(ctx context.Context, mapper Mapper, opts ...StreamerOpt) ChannelStreamer {

	myCtx, cancel := context.WithCancel(ctx)

	return func(inputChan chan interface{}, errChan chan error) Iterator {

		inIter := &GenericIter{
			itemChan: inputChan,
			errChan:  errChan,
			cancel:   cancel,
		}

		return NewStreamer(myCtx, mapper, opts...)(inIter)
	}
}
