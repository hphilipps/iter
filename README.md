# iter
Go iterator tools

 - setting less than 1 worker will cause a panic
 - after `iter.Close()`, `iter.Next()` may still return buffered results
 - if an `iter.Next()` call returns an error, subsequent calls may still return valid results
 from other workers or a buffered channel
 - choosing more than 1 worker will make the order of results unpredictable!
 - by default, a Streamer will eventually stop streaming after a Mapper returned an error
   - use `ContOnErrOpt(true)` to change this behavior
 - make sure that generators and mappers are threadsafe if you want to use more than one worker!
   - GenericIterator is threadsafe
 - add buffer only once (below the slowest layer)