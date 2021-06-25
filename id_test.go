package unique

import (
	"fmt"
	"reflect"
	"testing"
	"time"
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
	for i := 0; i < 100000; i++ {
		id := NewID()

		// Smoke-test reasonable entropy by checking for collisions. The chance
		// of collision is less than (1/2)^80, so this test should be reliable.
		if _, ok := seen[id]; ok {
			t.Errorf("ID %s collided but high entropy should make this (nearly) impossible.", id)
		}
		seen[id] = true
	}
}

func TestSetTime(t *testing.T) {
	t.Parallel()

	var id ID
	var zeroTime time.Time
	assertEqualErr(t, "id time out of range", id.SetTime(zeroTime), "Zero time")

	tooSoon := minTime.Add(-time.Millisecond)
	assertEqualErr(t, "id time out of range", id.SetTime(tooSoon), "Time below minimum")

	tooLate := maxTime.Add(time.Millisecond)
	assertEqualErr(t, "id time out of range", id.SetTime(tooLate), "Time above maximum")

	// These also exercise SetTime and MustSetTime.
	assertEqual(t, minTime, NewID().WithTime(minTime).Time(), "Minimum time")
	assertEqual(t, midTime, NewID().WithTime(midTime).Time(), "Arbitrary time")
	assertEqual(t, maxTime, NewID().WithTime(maxTime).Time(), "Maximum time")
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
		if reflect.DeepEqual(id1.Entropy(), id2.Entropy()) {
			t.Error("WithEntropy should not change the receiver.")
		}
		assertEqual(t, tt.Expect, id2.Entropy(), "New ID's entropy should match set value.")

		id1.SetEntropy(tt.Set)
		assertEqual(t, id2, id1, "SetEntropy should change the receiver.")
	}
}

func TestIsZero(t *testing.T) {
	var id ID
	if !id.IsZero() {
		t.Error("Uninitialized ID should be zero.")
	}

	id.MustSetTime(minTime)
	id.SetEntropy(nil)
	if !id.IsZero() {
		t.Error("Explicitly zero ID should be zero.")
	}

	id = NewID()
	id.MustSetTime(minTime)
	if id.IsZero() {
		t.Error("Non-zero entropy should be non-zero ID.")
	}

	id = NewID()
	id.SetEntropy(nil)
	if id.IsZero() {
		t.Error("Non-zero time should be non-zero ID.")
	}
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
		assertEqual(t, tt.Binary, id, "Constructed form for %s (time %v and entropy %v)", tt.Text, tt.Time, tt.Entropy)
		assertEqual(t, tt.Text, id.String(), "String() encoding")

		// Test text encoding.
		text, err := id.MarshalText()
		if assertNoErr(t, err, "MarshalText should succeed") {
			assertEqual(t, tt.Text, string(text), "Text encoding")
		}

		// Test text decoding.
		var parsed ID
		err = parsed.UnmarshalText([]byte(tt.Text))
		if assertNoErr(t, err, "UnmarshalText should succeed") {
			assertEqual(t, tt.Binary, parsed, "Text decoding for %s", tt.Text)
		}

		// Test binary encoding.
		b, err := id.MarshalBinary()
		if assertNoErr(t, err, "MarshalBinary should succeed") {
			assertEqual(t, tt.Binary[:], b, "Binary encoding for %s", tt.Text)
		}

		// Test binary decoding.
		var decoded ID
		err = decoded.UnmarshalBinary(tt.Binary[:])
		if assertNoErr(t, err, "UnmarshalBinary should succeed") {
			assertEqual(t, tt.Binary, decoded, "Binary decoding for %s", tt.Text)
		}
	}
}

func assertNoErr(t *testing.T, err error, message string, args ...interface{}) bool {
	t.Helper()
	if err == nil {
		return true
	}
	t.Error(fmt.Sprintf(message, args...),
		"\n    Error:", err)
	return false
}

func assertEqualErr(
	t *testing.T,
	expect string,
	actual error,
	message string,
	args ...interface{},
) bool {
	t.Helper()
	if actual != nil {
		return assertEqual(t, expect, actual.Error(), message, args...)
	}
	return assertEqual(t, expect, actual, message, args...)
}

func assertEqual(
	t *testing.T,
	expect interface{},
	actual interface{},
	message string,
	args ...interface{},
) bool {
	t.Helper()
	if reflect.DeepEqual(expect, actual) {
		return true
	}
	t.Error(fmt.Sprintf(message, args...),
		"\n    Expected:", expect,
		"\n    Actual:  ", actual)
	return false
}
