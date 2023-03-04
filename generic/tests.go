package generic

import (
	"errors"
	"testing"
)

func TestEqual[I any, O comparable](
	t *testing.T, label string, input I,
	expected, actual O,
) bool {
	t.Helper()

	return TestEqualFunc(t, label, input, expected, actual, func(a, e O) bool {
		return a == e
	})
}

func TestEqualFunc[I any, O any](
	t *testing.T, label string, input I,
	expected, actual O, eq func(O, O) bool,
) bool {
	t.Helper()

	if eq(actual, expected) {
		return true
	}

	t.Errorf("%s %#v:\n  expected: %#v\n    actual: %#v",
		label,
		input,
		expected,
		actual,
	)
	return false
}

func TestError[I any](
	t *testing.T, label string, input I,
	expected, actual error,
) bool {
	t.Helper()

	if errors.Is(actual, expected) {
		return true
	}

	t.Errorf("%s %#v:\n  expected error: %v\n    actual error: %v",
		label,
		input,
		expected,
		actual,
	)
	return false
}
