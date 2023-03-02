package generic

import "testing"

func TestFailure[I any, O any](t *testing.T, function string, input I, expected, actual O) {
	t.Errorf("%s %#v:\n  expected %v\n  actual   %v",
		function,
		input,
		expected,
		actual,
	)
}

func TestError[I any](t *testing.T, function string, input I, err error) {
	t.Errorf("%s %#v:\n  error    %s",
		function,
		input,
		err.Error(),
	)
}
