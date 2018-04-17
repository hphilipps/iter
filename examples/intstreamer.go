package examples

import (
	"context"

	"github.com/hphilipps/iter"
)

type IntIter struct {
	iter.Iterator
}

func (i *IntIter) Next() (int, error) {
	val, err := i.Iterator.Next()
	if err != nil {
		return 0, err
	}
	return val.(int), nil
}

type IntMapper func(ctx context.Context, input interface{}) (output int, err error)

type IntStreamer func(iter.Iterator) *IntIter

func NewIntStreamer(ctx context.Context, intMapper IntMapper, opts ...iter.StreamerOpt) IntStreamer {

	mapper := func(ctx context.Context, input interface{}) (output interface{}, err error) {
		return intMapper(ctx, input)
	}

	streamer := iter.NewStreamer(ctx, mapper, opts...)

	return func(inputIter iter.Iterator) *IntIter {
		return &IntIter{streamer(inputIter)}
	}
}
