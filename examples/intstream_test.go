package examples

import (
	"context"
	"errors"
	"io"
	"sync"
	"testing"
)

var testData = []string{"foo", "bar", "toooooooooooo long", "short"}

type testDataIter struct {
	testData []string
	mu       sync.Mutex
	cursor   int
}

func (i *testDataIter) Next() (interface{}, error) {
	i.mu.Lock()
	defer i.mu.Unlock()
	if i.cursor >= len(i.testData) {
		return "", io.EOF
	}
	out := i.testData[i.cursor]
	i.cursor++
	return out, nil
}

func (i *testDataIter) Close() {
}

func charCountMapper(_ context.Context, input interface{}) (output int, err error) {
	str, ok := input.(string)
	if !ok {
		return 0, errors.New("Input is not of type string")
	}
	return len(str), nil
}

// An example of how to use an IntStream. The mapper function has to check the type of the input data.
// We also could have defined the mapper function to take a concrete type as input instead of interface{}
// to get compile-time checks.
func TestIntStreamer(t *testing.T) {

	inIter := &testDataIter{testData: testData}

	stream := NewIntStream(context.Background(), charCountMapper)

	outIter := stream(inIter)
	defer outIter.Close()

	var i int
	for i = 0; i < len(testData); i++ {
		a, err := outIter.Next()
		if err != nil {
			t.Fatalf("item %s: %v", testData[i], err)
		}

		if want, got := len(testData[i]), a; want != got {
			t.Fatalf("item %s: Expected len %d, got %d", testData[i], want, got)
		}
	}

	if _, err := outIter.Next(); err != io.EOF {
		t.Fatal("Expected io.EOF")
	}
}
