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
	t.Parallel()

	for i := range stringTests {
		test := stringTests[i]
		t.Run("encodeString", func(t *testing.T) {
			t.Parallel()

			var buf bytes.Buffer
			err := encodeString(&buf, test.string)
			if !generic.TestError(t, "encode", test.string, nil, err) {
				return
			}
			actual := buf.Bytes()

			generic.TestEqual(t, "encode", test.string, test.bytes, actual)
		})
	}
}

func TestDecodeString(t *testing.T) {
	t.Parallel()

	for i := range stringTests {
		test := stringTests[i]
		t.Run("decodeString", func(t *testing.T) {
			t.Parallel()

			var actual string
			err := decodeString(bytes.NewReader(test.bytes), &actual)
			if !generic.TestError(t, "decode", test.bytes, nil, err) {
				return
			}

			generic.TestEqual(t, "decode", test.bytes, test.string, actual)
		})
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
			generic.TestError(t, "encode", input, ErrInvalidUtf8String, err)
			return
		} else if !generic.TestError(t, "encode", input, nil, err) {
			return
		}
		encoded := buf.Bytes()

		var decoded string
		err = decodeString(&buf, &decoded)
		if !generic.TestError(t, "decode", encoded, nil, err) {
			return
		}

		generic.TestEqual(t, "roundtrip", input, input, decoded)
	})
}
