package main

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"golang.org/x/sync/errgroup"
)

func TestReadPageviews(t *testing.T) {
	tests := []struct{ input, expected string }{
		{"", ""},
		{"only three columns", ""},
		{
			"als.wikipedia Ägypten 4623 mobile-web 2 N1P1\n" +
				"als.wikipedia Ägypten 8911 desktop 3 A2X1\n" +
				"ang.wikipedia Lech_Wałęsa 10374 desktop 1 Q1",
			"als.wiki/ägypten 5|ang.wiki/lech_wałęsa 1",
		},
	}
	for _, c := range tests {
		checkReadPageviews(t, c.input, c.expected)
	}
}

func checkReadPageviews(t *testing.T, input, expected string) {
	ch := make(chan string, 5)
	g, ctx := errgroup.WithContext(context.Background())
	g.Go(func() error {
		defer close(ch)
		return readPageviews(strings.NewReader(input), ch, ctx)
	})
	if err := g.Wait(); err != nil {
		t.Error(err)
		return
	}
	result := make([]string, 0, 5)
	for s := range ch {
		result = append(result, s)
	}
	got := strings.Join(result, "|")
	if expected != got {
		t.Error(fmt.Sprintf("expected %s for %s, got %s", expected, input, got))
		return
	}
}