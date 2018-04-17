package examples

import (
	"context"
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

func charCountWorker(ctx context.Context, input interface{}) (output int, err error) {
	str := input.(string)
	return len(str), nil
}

func TestIntStreamer(t *testing.T) {

	inIter := &testDataIter{testData: testData}

	streamer := NewIntStreamer(context.Background(), charCountWorker)

	outIter := streamer(inIter)
	defer outIter.Close()

	i := 0
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
