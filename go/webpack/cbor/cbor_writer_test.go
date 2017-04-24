package cbor_test

import (
	"encoding/hex"
	"fmt"
	"strings"
	"testing"

	"github.com/dimich-g/webpackage/go/webpack/cbor"
	"github.com/stretchr/testify/assert"
)

// Converts strings of the form "12 34  5678 9a" to byte slices.
func fromHex(h string) []byte {
	bytes, err := hex.DecodeString(strings.Replace(h, " ", "", -1))
	if err != nil {
		panic(err)
	}
	return bytes
}

func TestIntegers(t *testing.T) {
	assert := assert.New(t)

	var inttests = []struct {
		i        int64
		encoding string
	}{
		{0, "00"},
		{1, "01"},
		{10, "0a"},
		{23, "17"},
		{24, "1818"},
		{25, "1819"},
		{100, "1864"},
		{255, "18ff"},
		{256, "190100"},
		{1000, "1903e8"},
		{1000000, "1a000f4240"},
		{1000000000000, "1b000000e8d4a51000"},
		{-1, "20"},
		{-10, "29"},
		{-100, "3863"},
		{-1000, "3903e7"},
		{-9223372036854775808, "3b7fffffffffffffff"},
	}
	for _, test := range inttests {
		c := cbor.New()
		c.AppendInt64(test.i)
		assert.Equal(fromHex(test.encoding), c.Finish(),
			fmt.Sprintf("%v", test.i))

		if test.i >= 0 {
			c = cbor.New()
			c.AppendUint64(uint64(test.i))
			assert.Equal(fromHex(test.encoding), c.Finish(),
				fmt.Sprintf("Unsigned %v", test.i))
		}
	}

	c := cbor.New()
	c.AppendUint64(18446744073709551615)
	assert.Equal(fromHex("1bffffffffffffffff"), c.Finish())

	c = cbor.New()
	c.AppendFixedSizeUint64(1)
	assert.Equal(fromHex("1b0000000000000001"), c.Finish())
}

