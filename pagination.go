package firewalla

import (
	"context"
	"fmt"
	"iter"
)

// Page is one page of results from a paginated list endpoint.
type Page[T any] struct {
	Count      int    `json:"count"`
	Results    []T    `json:"results"`
	NextCursor string `json:"next_cursor"`
}

// paginate returns an iter.Seq2[T, error] that walks all pages by repeatedly
// invoking fetch with the previous page's NextCursor. The iterator yields
// each item once. On error, it yields a zero T with the error and stops.
//
// If the server returns a NextCursor that has been seen before, the iterator
// yields an error rather than looping forever. This protects against server
// bugs or hostile peers from inducing an unbounded loop.
func paginate[T any](ctx context.Context, fetch func(cursor string) (*Page[T], error)) iter.Seq2[T, error] {
	return func(yield func(T, error) bool) {
		var zero T
		cursor := ""
		seen := map[string]struct{}{}
		for {
			if err := ctx.Err(); err != nil {
				yield(zero, err)
				return
			}
			page, err := fetch(cursor)
			if err != nil {
				yield(zero, err)
				return
			}
			for _, item := range page.Results {
				if !yield(item, nil) {
					return
				}
			}
			if page.NextCursor == "" {
				return
			}
			if _, dup := seen[page.NextCursor]; dup {
				yield(zero, fmt.Errorf("firewalla: paginator cursor cycle detected (cursor %q seen twice)", page.NextCursor))
				return
			}
			seen[page.NextCursor] = struct{}{}
			cursor = page.NextCursor
		}
	}
}
