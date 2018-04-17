package iter

import (
	"context"
	"fmt"
)

type streamerConf struct {
	BufSize         int
	Workers         int
	ContinueOnError bool
}

func NewStreamerConf() *streamerConf {
	return &streamerConf{Workers: 1, BufSize: 0, ContinueOnError: false}
}

type StreamerOpt func(conf *streamerConf)

func BufSizeOpt(size int) StreamerOpt {
	return func(conf *streamerConf) {
		conf.BufSize = size
	}
}

func WorkersOpt(workers int) StreamerOpt {
	if workers < 1 {
		panic(fmt.Sprintf("nr of streamer Workers: %d - need a least 1 worker", workers))
	}
	return func(conf *streamerConf) {
		conf.Workers = workers
	}
}

func ContOnErrOpt(cont bool) StreamerOpt {
	return func(conf *streamerConf) {
		conf.ContinueOnError = cont
	}
}

type Mapper func(ctx context.Context, input interface{}) (outpout interface{}, err error)

func NopMapper(ctx context.Context, input interface{}) (output interface{}, err error) {
	return input, nil
}

type Streamer func(Iterator) Iterator

func NewStreamer(ctx context.Context, mapper Mapper, opts ...StreamerOpt) Streamer {

	return func(inIter Iterator) Iterator {

		generator := func() (interface{}, error) {
			return inIter.Next()
		}

		return NewGeneratorStreamer(ctx, mapper, opts...)(generator)
	}
}
