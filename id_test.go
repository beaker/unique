package unique

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var (
	// Lower bound: epoch (0)
	minTime = time.Date(1970, time.January, 1, 0, 0, 0, 0, time.UTC)

	// Arbitrary date near time of writing.
	midTime = time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC)

	// Upper bound: 0xFFFFFFFFFFFF milliseconds since epoch.
	maxTime = time.Date(10889, time.August, 2, 5, 31, 50, 655000000, time.UTC)
)

func TestNewID(t *testing.T) {
	t.Parallel()

	seen := map[ID]bool{}
	for i := 0; i < 1000; i++ {
		id := NewID()
		assert.False(t, id.Time().IsZero())
		assert.NotZero(t, id.Entropy())

		// Smoke-test reasonable entropy by checking for collisions. The chance
		// of collision is less than (1/2)^80, so this test should be reliable.
		assert.NotContains(t, seen, id)
		seen[id] = true
	}
}

func TestSetTime(t *testing.T) {
	t.Parallel()

	var id ID
	var zeroTime time.Time
	assert.EqualError(t, id.SetTime(zeroTime), "id time out of range")
	assert.Panics(t, func() { id.MustSetTime(zeroTime) })

	tooSoon := minTime.Add(-time.Millisecond)
	assert.EqualError(t, id.SetTime(tooSoon), "id time out of range")
	assert.Panics(t, func() { id.MustSetTime(tooSoon) })

	tooLate := maxTime.Add(time.Millisecond)
	assert.EqualError(t, id.SetTime(tooLate), "id time out of range")
	assert.Panics(t, func() { id.MustSetTime(tooLate) })

	// These also exercise SetTime and MustSetTime.
	assert.Equal(t, minTime, NewID().WithTime(minTime).Time())
	assert.Equal(t, midTime, NewID().WithTime(midTime).Time())
	assert.Equal(t, maxTime, NewID().WithTime(maxTime).Time())
}

func TestSetEntropy(t *testing.T) {
	t.Parallel()

	tests := []struct {
		Set    []byte
		Expect []byte
	}{
		{nil, []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0}},
		{[]byte{}, []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0}},
		{[]byte{1, 2, 3}, []byte{0, 0, 0, 0, 0, 0, 0, 1, 2, 3}},
		{[]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}, []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}},
		{[]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13}, []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}},
	}

	for _, tt := range tests {
		id1 := NewID()
		id2 := id1.WithEntropy(tt.Set)
		assert.NotEqual(t, id1.Entropy(), id2.Entropy(), "WithEntropy should not change the source.")
		assert.Equal(t, tt.Expect, id2.Entropy())

		id1.SetEntropy(tt.Set)
		assert.Equal(t, id2, id1)
	}
}

func TestIsZero(t *testing.T) {
	var id ID
	assert.True(t, id.IsZero(), "Uninitialized ID should be zero.")

	id.MustSetTime(minTime)
	id.SetEntropy(nil)
	assert.True(t, id.IsZero(), "Explicitly zero ID should be zero.")

	id = NewID()
	id.MustSetTime(minTime)
	assert.False(t, id.IsZero(), "Non-zero entropy should be non-zero ID.")

	id = NewID()
	id.SetEntropy(nil)
	assert.False(t, id.IsZero(), "Non-zero time should be non-zero ID.")
}

func TestIDEncoding(t *testing.T) {
	t.Parallel()

	tests := []struct {
		Time    time.Time
		Entropy []byte
		Binary  ID
		Text    string
	}{
		{
			Time:    minTime,
			Entropy: nil,
			Binary:  ID{},
			Text:    "00000000000000000000000000",
		},
		{
			Time:    midTime,
			Entropy: []byte{0x01, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef, 0xf0, 0x0d},
			Binary:  ID{0x1, 0x6f, 0x5e, 0x66, 0xe8, 0x0, 0x01, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef, 0xf0, 0x0d},
			Text:    "01DXF6DT0004HMASW9NF6YZW0D",
		},
		{
			Time:    maxTime,
			Entropy: []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
			Binary:  ID{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
			Text:    "7ZZZZZZZZZZZZZZZZZZZZZZZZZ",
		},
	}

	for _, tt := range tests {
		id := ID{}.WithTime(tt.Time).WithEntropy(tt.Entropy)
		assert.Equal(t, tt.Binary, id)

		// Test text encoding.
		assert.Equal(t, tt.Text, id.String())
		text, err := id.MarshalText()
		if assert.NoError(t, err) {
			assert.Equal(t, tt.Text, string(text))
		}

		// Test text decoding.
		var parsed ID
		err = parsed.UnmarshalText([]byte(tt.Text))
		if assert.NoError(t, err) {
			assert.Equal(t, tt.Binary, parsed)
		}

		// Test binary encoding.
		b, err := id.MarshalBinary()
		if assert.NoError(t, err) {
			assert.Equal(t, tt.Binary[:], b)
		}

		// Test binary decoding.
		var decoded ID
		err = decoded.UnmarshalBinary(tt.Binary[:])
		if assert.NoError(t, err) {
			assert.Equal(t, tt.Binary, decoded)
		}
	}
}
