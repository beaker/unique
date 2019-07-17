package unique

import (
	"crypto/rand"
	"errors"
	"time"

	"github.com/oklog/ulid"
)

var entropyReader = rand.Reader

// ID is 128-bit sortable unique ID.
//
// See specification at https://github.com/ulid/spec
type ID [16]byte

// NewID creates a new unique ID.
func NewID() ID {
	return ID(ulid.MustNew(ulid.Timestamp(time.Now()), entropyReader))
}

// Bytes returns a byte slice representation of an ID.
func (id ID) Bytes() []byte {
	return id[:]
}

// String returns the ID's canonical string form.
func (id ID) String() string {
	return ulid.ULID(id).String()
}

// Time returns the time component of an ID.
func (id ID) Time() time.Time {
	return ulid.Time(ulid.ULID(id).Time()).UTC()
}

// WithTime is a convenience function equivalent to MustSetTime which returns a
// copy of an ID with the given time component.
func (id ID) WithTime(t time.Time) ID {
	id.MustSetTime(t)
	return id
}

// SetTime sets an ID's time component.
//
// If time is known to be between the years [1970, 10889), the error may be
// safely ignored; see MustSetTime.
func (id *ID) SetTime(t time.Time) error {
	if err := ((*ulid.ULID)(id)).SetTime(ulid.Timestamp(t)); err != nil {
		return errors.New("id time out of range")
	}
	return nil
}

// MustSetTime is a convenience function equivalent to SetTime that panics on
// failure instead of returning an error.
func (id *ID) MustSetTime(t time.Time) {
	if err := id.SetTime(t); err != nil {
		panic(err)
	}
}

// Entropy returns the entropy component of an ID.
func (id ID) Entropy() []byte {
	return ulid.ULID(id).Entropy()
}

// WithEntropy is a convenience function equivalent to SetEntropy which returns
// a copy of an ID with the given entropy.
func (id ID) WithEntropy(entropy []byte) ID {
	id.SetEntropy(entropy)
	return id
}

// SetEntropy sets an ID's entropy. Excess entropy is truncated to the first 10
// bytes. Insufficient entropy will be padded to 10 bytes with leading zeroes.
func (id *ID) SetEntropy(entropy []byte) {
	if len(entropy) < 10 {
		buf := [10]byte{}
		copy(buf[10-len(entropy):], entropy)
		entropy = buf[:]
	}
	// This can't fail because we guarantee correct length above.
	_ = ((*ulid.ULID)(id)).SetEntropy(entropy[:10])
}

// IsZero reports whether an ID is all zeroes, such as when it is uninitialized.
func (id ID) IsZero() bool {
	return id == [16]byte{}
}

// MarshalBinary implements the encoding.BinaryMarshaler interface.
func (id ID) MarshalBinary() ([]byte, error) {
	return ulid.ULID(id).MarshalBinary()
}

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface.
func (id *ID) UnmarshalBinary(data []byte) error {
	return ((*ulid.ULID)(id)).UnmarshalBinary(data)
}

// MarshalText implements the encoding.TextMarshaler interface.
func (id ID) MarshalText() ([]byte, error) {
	return ulid.ULID(id).MarshalText()
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
func (id *ID) UnmarshalText(data []byte) error {
	// Avoid ULID's UnmarshalText since it doesn't enforce the character set.
	result, err := ulid.ParseStrict(string(data))
	if err != nil {
		return errors.New("id uses invalid encoding")
	}
	*id = ID(result)
	return nil
}
