package firewalla

import (
	"context"
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
func paginate[T any](ctx context.Context, fetch func(cursor string) (*Page[T], error)) iter.Seq2[T, error] {
	return func(yield func(T, error) bool) {
		var zero T
		cursor := ""
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
			cursor = page.NextCursor
		}
	}
}
