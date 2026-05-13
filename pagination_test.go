package firewalla

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func TestPaginate_SinglePage(t *testing.T) {
	calls := 0
	fetch := func(cursor string) (*Page[int], error) {
		calls++
		if cursor != "" {
			t.Fatalf("unexpected cursor on first call: %q", cursor)
		}
		return &Page[int]{Count: 3, Results: []int{1, 2, 3}}, nil
	}

	var got []int
	for v, err := range paginate(context.Background(), fetch) {
		if err != nil {
			t.Fatalf("err: %v", err)
		}
		got = append(got, v)
	}
	if calls != 1 {
		t.Errorf("calls = %d", calls)
	}
	if len(got) != 3 || got[0] != 1 || got[2] != 3 {
		t.Errorf("got = %v", got)
	}
}

func TestPaginate_MultiplePages(t *testing.T) {
	pages := []*Page[int]{
		{Results: []int{1, 2}, NextCursor: "p2"},
		{Results: []int{3, 4}, NextCursor: "p3"},
		{Results: []int{5}, NextCursor: ""},
	}
	idx := 0
	fetch := func(cursor string) (*Page[int], error) {
		expected := []string{"", "p2", "p3"}[idx]
		if cursor != expected {
			t.Fatalf("call %d cursor = %q, want %q", idx, cursor, expected)
		}
		p := pages[idx]
		idx++
		return p, nil
	}

	var got []int
	for v, err := range paginate(context.Background(), fetch) {
		if err != nil {
			t.Fatalf("err: %v", err)
		}
		got = append(got, v)
	}
	if len(got) != 5 {
		t.Errorf("got = %v", got)
	}
}

func TestPaginate_ErrorMidStream(t *testing.T) {
	fetchErr := errors.New("boom")
	calls := 0
	fetch := func(cursor string) (*Page[int], error) {
		calls++
		if calls == 1 {
			return &Page[int]{Results: []int{1, 2}, NextCursor: "p2"}, nil
		}
		return nil, fetchErr
	}

	var got []int
	var seenErr error
	for v, err := range paginate(context.Background(), fetch) {
		if err != nil {
			seenErr = err
			break
		}
		got = append(got, v)
	}
	if !errors.Is(seenErr, fetchErr) {
		t.Errorf("err = %v", seenErr)
	}
	if len(got) != 2 {
		t.Errorf("got = %v", got)
	}
}

func TestPaginate_BreakStopsIteration(t *testing.T) {
	calls := 0
	fetch := func(cursor string) (*Page[int], error) {
		calls++
		return &Page[int]{Results: []int{1, 2, 3}, NextCursor: "next"}, nil
	}
	for v, err := range paginate(context.Background(), fetch) {
		if err != nil {
			t.Fatalf("err: %v", err)
		}
		if v == 2 {
			break
		}
	}
	if calls != 1 {
		t.Errorf("calls = %d, want 1 (break should stop fetching)", calls)
	}
}

func TestPaginate_DetectsCursorCycle(t *testing.T) {
	calls := 0
	fetch := func(cursor string) (*Page[int], error) {
		calls++
		// Always return the same cursor "loop" after the first page.
		return &Page[int]{Results: []int{calls}, NextCursor: "loop"}, nil
	}

	var got []int
	var seenErr error
	for v, err := range paginate(context.Background(), fetch) {
		if err != nil {
			seenErr = err
			break
		}
		got = append(got, v)
		// Defensive cap so a real loop bug here can't hang the test forever.
		if calls > 100 {
			t.Fatalf("paginator did not detect cycle after %d calls", calls)
		}
	}
	if seenErr == nil {
		t.Fatal("expected cycle error")
	}
	if !strings.Contains(seenErr.Error(), "cycle") {
		t.Errorf("error %q does not mention cycle", seenErr.Error())
	}
	// We should have made exactly 2 fetches: the first established "loop",
	// the second detected the repeat.
	if calls != 2 {
		t.Errorf("calls = %d, want 2", calls)
	}
	if len(got) != 2 {
		t.Errorf("got = %v, want 2 items before cycle", got)
	}
}

func TestPaginate_ContextCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	fetch := func(cursor string) (*Page[int], error) {
		t.Fatalf("fetch should not be called")
		return nil, nil
	}
	var seenErr error
	for _, err := range paginate(ctx, fetch) {
		seenErr = err
		break
	}
	if !errors.Is(seenErr, context.Canceled) {
		t.Errorf("err = %v", seenErr)
	}
}
