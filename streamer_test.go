package iter

import (
	"context"
	"errors"
	"io"
	"sync"
	"testing"
	"time"
)

var errFive = errors.New("I don't like 5")

var testCases = []struct {
	bufSize int
	workers int
}{
	{bufSize: 0, workers: 1},
	{bufSize: 0, workers: 2},
	{bufSize: 0, workers: 3},
	{bufSize: 1, workers: 1},
	{bufSize: 1, workers: 2},
	{bufSize: 1, workers: 3},
	{bufSize: 2, workers: 1},
	{bufSize: 2, workers: 2},
	{bufSize: 2, workers: 3},
	{bufSize: 0, workers: 20},
	{bufSize: 10, workers: 20},
}

type data struct {
	input  int
	result int
}

var list = []data{{input: 1}, {input: 2}, {input: 3}, {input: 4}, {input: 5}, {input: 6}, {input: 7}, {input: 8}, {input: 9}}

type testIter struct {
	list   []data
	mu     sync.Mutex
	cursor int
}

func (i *testIter) Next() (interface{}, error) {
	i.mu.Lock()
	defer i.mu.Unlock()
	if i.cursor >= len(i.list) {
		return nil, io.EOF
	}
	out := i.list[i.cursor]
	i.cursor++
	return out, nil
}

func (i *testIter) Close() {
}

func squareMapper(ctx context.Context, input interface{}) (output interface{}, err error) {
	in := input.(data)
	in.result = in.input * in.input
	return in, nil
}

func failingMapper(ctx context.Context, input interface{}) (output interface{}, err error) {
	in := input.(data)
	if in.input == 5 {
		return nil, errFive
	}
	in.result = in.input * in.input
	return in, nil
}

func TestStreamer(t *testing.T) {

	for testnr, parms := range testCases {

		inIter := &testIter{list: list}

		streamer := NewStreamer(context.Background(), squareMapper, BufSizeOpt(parms.bufSize), WorkersOpt(parms.workers))

		outIter := streamer(inIter)
		defer outIter.Close()

		i := 0
		for i = 0; i < len(list); i++ {
			a, err := outIter.Next()
			if err != nil {
				t.Fatalf("test %d, item %d: %v", testnr, i, err)
			}

			if want, got := a.(data).input*a.(data).input, a.(data).result; want != got {
				t.Fatalf("test %d: Expected %d^2 = %d, got %d", testnr, a.(data).input, want, got)
			}
		}

		if a, err := outIter.Next(); err != io.EOF {
			t.Fatalf("test %d: Expected io.EOF: %v %v", testnr, a, err)
		}
	}
}

func TestStreamerReadAfterClose(t *testing.T) {

	for testnr, parms := range testCases {

		inIter := &testIter{list: list}

		streamer := NewStreamer(context.Background(), squareMapper, BufSizeOpt(parms.bufSize), WorkersOpt(parms.workers))

		outIter := streamer(inIter)

		_, err := outIter.Next()
		if err != nil {
			t.Fatal(err)
		}

		outIter.Close()

		// wait for cancel signal to be received by go routines
		time.Sleep(time.Millisecond)

		_, err = outIter.Next()

		// only without buffering we can be sure to not receive results after a close
		if parms.bufSize == 0 {
			if err != io.EOF {
				t.Fatalf("test %d: Expected io.EOF", testnr)
			}
		}
	}
}

func TestStreamerInjectFailure(t *testing.T) {

	for testnr, parms := range testCases {

		inIter := &testIter{list: list}

		streamer := NewStreamer(context.Background(), failingMapper, BufSizeOpt(parms.bufSize), WorkersOpt(parms.workers))

		outIter := streamer(inIter)
		defer outIter.Close()

		for i := 0; i < len(list); i++ {
			_, err := outIter.Next()

			if err != nil {
				if err == errFive {
					break
				}
				t.Fatalf("test %d: Expected worker to fail for input 5 but got err %v", testnr, err)
			}
		}

		// only with one worker and no buffers we can be sure to not receive results after a failed Next()
		if parms.workers == 1 && parms.bufSize == 0 {
			if a, err := outIter.Next(); err != io.EOF {
				t.Fatalf("test %d: Expected io.EOF: %v, %v", testnr, a, err)
			}
		}
	}
}

func TestChannelStreamer(t *testing.T) {

	for testnr, parms := range testCases {

		inputChan := make(chan interface{}, parms.bufSize)
		errChan := make(chan error)

		go func() {
			for _, i := range list {
				inputChan <- i
			}
			close(inputChan)
		}()

		streamer := NewChannelStreamer(context.Background(), squareMapper, BufSizeOpt(parms.bufSize), WorkersOpt(parms.workers))

		outIter := streamer(inputChan, errChan)

		for i := 0; i < len(list); i++ {
			a, err := outIter.Next()
			if err != nil {
				t.Fatal(err)
			}

			if want, got := a.(data).input*a.(data).input, a.(data).result; want != got {
				t.Fatalf("test %d: Expected %d^2 = %d, got %d", testnr, a.(data).input, want, got)
			}
		}

		if _, err := outIter.Next(); err != io.EOF {
			t.Fatalf("test %d: Expected io.EOF", testnr)
		}
	}
}

func TestChannelStreamerReadAfterClose(t *testing.T) {

	for testnr, parms := range testCases {

		inputChan := make(chan interface{}, parms.bufSize)
		errChan := make(chan error)

		go func() {
			for _, i := range list {
				inputChan <- i
			}
			close(inputChan)
		}()

		streamer := NewChannelStreamer(context.Background(), squareMapper, BufSizeOpt(parms.bufSize), WorkersOpt(parms.workers))

		outIter := streamer(inputChan, errChan)

		_, err := outIter.Next()
		if err != nil {
			t.Fatalf("test %d: %v", testnr, err)
		}

		outIter.Close()

		// wait for cancel signal to be received
		time.Sleep(time.Millisecond)

		_, err = outIter.Next()
		if err != io.EOF {
			t.Fatalf("test %d: Expected io.EOF", testnr)
		}
	}
}

