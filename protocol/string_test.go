package protocol

import (
	"bytes"
	"testing"
	"unicode/utf8"

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
		if !generic.TestError(t, "encodeString", test.string, nil, err) {
			continue
		}

		actual := buf.Bytes()

		if !generic.TestEqualFunc(t, "encodeString", test.string, test.bytes, actual, generic.SliceEqual[byte]) {
			continue
		}
	}
}

func TestDecodeString(t *testing.T) {
	for _, test := range stringTests {
		var actual string
		err := decodeString(bytes.NewReader(test.bytes), &actual)
		if !generic.TestError(t, "decodeString", test.bytes, nil, err) {
			continue
		}

		if !generic.TestEqual(t, "decodeString", test.bytes, test.string, actual) {
			continue
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
		if !utf8.ValidString(input) {
			generic.TestError(t, "encodeString", input, ErrInvalidUtf8String, err)
			return
		} else if !generic.TestError(t, "encodeString", input, nil, err) {
			return
		}
		encoded := buf.Bytes()

		var decoded string
		err = decodeString(&buf, &decoded)
		if !generic.TestError(t, "decodeString", encoded, nil, err) {
			return
		}

		if !generic.TestEqual(t, "roundtrip", input, input, decoded) {
			return
		}
	})
}
