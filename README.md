# acoid

Deterministic, human-readable identifiers using BLAKE3 and the ID57 alphabet.

Available for **Go**, **Rust**, **Node.js**, and **Dart**.

## Overview

ACOID generates short, collision-resistant IDs suitable for any project.
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

## Cross-language stable vectors

All implementations produce these exact outputs for the same input:

| Input              | Length | ACOID          |
|--------------------|--------|----------------|
| `acoid:stable:v1`  | 6      | `MyZx2x`       |
| `acoid:stable:v1`  | 8      | `MyZx2x9Y`     |
| `acoid:stable:v1`  | 10     | `MyZx2x9YGy`   |
| `acoid:stable:v1`  | 12     | `MyZx2x9YGyMr` |

## Language implementations

### Go

**Installation**

```sh
go get github.com/abevita/acoid/go
```

**Quick start**

```go
import "github.com/abevita/acoid/go"

id, err := acoid.GenerateString("user:123", 6)
id, err  = acoid.Generate([]byte("match:abc:1"), 10)
err       = acoid.Validate(id, 10)
ok       := acoid.IsValid(id, 10)
```

### Rust

**Installation** — add to `Cargo.toml`:

```toml
[dependencies]
acoid = { path = "path/to/acoid/rust" }
```

**Quick start**

```rust
use acoid::{generate_str, validate, is_valid};

let id = generate_str("user:123", 6)?;
validate(&id, 6)?;
let ok = is_valid(&id, 6);
```

### Node.js

**Installation**

```sh
npm install acoid
```

Or locally:

```sh
npm install path/to/acoid/node
```

**Quick start**

```js
const { generateString, validate, isValid } = require('acoid');

const id = generateString('user:123', 6);
validate(id, 6);                // throws AcoidError on failure
const ok = isValid(id, 6);
```

### Dart

**Installation** — add to `pubspec.yaml`:

```yaml
dependencies:
  acoid:
    path: path/to/acoid/dart
```

**Quick start**

```dart
import 'package:acoid/acoid.dart';

final id = generateString('user:123', 6);
validate(id, 6);           // throws AcoidError on failure
final ok = isValid(id, 6);
```

## API (all languages follow this contract)

| Operation          | Description |
|--------------------|-------------|
| `generate`         | Hash bytes with BLAKE3 → ACOID |
| `generateString`   | Hash a UTF-8 string with BLAKE3 → ACOID |
| `fromDigest`       | Encode an existing BLAKE3 digest → ACOID |
| `validate`         | Check length + charset; throw/return error on failure |
| `isValid`          | Boolean version of `validate` |
| `validateLength`   | Throw/return error if length ∉ {6, 8, 10, 12} |
| `isSupportedLength`| Boolean version of `validateLength` |

### Error conditions (all languages)

| Code / Sentinel         | When |
|-------------------------|------|
| `UnsupportedLength`     | length not in {6, 8, 10, 12} |
| `InvalidCharset`        | character outside ID57 |
| `LengthMismatch`        | char count ≠ requested length |

## Go API reference

### Generation

```go
func Generate(input []byte, length int) (string, error)
func MustGenerate(input []byte, length int) string
func GenerateString(input string, length int) (string, error)
```

### Conversion

```go
// FromDigest encodes an existing BLAKE3 digest into an ACOID.
func FromDigest(digest []byte, length int) (string, error)
```

### Validation

```go
func ValidateLength(length int) error
func IsSupportedLength(length int) bool
func Validate(value string, length int) error
func IsValid(value string, length int) bool
```

### Sentinel errors

```go
acoid.ErrUnsupportedLength
acoid.ErrInvalidCharset
acoid.ErrLengthMismatch
```

All errors wrap the sentinel, so `errors.Is` works:

```go
import "github.com/abevita/acoid/go"

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

This holds across calls, threads, processes, languages, and runtimes as long as
the BLAKE3 implementation is standards-compliant and the package version is
unchanged. Stable reference vectors are verified in the test suite for each
language.

## Out of scope

- Storage adapters
- HTTP / API boundary behavior
- Persistence collision handling
- Project-specific entity mappings

## License

MIT
