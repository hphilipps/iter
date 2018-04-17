package iter

import (
	"context"
	"io"
	"sync/atomic"
)

const failed int64 = 1

type Iterator interface {
	Next() (interface{}, error)
	Close()
}

type GenericIter struct {
	itemChan chan interface{}
	errChan  chan error
	failed   atomic.Value
	cancel   context.CancelFunc
}

func (i *GenericIter) Next() (interface{}, error) {

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

func (i *GenericIter) Close() {
	i.cancel()
}

func NewGenericIter(outChan chan interface{}, errChan chan error, cancel context.CancelFunc) *GenericIter {
	return &GenericIter{itemChan: outChan, errChan: errChan, cancel: cancel}
}
