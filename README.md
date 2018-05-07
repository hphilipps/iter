# Package iter - Streaming Iterators for Go

## Overview

Package `iter` is providing simple interfaces and facilities for streaming data through
iterators in a functional style. This is useful if you want to create an application that
is reading data from a DB, processing it and immediately streaming the results to some client
(without waiting for all data to be processed first).

This package makes it easy to instantiate Streaming Iterators, taking care of all the brittle
details of asynchronous, concurrent streaming while you just need to provide a mapper function
to be applied to the stream.

 
## Features

- provides a simple `Iterator` interface
- takes care of all streaming details behind the scene
- you just need to provide a `Mapper` function for processing of the streamed data
- configurable amount of worker routines
- configurable channel buffer size
- supports *continue on error*
- supports streaming from the 3 most common sources directly:
  - Generators, Iterators and Channels
- Iterators can be chained
- easy to extend to specific types

## Problem Statement

Implementing the iterator pattern in Go is straightforward. And thanks to Channels and Go-routines
it is very easy to stream values asynchronously - just spin up a sender and a receiver Go-routine
and let them communicate over a channel. If you want to also notify about errors
you add another channel. If you want to cancel the sender when the receiver is done things start
 to become complex.
At latest when you start to have multiple senders/receivers running concurrently, maybe using
buffered channels, you will inevitably run into issues and surprising effects, like

- race conditions / deadlocks
- flaky tests / Heisenbugs that are horrible to debug
- unpredictable ordering
- valid results on subsequent calls after an error
- getting an EOF before all go routines are done
- ...

I hit these issues many times and if you want to stream values through several
layers of an application you start to re-implement all this functionality (and the mistakes)
again and again.

## Usage

Package iter is build up on three concepts:

* a `Mapper` func is applied to an input item and creates an output item 
* a `Stream` is a function that can be applied on an input source, sets up go-routines to
apply the `Mapper` func on the input items and returns an `Iterator` for iterating over the results.
* an `Iterator` can be used to iterate over the result items of a stream and to cancel the
background processes when done.

There are three types of directly supported input sources to stream from: *Generator* funcs, *Iterators*
and *Channels*.

### Stream from a Generator Func

```golang
stream := iter.NewGeneratorStream(context.Background(), mapperFunc)
iterator := stream(generatorFunc)
defer iterator.Close()
for {
    item, err := iterator.Next()
    ...
}
 
```

### Stream from an Iterator

```golang
stream := iter.NewStream(context.Background(), mapperFunc)
iterator := stream(inputIter)
defer iterator.Close()
for {
    item, err := iterator.Next()
    ...
}
 
```

### Stream from Channels

```golang
stream := iter.NewChannelStream(context.Background(), mapperFunc)
iterator := stream(inputChan, errChan)
defer iterator.Close()
for {
    item, err := iterator.Next()
    ...
}
 
```

### Stream Options

```golang
// supported options with their defaults:
bufsize := iter.BufSizeOpt(0)
workers := iter.WorkersOpt(1)
contOnErr := iter.ContOnErrOpt(false)

stream := iter.NewStream(context.Background(), mapperFunc, bufSize, workers, contOnErr)
...
```
### Important Properties

 - setting less than 1 worker will cause a panic
 - after `iter.Close()`, `iter.Next()` may still return results with more than 1 worker
 or when using buffers
 - if an `iter.Next()` call returns an error, subsequent calls may still return valid results
 from other workers or a buffered channel
 - choosing more than 1 worker will make the order of results unpredictable
 - by default, a Stream will eventually stop streaming after a Mapper returned an error
   - use `ContOnErrOpt(true)` to change this behavior
 - make sure that generators and mappers are threadsafe if you want to use more than one worker
   - the iterator returned by `New` is threadsafe
 - adding many buffers in a chain of streams will lead to pre-fetching of many items that may
be disregarded when downstream is canceling the stream