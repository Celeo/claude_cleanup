package main

import (
	"strings"
	"testing"
)

func TestListFooterHint(t *testing.T) {
	cases := []struct {
		name string
		mode mode
		want string // substring that pins the user-facing meaning
	}{
		{"delete mode advertises view-trash and tags itself", modeDelete, "view trash"},
		{"delete mode includes the [DELETE] tag", modeDelete, "[DELETE]"},
		{"trash mode advertises back-to-active and tags itself", modeTrash, "back to active"},
		{"trash mode includes the [TRASH] tag", modeTrash, "[TRASH]"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := rootModel{mode: tc.mode}.listFooterHint()
			if !strings.Contains(got, tc.want) {
				t.Errorf("hint for mode=%v = %q, want substring %q", tc.mode, got, tc.want)
			}
		})
	}
}
