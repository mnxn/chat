package protocol

import (
	"bytes"
	"testing"

	"github.com/mnxn/chat/generic"
)

var stringTests = []struct {
	string
	bytes []byte
}{
	{"", []byte{
		0, 0, 0, 0, // int32(0)
	}},
	{"abc", []byte{
		0, 0, 0, 3, // int32(1)
		97, // 'a'
		98, // 'b'
		99, // 'c'
	}},
	{"αβγ", []byte{
		0, 0, 0, 6, // int32(6)
		0xCE, 0xB1, //  "α"
		0xCE, 0xB2, //  "β"
		0xCE, 0xB3, //  "γ"
	}},
}

func TestEncodeString(t *testing.T) {
	for _, test := range stringTests {
		var buf bytes.Buffer
		err := encodeString(&buf, test.string)
		if err != nil {
			generic.TestError(t, "encodeString", test.bytes, err)
			continue
		}

		actual := buf.Bytes()

		if !generic.SliceEqual(actual, test.bytes) {
			generic.TestFailure(t, "encodeString", test.string, test.bytes, actual)
		}
	}
}

func TestDecodeString(t *testing.T) {
	for _, test := range stringTests {
		var actual string
		err := decodeString(bytes.NewReader(test.bytes), &actual)
		if err != nil {
			generic.TestError(t, "decodeString", test.bytes, err)
			continue
		}

		if actual != test.string {
			generic.TestFailure(t, "decodeString", test.bytes, test.string, actual)
		}
	}
}

func FuzzRoundtripString(f *testing.F) {
	seeds := []string{"", "hello123!", "åßçœ®¥"}
	for _, seed := range seeds {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, input string) {
		var buf bytes.Buffer
		err := encodeString(&buf, input)
		if err != nil {
			generic.TestError(t, "encodeString", input, err)
			return
		}
		encoded := buf.Bytes()

		var decoded string
		err = decodeString(&buf, &decoded)
		if err != nil {
			generic.TestError(t, "decodeString", encoded, err)
			return
		}

		if decoded != input {
			generic.TestFailure(t, "roundtrip", input, input, decoded)
		}
	})
}
