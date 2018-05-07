package examples

import (
	"context"

	"github.com/hphilipps/iter"
)

// IntIter is an example for how to construct streaming iterators for specific types using the generic Iterator.
type IntIter struct {
	iter.Iterator
}

// Next is returning an int instead of interface{}.
func (i *IntIter) Next() (int, error) {
	val, err := i.Iterator.Next()
	if err != nil {
		return 0, err
	}
	return val.(int), nil
}

// IntMapper is a mapper func returning an int instead of interface{}.
// We also could define input to be a concrete type instead of interface{} to get
// compile-time type checks. Using interface{} gives us more flexibility but defers
// type checking to the mapper function.
type IntMapper func(ctx context.Context, input interface{}) (output int, err error)

// IntStream is similar to iter.Stream, but returning an IntIter
type IntStream func(iter.Iterator) *IntIter

// NewIntStream is creating a stream of integers.
func NewIntStream(ctx context.Context, intMapper IntMapper, opts ...iter.StreamOpt) IntStream {

	mapper := func(ctx context.Context, input interface{}) (output interface{}, err error) {
		return intMapper(ctx, input)
	}

	stream := iter.NewStream(ctx, mapper, opts...)

	return func(inputIter iter.Iterator) *IntIter {
		return &IntIter{stream(inputIter)}
	}
}
