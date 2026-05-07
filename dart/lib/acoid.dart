/// ACOID — Deterministic, human-readable identifiers.
///
/// Canonical flow: BLAKE3(input) → base-57 projection → ACOID string
///
/// Supported lengths: 6, 8, 10, 12.
///
/// Projection algorithm:
///   The 32-byte BLAKE3 digest is interpreted as a big-endian integer N.
///   Repeated base-57 division maps each remainder to an ID57 character.
library acoid;

import 'dart:convert';
import 'dart:typed_data';

import 'package:blake3_dart/blake3_dart.dart';

// ── Alphabet ──────────────────────────────────────────────────────────────────

/// ID57 alphabet: Base62 minus {0, o, O, I, l}.
const String _alphabet = 'ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnpqrstuvwxyz123456789';
final BigInt _alphabetSize = BigInt.from(_alphabet.length); // 57
final Set<int> _alphabetCodeUnits = _alphabet.codeUnits.toSet();
const Set<int> _supported = {6, 8, 10, 12};

// ── Error ─────────────────────────────────────────────────────────────────────

/// Codes for [AcoidError].
enum AcoidErrorCode {
  unsupportedLength,
  invalidCharset,
  lengthMismatch,
}

/// Thrown when ACOID validation or generation fails.
class AcoidError implements Exception {
  /// Machine-readable error code.
  final AcoidErrorCode code;

  /// Human-readable message.
  final String message;

  const AcoidError(this.code, this.message);

  @override
  String toString() => 'AcoidError(${code.name}): $message';
}

// ── Validation helpers ────────────────────────────────────────────────────────

/// Returns true if [length] is one of the supported ACOID lengths (6, 8, 10, 12).
bool isSupportedLength(int length) => _supported.contains(length);

/// Throws [AcoidError] if [length] is not 6, 8, 10, or 12.
void validateLength(int length) {
  if (!_supported.contains(length)) {
    throw AcoidError(
      AcoidErrorCode.unsupportedLength,
      'acoid: unsupported length $length; must be 6, 8, 10, or 12',
    );
  }
}

/// Throws [AcoidError] if [value] is not a valid ACOID of the requested [length].
///
/// Checks (in order):
///   1. [length] is a supported value (6, 8, 10, 12)
///   2. [value]'s character count equals [length]
///   3. every character belongs to the ID57 alphabet
void validate(String value, int length) {
  validateLength(length);
  if (value.length != length) {
    throw AcoidError(
      AcoidErrorCode.lengthMismatch,
      'acoid: value length ${value.length} does not match requested length $length',
    );
  }
  for (final cu in value.codeUnits) {
    if (!_alphabetCodeUnits.contains(cu)) {
      throw AcoidError(
        AcoidErrorCode.invalidCharset,
        'acoid: value contains character "${String.fromCharCode(cu)}" outside the ID57 alphabet',
      );
    }
  }
}

/// Returns true if [value] is a valid ACOID of the requested supported [length].
bool isValid(String value, int length) {
  try {
    validate(value, length);
    return true;
  } on AcoidError {
    return false;
  }
}

// ── Core projection ───────────────────────────────────────────────────────────

/// Converts [bytes] to a [BigInt] treating them as a big-endian integer.
/// An empty list yields [BigInt.zero].
BigInt _bytesToBigInt(List<int> bytes) {
  var n = BigInt.zero;
  for (final b in bytes) {
    n = (n << 8) | BigInt.from(b);
  }
  return n;
}

/// Encodes [digest] bytes into a deterministic ACOID of the requested
/// supported [length].
///
/// The digest is interpreted as a big-endian integer. Repeated base-57
/// division maps each remainder to an ID57 character.
///
/// An empty digest is treated as zero and produces a deterministic all-'A'
/// output. Callers that want to hash before encoding should use [generate]
/// or [generateString].
String fromDigest(List<int> digest, int length) {
  validateLength(length);
  var n = _bytesToBigInt(digest);
  final buf = StringBuffer();
  for (var i = 0; i < length; i++) {
    final rem = n.remainder(_alphabetSize).toInt();
    buf.writeCharCode(_alphabet.codeUnitAt(rem));
    n = n ~/ _alphabetSize;
  }
  return buf.toString();
}

/// Hashes [input] bytes with BLAKE3 and returns a deterministic ACOID of
/// the requested supported [length].
///
/// Same input + same length always produces the same ACOID.
String generate(List<int> input, int length) {
  validateLength(length);
  final digest = blake3(Uint8List.fromList(input));
  return fromDigest(digest, length);
}

/// Convenience wrapper around [generate] for string input.
///
/// Encodes [input] as UTF-8 before hashing.
/// Equivalent to `generate(utf8.encode(input), length)`.
String generateString(String input, int length) {
  return generate(utf8.encode(input), length);
}
