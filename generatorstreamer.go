package iter

import (
	"context"
	"io"

	"golang.org/x/sync/errgroup"
)

type Generator func() (interface{}, error)

type GeneratorStreamer func(Generator) Iterator

func NewGeneratorStreamer(ctx context.Context, mapper Mapper, opts ...StreamerOpt) GeneratorStreamer {

	cfg := NewStreamerConf()
	for _, opt := range opts {
		opt(cfg)
	}

	itemChan := make(chan interface{}, cfg.BufSize)
	errChan := make(chan error)

	myCtx, cancel := context.WithCancel(ctx)

	iter := NewGenericIter(itemChan, errChan, cancel)

	return func(next Generator) Iterator {

		go func() {
			defer close(itemChan)

			// errgroup for Workers - all Workers will be canceled after the first error
			eg, egCtx := errgroup.WithContext(myCtx)

			for i := 0; i < cfg.Workers; i++ {

				eg.Go(func() error {

					for {
						var res interface{}

						item, err := next()

						if err == nil {
							res, err = mapper(myCtx, item)
						}

						if err != nil {
							if err == io.EOF {
								return nil
							}
							select {
							case errChan <- err:
								if cfg.ContinueOnError {
									continue
								}
								return err
							case <-egCtx.Done():
								return nil
							}
						}

						select {
						case itemChan <- res:
						case <-egCtx.Done():
							return nil
						}
					}
				})
			}

			// wait for all Workers to finish or cancel the remaining ones after the first error
			eg.Wait()
		}()

		return iter
	}
}
