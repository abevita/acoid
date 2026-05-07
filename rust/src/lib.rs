//! ACOID — Deterministic, human-readable identifiers.
//!
//! Canonical flow: `BLAKE3(input) → base-57 projection → ACOID string`
//!
//! Supported lengths: 6, 8, 10, 12.
//!
//! # Projection algorithm
//!
//! The 32-byte BLAKE3 digest is interpreted as a big-endian integer `N`.
//! Repeated base-57 division maps each remainder to an ID57 character.
//! A 32-byte (256-bit) digest encodes up to 43 ID57 characters, covering
//! all supported lengths without additional hashing rounds.
//!
//! # Example
//!
//! ```rust
//! use acoid::{generate_str, validate, is_valid};
//!
//! let id = generate_str("user:123", 6).unwrap();
//! assert_eq!(id.len(), 6);
//! assert!(is_valid(&id, 6));
//! validate(&id, 6).unwrap();
//! ```

use std::fmt;

// ── Alphabet ──────────────────────────────────────────────────────────────────

/// ID57 alphabet: Base62 minus {0, o, O, I, l}.
///
/// Composition:
/// - uppercase: ABCDEFGHJKLMNPQRSTUVWXYZ (24 — removed I, O)
/// - lowercase: abcdefghijkmnpqrstuvwxyz (24 — removed l, o)
/// - digits:    123456789                 (9  — removed 0)
/// - total:     57
const ALPHABET: &[u8; 57] =
    b"ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnpqrstuvwxyz123456789";

const fn build_alphabet_set() -> [bool; 256] {
    let mut set = [false; 256];
    let mut i = 0;
    while i < ALPHABET.len() {
        set[ALPHABET[i] as usize] = true;
        i += 1;
    }
    set
}
const ALPHABET_SET: [bool; 256] = build_alphabet_set();

// ── Error ─────────────────────────────────────────────────────────────────────

/// Errors returned by this crate.
#[derive(Debug, PartialEq, Eq)]
pub enum Error {
    /// Requested length is not 6, 8, 10, or 12.
    UnsupportedLength(usize),
    /// Value contains a character outside the ID57 alphabet.
    InvalidCharset,
    /// Value's character count does not match the requested length.
    LengthMismatch { got: usize, want: usize },
}

impl fmt::Display for Error {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        match self {
            Error::UnsupportedLength(n) => write!(
                f,
                "acoid: unsupported length {n}; must be 6, 8, 10, or 12"
            ),
            Error::InvalidCharset => write!(
                f,
                "acoid: value contains a character outside the ID57 alphabet"
            ),
            Error::LengthMismatch { got, want } => write!(
                f,
                "acoid: value length {got} does not match requested length {want}"
            ),
        }
    }
}

impl std::error::Error for Error {}

pub type Result<T> = std::result::Result<T, Error>;

// ── Validation helpers ────────────────────────────────────────────────────────

/// Returns `true` if `length` is one of the supported ACOID lengths (6, 8, 10, 12).
pub fn is_supported_length(length: usize) -> bool {
    matches!(length, 6 | 8 | 10 | 12)
}

/// Returns `Err(Error::UnsupportedLength)` if `length` is not 6, 8, 10, or 12.
pub fn validate_length(length: usize) -> Result<()> {
    if !is_supported_length(length) {
        return Err(Error::UnsupportedLength(length));
    }
    Ok(())
}

/// Returns an error if `value` is not a valid ACOID of the requested supported length.
///
/// Checks (in order):
/// 1. `length` is a supported value (6, 8, 10, 12)
/// 2. `value`'s character count equals `length`
/// 3. every character belongs to the ID57 alphabet
pub fn validate(value: &str, length: usize) -> Result<()> {
    validate_length(length)?;
    if value.len() != length {
        return Err(Error::LengthMismatch {
            got: value.len(),
            want: length,
        });
    }
    for b in value.bytes() {
        if !ALPHABET_SET[b as usize] {
            return Err(Error::InvalidCharset);
        }
    }
    Ok(())
}

/// Returns `true` if `value` is a valid ACOID of the requested supported length.
pub fn is_valid(value: &str, length: usize) -> bool {
    validate(value, length).is_ok()
}

// ── Core projection ───────────────────────────────────────────────────────────

/// Big-endian long division: divides the number in `n` by `divisor` in place,
/// returns the remainder.
fn div_mod_inplace(n: &mut [u8; 32], divisor: u32) -> u32 {
    let mut rem: u32 = 0;
    for byte in n.iter_mut() {
        let cur = rem * 256 + u32::from(*byte);
        *byte = (cur / divisor) as u8;
        rem = cur % divisor;
    }
    rem
}

/// Encodes `digest` bytes into a deterministic ACOID of the requested supported length.
///
/// The digest is interpreted as a big-endian integer. Repeated base-57 division
/// maps each remainder to an ID57 character.
///
/// A nil or empty digest is treated as zero and produces a deterministic all-`A` output.
/// Callers that want to hash before encoding should use [`generate`] or [`generate_str`].
///
/// Returns [`Error::UnsupportedLength`] if `length` is not 6, 8, 10, or 12.
pub fn from_digest(digest: &[u8], length: usize) -> Result<String> {
    validate_length(length)?;

    let mut n = [0u8; 32];
    let src_len = digest.len().min(32);
    // right-align into the 32-byte buffer (big-endian)
    n[32 - src_len..].copy_from_slice(&digest[..src_len]);

    let mut buf = Vec::with_capacity(length);
    for _ in 0..length {
        let rem = div_mod_inplace(&mut n, 57);
        buf.push(ALPHABET[rem as usize]);
    }

    // Safety: all bytes come from ALPHABET which is valid ASCII
    Ok(String::from_utf8(buf).expect("ALPHABET is valid ASCII"))
}

/// Hashes `input` with BLAKE3 and returns a deterministic ACOID of the requested
/// supported length.
///
/// Same input + same length always produces the same ACOID.
/// Returns [`Error::UnsupportedLength`] if `length` is not 6, 8, 10, or 12.
pub fn generate(input: &[u8], length: usize) -> Result<String> {
    validate_length(length)?;
    let digest = blake3::hash(input);
    from_digest(digest.as_bytes(), length)
}

/// Convenience wrapper around [`generate`] for string input.
///
/// Equivalent to `generate(input.as_bytes(), length)`.
pub fn generate_str(input: &str, length: usize) -> Result<String> {
    generate(input.as_bytes(), length)
}
