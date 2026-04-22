package tui

import "testing"

func TestScrollWindow(t *testing.T) {
	tests := []struct {
		name                 string
		cursor, offset, size int
		total                int
		wantOffset           int
	}{
		{"cursor_above_offset", 0, 5, 10, 50, 0},
		{"cursor_below_window", 15, 0, 10, 50, 6},
		{"cursor_inside", 3, 0, 10, 50, 0},
		{"cursor_at_last", 49, 40, 10, 50, 40},
		{"tiny_total", 2, 0, 10, 3, 0},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := scrollOffset(tc.cursor, tc.offset, tc.size, tc.total)
			if got != tc.wantOffset {
				t.Errorf("scrollOffset(cursor=%d, offset=%d, size=%d, total=%d) = %d, want %d",
					tc.cursor, tc.offset, tc.size, tc.total, got, tc.wantOffset)
			}
		})
	}
}
