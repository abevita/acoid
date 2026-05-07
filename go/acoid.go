// Package acoid generates and validates ACOID values.
//
// An ACOID is a deterministic, human-readable identifier built from the ID57
// alphabet — Base62 minus the visually ambiguous characters 0, o, O, I, l.
//
// Canonical flow:
//
//	BLAKE3(input) → base-57 projection → truncate to length → ACOID string
//
// Supported lengths: 6, 8, 10, 12.
//
// Projection algorithm:
//
// The 32-byte BLAKE3 digest is interpreted as a big-endian integer N.
// Repeated division by 57 maps each remainder to an ID57 character.
// A 32-byte digest encodes up to 43 ID57 characters, comfortably covering
// all supported lengths. BLAKE3's XOF mode can extend this if needed.
package acoid

import (
	"errors"
	"fmt"
	"math/big"

	"lukechampine.com/blake3"
)

// alphabet is the ID57 character set: Base62 minus {0, o, O, I, l}.
//
// Composition:
//
//	uppercase: ABCDEFGHJKLMNPQRSTUVWXYZ  (24 — removed I, O)
//	lowercase: abcdefghijkmnpqrstuvwxyz  (24 — removed l, o)
//	digits:    123456789                  (9  — removed 0)
//	total:     57
const alphabet = "ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnpqrstuvwxyz123456789"

var (
	alphabetSize = big.NewInt(int64(len(alphabet))) // 57

	// alphabetSet enables O(1) charset membership checks.
	alphabetSet [256]bool
)

func init() {
	for i := 0; i < len(alphabet); i++ {
		alphabetSet[alphabet[i]] = true
	}
}

// Sentinel errors returned by this package.
var (
	// ErrUnsupportedLength is returned when a length outside {6, 8, 10, 12} is requested.
	ErrUnsupportedLength = errors.New("acoid: unsupported length; must be 6, 8, 10, or 12")

	// ErrInvalidCharset is returned when a value contains a character outside the ID57 alphabet.
	ErrInvalidCharset = errors.New("acoid: value contains a character outside the ID57 alphabet")

	// ErrLengthMismatch is returned when a value's character count does not match the requested length.
	ErrLengthMismatch = errors.New("acoid: value length does not match the requested length")
)

// IsSupportedLength reports whether length is one of the supported ACOID lengths (6, 8, 10, 12).
func IsSupportedLength(length int) bool {
	switch length {
	case 6, 8, 10, 12:
		return true
	}
	return false
}

// ValidateLength returns ErrUnsupportedLength if length is not 6, 8, 10, or 12.
func ValidateLength(length int) error {
	if !IsSupportedLength(length) {
		return fmt.Errorf("%w: got %d", ErrUnsupportedLength, length)
	}
	return nil
}

// FromDigest encodes digest bytes into a deterministic ACOID of the requested
// supported length.
//
// The digest is interpreted as a big-endian integer. Repeated base-57 division
// maps each remainder to an ID57 character. The caller is responsible for
// supplying a digest produced by a deterministic hash function (e.g. BLAKE3).
//
// A nil or empty digest is treated as the integer zero and produces a
// deterministic (all-'A') output. Callers that want to hash before encoding
// should use [Generate] or [GenerateString] instead.
//
// Returns ErrUnsupportedLength if length is not 6, 8, 10, or 12.
func FromDigest(digest []byte, length int) (string, error) {
	if err := ValidateLength(length); err != nil {
		return "", err
	}

	n := new(big.Int).SetBytes(digest)
	mod := new(big.Int)
	buf := make([]byte, length)

	for i := 0; i < length; i++ {
		n.DivMod(n, alphabetSize, mod)
		buf[i] = alphabet[mod.Int64()]
	}

	return string(buf), nil
}

// Generate hashes input with BLAKE3 and returns a deterministic ACOID of the
// requested supported length.
//
// Same input + same length always produces the same ACOID.
// Returns ErrUnsupportedLength if length is not 6, 8, 10, or 12.
func Generate(input []byte, length int) (string, error) {
	if err := ValidateLength(length); err != nil {
		return "", err
	}
	digest := blake3.Sum256(input)
	return FromDigest(digest[:], length)
}

// MustGenerate is like Generate but panics on error.
// Use only when length is a compile-time constant you control.
func MustGenerate(input []byte, length int) string {
	v, err := Generate(input, length)
	if err != nil {
		panic(err)
	}
	return v
}

// GenerateString is a convenience wrapper around Generate for string input.
//
// Equivalent to Generate([]byte(input), length).
func GenerateString(input string, length int) (string, error) {
	return Generate([]byte(input), length)
}

// Validate returns an error if value is not a valid ACOID of the requested
// supported length. Checks (in order):
//
//  1. length is a supported value (6, 8, 10, 12)
//  2. value's character count equals length
//  3. every character belongs to the ID57 alphabet
//
// Returns ErrUnsupportedLength, ErrLengthMismatch, or ErrInvalidCharset.
func Validate(value string, length int) error {
	if err := ValidateLength(length); err != nil {
		return err
	}
	if len(value) != length {
		return fmt.Errorf("%w: got %d, want %d", ErrLengthMismatch, len(value), length)
	}
	for i := 0; i < len(value); i++ {
		if !alphabetSet[value[i]] {
			return fmt.Errorf("%w: character %q at index %d", ErrInvalidCharset, value[i], i)
		}
	}
	return nil
}

// IsValid reports whether value is a valid ACOID of the requested supported length.
func IsValid(value string, length int) bool {
	return Validate(value, length) == nil
}