func TestChannelStreamerInjectFailure(t *testing.T) {

	for testnr, parms := range testCases {

		inputChan := make(chan interface{}, parms.bufSize)
		errChan := make(chan error)

		go func() {
			for _, i := range list {
				inputChan <- i
			}
			close(inputChan)
		}()

		streamer := NewChannelStreamer(context.Background(), failingMapper, BufSizeOpt(parms.bufSize), WorkersOpt(parms.workers))

		outIter := streamer(inputChan, errChan)
		defer outIter.Close()

		for i := 0; i < len(list); i++ {
			_, err := outIter.Next()

			if err != nil {
				if err == errFive {
					break
				}
				t.Fatalf("test %d: Expected worker to fail for input 5 but got err %v", testnr, err)
			}
		}

		// only with one worker we can be sure to not receive results after a failed Next()
		if parms.workers == 1 {
			if _, err := outIter.Next(); err != io.EOF {
				t.Fatalf("test %d: Expected io.EOF", testnr)
			}
		}
	}
}

func TestGeneratorStreamer(t *testing.T) {

	for testnr, parms := range testCases {
		mu := &sync.Mutex{}
		i := 0

		generator := func() (interface{}, error) {
			mu.Lock()
			defer mu.Unlock()
			defer func() { i++ }()

			if i >= len(list) {
				return nil, io.EOF
			}
			if i == 4 {
				return nil, errFive
			}
			return list[i], nil
		}

		streamer := NewGeneratorStreamer(context.Background(), NopMapper, BufSizeOpt(parms.bufSize), WorkersOpt(parms.workers))
		iter := streamer(generator)
		defer iter.Close()

		j := 0
		for ; j < len(list); j++ {
			_, err := iter.Next()
			if err != nil {
				if err == errFive {
					break
				}
				t.Fatalf("test %d: Expected worker to fail for input 5 but got err %v", testnr, err)
			}
		}

		// only with one worker and without buffer we can be sure to receive results in the right order
		if parms.workers == 1 && parms.bufSize == 0 {
			if want, got := 4, j; want != got {
				t.Fatalf("Expected %d processed items, got %d", want, got)
			}
		}

		// only with one worker and without buffer we can be sure to not receive results after a failed Next()
		if parms.workers == 1 && parms.bufSize == 0 {
			if _, err := iter.Next(); err != io.EOF {
				t.Fatalf("test %d: Expected io.EOF: %v", testnr, err)
			}
		}
	}
}

func TestChaining(t *testing.T) {
	for testnr, parms := range testCases {

		mu := sync.Mutex{}
		i := 0

		generator := func() (interface{}, error) {
			mu.Lock()
			defer mu.Unlock()

			if i >= len(list) {
				return nil, io.EOF
			}

			i++
			return list[i-1], nil
		}

		ctx := context.Background()
		iter := NewStreamer(ctx, failingMapper)(NewGeneratorStreamer(ctx, NopMapper)(generator))
		defer iter.Close()

		j := 0
		for ; j < len(list); j++ {
			_, err := iter.Next()
			if err != nil {
				if err == errFive {
					break
				}
				t.Fatalf("test %d: Expected worker to fail for input 5 but got err %v", testnr, err)
			}
		}

		// only with one worker and no buffers we can be sure to not receive results after a failed Next()
		if parms.workers == 1 && parms.bufSize == 0 {
			if want, got := 4, j; want != got {
				t.Fatalf("Expected %d processed items, got %d", want, got)
			}
		}

		// only with one worker and no buffers we can be sure to not receive results after a failed Next()
		if parms.workers == 1 && parms.bufSize == 0 {
			if _, err := iter.Next(); err != io.EOF {
				t.Fatalf("test %d: Expected io.EOF", testnr)
			}
		}
	}
}

func TestContOnError(t *testing.T) {

	for testnr, parms := range testCases {

		mu := &sync.Mutex{}
		i := 0

		generator := func() (interface{}, error) {
			mu.Lock()
			defer mu.Unlock()
			defer func() { i++ }()

			if i >= len(list) {
				return nil, io.EOF
			}
			if i == 4 {
				return nil, errFive
			}
			return list[i], nil
		}

		ctx := context.Background()

		workers := WorkersOpt(parms.workers)
		bufSize := BufSizeOpt(parms.bufSize)
		contOnErr := ContOnErrOpt(true)

		genStreamer := NewGeneratorStreamer(ctx, NopMapper, workers, bufSize, contOnErr)
		streamer := NewStreamer(ctx, failingMapper, workers, bufSize, contOnErr)

		iter := streamer(genStreamer(generator))
		defer iter.Close()

		j := 0
		for ; j < len(list); j++ {
			a, err := iter.Next()
			if err != nil {
				if err == errFive {
					continue
				}
				t.Fatalf("test %d: Expected worker to fail for input 5 but got err %v", testnr, err)
			}

			if want, got := a.(data).input*a.(data).input, a.(data).result; want != got {
				t.Fatalf("test %d: Expected %d^2 = %d, got %d", testnr, a.(data).input, want, got)
			}
		}

		if want, got := 9, j; want != got {
			t.Fatalf("Expected %d processed items, got %d", want, got)
		}

		if _, err := iter.Next(); err != io.EOF {
			t.Fatalf("test %d - %d: Expected io.EOF: %v", testnr, j, err)
		}
	}
}
