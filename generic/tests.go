package generic

import (
	"errors"
	"reflect"
	"testing"
)

func TestEqual[I any, O any](
	t *testing.T, label string, input I,
	expected, actual O,
) bool {
	t.Helper()

	if reflect.DeepEqual(actual, expected) {
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
