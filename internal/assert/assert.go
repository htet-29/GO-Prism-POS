// Package assert provides helper test functions
package assert

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func Equal[T comparable](t *testing.T, expected, actual T) {
	t.Helper()

	if diff := cmp.Diff(expected, actual); diff != "" {
		t.Logf("expectd\n%v\n", expected)
		t.Logf("got\n%v\n", actual)
		t.Errorf("parseContent() mismatch (-expected +actual):\n%s", diff)
	}
}
