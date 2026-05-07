# acoid

Deterministic, human-readable identifiers using BLAKE3 and the ID57 alphabet.

## Overview

ACOID generates short, collision-resistant IDs suitable for any Go project.
The same input and length always produce the same ID — no randomness, no state.

```
BLAKE3(input) → base-57 projection → ACOID string
```

### ID57 alphabet

Base62 minus the four visually ambiguous characters `0 o O I l`:

```
ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnpqrstuvwxyz123456789
```

57 characters total.

### Supported lengths

| Length | Typical use |
|--------|-------------|
| 6      | user, venue, sport, role |
| 8      | player, rating |
| 10     | match, activity, pair, result |
| 12     | activity_participant, submission |

## Installation

```sh
go get github.com/abevita/acoid
```

## Quick start

```go
import "github.com/abevita/acoid"

// Generate from a string seed
id, err := acoid.GenerateString("user:123", 6)

// Generate from raw bytes
id, err := acoid.Generate([]byte("match:abc:1"), 10)

// Validate an existing value
err = acoid.Validate(id, 10)

// Convenience boolean
ok := acoid.IsValid(id, 10)
```

## API

### Generation

```go
// Generate hashes input with BLAKE3 and returns a deterministic ACOID.
func Generate(input []byte, length int) (string, error)

// MustGenerate panics on error. Use with compile-time constant lengths.
func MustGenerate(input []byte, length int) string

// GenerateString is a convenience wrapper for string input.
func GenerateString(input string, length int) (string, error)
```

### Conversion

```go
// FromDigest encodes an existing BLAKE3 digest into an ACOID.
// Use when you already have a digest and only need the projection step.
func FromDigest(digest []byte, length int) (string, error)
```

### Validation

```go
// ValidateLength returns an error if length is not 6, 8, 10, or 12.
func ValidateLength(length int) error

// IsSupportedLength reports whether length is valid.
func IsSupportedLength(length int) bool

// Validate checks supported length, exact char count, and ID57 charset.
func Validate(value string, length int) error

// IsValid reports whether value is a valid ACOID of the requested length.
func IsValid(value string, length int) bool
```

### Sentinel errors

```go
acoid.ErrUnsupportedLength  // length not in {6, 8, 10, 12}
acoid.ErrInvalidCharset     // character outside ID57
acoid.ErrLengthMismatch     // char count ≠ requested length
```

All errors wrap the sentinel, so `errors.Is` works:

```go
_, err := acoid.Generate(input, 7)
if errors.Is(err, acoid.ErrUnsupportedLength) { ... }
```

## Projection algorithm

1. Hash `input` with BLAKE3 → 32-byte digest.
2. Interpret the digest as a big-endian integer `N`.
3. Repeat `length` times: take `N mod 57` → map to alphabet character; `N = N / 57`.
4. Return the resulting string.

A 32-byte digest encodes up to 43 ID57 characters, covering all supported
lengths with no additional hashing rounds.

## Determinism guarantee

> Same input + same length = same ACOID, always.

This holds across calls, goroutines, processes, and Go versions as long as
the package version and BLAKE3 dependency are unchanged. Stable reference
vectors are recorded in `acoid_e2e_test.go`.

## Out of scope

- Storage adapters
- HTTP / API boundary behavior
- Persistence collision handling
- Project-specific entity mappings

## License

MIT