func TestString(t *testing.T) {
	assert := assert.New(t)

	c := cbor.New()
	c.AppendBytes([]byte{})
	assert.Equal(fromHex("40"), c.Finish())

	c = cbor.New()
	c.AppendBytes([]byte{0xab})
	assert.Equal(fromHex("41ab"), c.Finish())

	c = cbor.New()
	bytes := make([]byte, 24)
	for i := range bytes {
		bytes[i] = byte(i)
	}
	c.AppendBytes(bytes)
	assert.Equal(fromHex("5818 00 01 02 03 04 05 06 07 08 09 0a 0b 0c 0d 0e 0f 10 11 12 13 14 15 16 17"),
		c.Finish())
	// This doesn't test every size of integer that might encode the length
	// of the byte string.

	var utf8tests = []struct {
		s        string
		encoding string
	}{
		{"", "60"},
		{"a", "6161"},
		{"IETF", "6449455446"},
		{`"\`, "62225c"},
		{"\u00fc", "62c3bc"},
		{"\u6c34", "63e6b0b4"},
		{"\U00010151", "64f0908591"},
	}
	for _, test := range utf8tests {
		c = cbor.New()
		c.AppendUtf8([]byte(test.s))
		assert.Equal(fromHex(test.encoding), c.Finish(), test.s)
	}

	assert.Panics(func() {
		c := cbor.New()
		c.AppendUtf8([]byte{0xff})
	})
}

func TestArrays(t *testing.T) {
	assert := assert.New(t)

	c := cbor.New()
	arr := c.AppendArray(0)
	arr.Finish()
	assert.Equal(fromHex("80"), c.Finish())

	c = cbor.New()
	arr = c.AppendArray(3)
	arr.AppendInt64(1)
	arr.AppendInt64(2)
	arr.AppendInt64(3)
	arr.Finish()
	assert.Equal(fromHex("83 01 02 03"), c.Finish(), "[1, 2, 3]")

	c = cbor.New()
	arr = c.AppendArray(3)
	arr.AppendInt64(1)
	nest1 := arr.AppendArray(2)
	nest1.AppendInt64(2)
	nest1.AppendInt64(3)
	nest1.Finish()
	nest2 := arr.AppendArray(2)
	nest2.AppendInt64(4)
	nest2.AppendInt64(5)
	nest2.Finish()
	arr.Finish()
	assert.Equal(fromHex("83 01 82 02 03 82 04 05"), c.Finish(),
		"[1, [2, 3], [4, 5]]")

	c = cbor.New()
	arr = c.AppendArray(25)
	for i := int64(1); i <= 25; i++ {
		arr.AppendInt64(i)
	}
	arr.Finish()
	assert.Equal(fromHex("9819 01 02 03 04 05 06 07 08 09 0a 0b 0c 0d 0e 0f 10 11 12 13 14 15 16 17 1818 1819"),
		c.Finish(), "[1, ..., 25]")

	assert.Panics(func() {
		c := cbor.New()
		arr := c.AppendArray(2)
		arr.Finish()
	})
}

func TestMaps(t *testing.T) {
	assert := assert.New(t)

	c := cbor.New()
	m := c.AppendMap(0)
	m.Finish()
	assert.Equal(fromHex("a0"), c.Finish(), "{}")

	c = cbor.New()
	m = c.AppendMap(2)
	m.AppendInt64(1)
	m.AppendInt64(2)
	m.AppendInt64(3)
	m.AppendInt64(4)
	m.Finish()
	assert.Equal(fromHex("a2 01 02 03 04"), c.Finish(), "{1: 2, 3: 4}")

	c = cbor.New()
	m = c.AppendMap(2)
	m.AppendUtf8S("a")
	m.AppendInt64(1)
	m.AppendUtf8S("b")
	arr := m.AppendArray(2)
	arr.AppendInt64(2)
	arr.AppendInt64(3)
	arr.Finish()
	m.Finish()
	assert.Equal(fromHex("a2 6161 01 6162 82 02 03"), c.Finish(),
		`{"a": 1, "b": [2, 3]}`)

	c = cbor.New()
	a := c.AppendArray(2)
	a.AppendUtf8S("a")
	m = a.AppendMap(1)
	m.AppendUtf8S("b")
	m.AppendUtf8S("c")
	m.Finish()
	a.Finish()
	assert.Equal(fromHex("82 6161 a1 6162 6163"), c.Finish(),
		`["a", {"b": "c"}]`)

	c = cbor.New()
	m = c.AppendMap(5)
	m.AppendUtf8S("a")
	m.AppendUtf8S("A")
	m.AppendUtf8S("b")
	m.AppendUtf8S("B")
	m.AppendUtf8S("c")
	m.AppendUtf8S("C")
	m.AppendUtf8S("d")
	m.AppendUtf8S("D")
	m.AppendUtf8S("e")
	m.AppendUtf8S("E")
	m.Finish()
	assert.Equal(fromHex("a5 6161 6141 6162 6142 6163 6143 6164 6144 6165 6145"),
		c.Finish(), `{"a": "A", "b": "B", "c": "C", "d": "D", "e": "E"}`)
}

func TestPendingInts(t *testing.T) {
	assert := assert.New(t)

	c := cbor.New()
	arr := c.AppendArray(6)
	p1 := arr.AppendPendingUint()
	p2 := arr.AppendPendingUint()
	p3 := arr.AppendPendingUint()
	p4 := arr.AppendPendingUint()
	p5 := arr.AppendPendingUint()
	arr.AppendUint64(0x0c)
	p1.Complete(1 << 48)
	p2.Complete(1 << 24)
	p3.Complete(1 << 10)
	p4.Complete(0x40)
	p5.Complete(1)
	arr.Finish()
	assert.Equal(fromHex("86 1b0001000000000000 1a01000000 190400 1840 01 0c"),
		c.Finish())
}

func TestByteLenSoFar(t *testing.T) {
	assert := assert.New(t)

	c := cbor.New()
	arr := c.AppendArray(3)
	arr.AppendInt64(1)
	assert.Equal(2, arr.ByteLenSoFar())
	arr.AppendInt64(0x42)
	assert.Equal(4, arr.ByteLenSoFar())

	nested := arr.AppendArray(2)
	assert.Equal(5, arr.ByteLenSoFar())
	assert.Equal(1, nested.ByteLenSoFar())
	p := nested.AppendPendingUint()
	nested.AppendInt64(2)
	assert.Panics(func() { nested.ByteLenSoFar() })
	assert.Panics(func() { arr.ByteLenSoFar() })
	p.Complete(0x43)
	assert.Equal(4, nested.ByteLenSoFar())
	assert.Equal(8, arr.ByteLenSoFar())

	assert.Equal(4, nested.ByteLenSoFar())
	nested.Finish()
	assert.Panics(func() { nested.ByteLenSoFar() })
	assert.Equal(8, arr.ByteLenSoFar())

	arr.Finish()

	assert.Equal(fromHex("83 01 1842 82 1843 02"),
		c.Finish())
}

func TestByteLenSoFarWithPendingInEarlierArray(t *testing.T) {
	assert := assert.New(t)

	c := cbor.New()
	arr := c.AppendArray(1)
	p := arr.AppendPendingUint()
	arr.Finish()

	arr2 := c.AppendArray(2)
	arr2.AppendUint64(0x123)
	lenSoFar := arr2.ByteLenSoFar()
	assert.Equal(4, lenSoFar)
	p.Complete(uint64(lenSoFar))
	arr2.AppendUint64(1)
	arr2.Finish()

	assert.Equal(fromHex("81 04 82 190123 01"),
		c.Finish())
}

func TestByteLenSoFarAfterEarlierPendingShift(t *testing.T) {
	assert := assert.New(t)

	c := cbor.New()
	arr := c.AppendArray(2)
	p1 := arr.AppendPendingUint()
	p2 := arr.AppendPendingUint()
	arr.Finish()

	arr2 := c.AppendArray(0)
	p1.Complete(0x1234)
	assert.Equal(1, arr2.ByteLenSoFar())
	p2.Complete(1)
	arr2.Finish()

	assert.Equal(fromHex("82 191234 01 80"),
		c.Finish())
}
