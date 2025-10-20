package shared

import (
	"reflect"
	"testing"
)

func TestSentenceChunkerSplitSkipsEmpty(t *testing.T) {
	chunker := NewSentenceChunker(50)
	input := "First line.\n\nSecond line.\n"

	got := chunker.Split(input)
	want := []string{"First line.", "Second line."}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Split() mismatch\nwant: %#v\ngot:  %#v", want, got)
	}

	for i, chunk := range got {
		if chunk == "" {
			t.Fatalf("chunk at index %d is empty", i)
		}
	}
}
