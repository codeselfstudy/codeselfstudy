package ingest

import (
	"testing"
	"time"
)

func TestResolveBudgetScalesAndCaps(t *testing.T) {
	cases := []struct {
		deals int
		want  time.Duration
	}{
		{0, 15 * time.Second},
		{1, 20 * time.Second},
		{5, 40 * time.Second},
		{6, 45 * time.Second},  // exactly at the ceiling
		{50, 45 * time.Second}, // capped
	}
	for _, tc := range cases {
		if got := resolveBudget(tc.deals); got != tc.want {
			t.Errorf("resolveBudget(%d) = %v, want %v", tc.deals, got, tc.want)
		}
	}
}
