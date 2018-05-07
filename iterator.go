package iter

import (
	"context"
	"io"
)

// Iterator is the interface for an object that can be used to iterate through a set of items.
type Iterator interface {
	Next() (interface{}, error)
	Close()
}

// iterator is implementing the Iterator interface.
// You need to call Close() to cancel the go routines of the related stream.
type iterator struct {
	itemChan chan interface{}
	errChan  chan error
	cancel   context.CancelFunc
}

// New is returning a new *iterator instance.
func New(itemChan chan interface{}, errChan chan error, cancel context.CancelFunc) Iterator {
	return &iterator{itemChan: itemChan, errChan: errChan, cancel: cancel}
}

// Next is returning the next item from the stream or an io.EOF error when the stream is closed.
// After an io.EOF error, subsequent calls still might return valid items if using buffered channels
// or multiple worker goroutines (which not all may have been canceled yet).
// When the stream was configured with continue-on-error, the goroutines will not be canceled after
// any other error and try to stream further items.
// When using multiple stream workers the result order is unpredictable.
func (i *iterator) Next() (interface{}, error) {

	select {
	case item, ok := <-i.itemChan:
		if !ok { // channel was closed by sender
			return nil, io.EOF
		}
		return item, nil

	case err := <-i.errChan:
		return nil, err
	}
}

// Close is sending a cancel signal to all goroutines of the stream.
func (i *iterator) Close() {
	i.cancel()
}
