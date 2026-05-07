'use strict';

/**
 * ACOID — Deterministic, human-readable identifiers.
 *
 * Canonical flow: BLAKE3(input) → base-57 projection → ACOID string
 *
 * Supported lengths: 6, 8, 10, 12.
 *
 * Projection algorithm:
 *   The 32-byte BLAKE3 digest is interpreted as a big-endian integer N.
 *   Repeated base-57 division maps each remainder to an ID57 character.
 */

const { blake3 } = require('@noble/hashes/blake3');

// ── Alphabet ──────────────────────────────────────────────────────────────────

/** ID57 alphabet: Base62 minus {0, o, O, I, l}. */
const ALPHABET = 'ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnpqrstuvwxyz123456789';
const ALPHABET_SIZE = BigInt(ALPHABET.length); // 57n
const ALPHABET_SET = new Set(ALPHABET);
const SUPPORTED = new Set([6, 8, 10, 12]);

// ── Error ─────────────────────────────────────────────────────────────────────

class AcoidError extends Error {
  /**
   * @param {'UNSUPPORTED_LENGTH'|'INVALID_CHARSET'|'LENGTH_MISMATCH'} code
   * @param {string} message
   */
  constructor(code, message) {
    super(message);
    this.name = 'AcoidError';
    this.code = code;
  }
}

// ── Validation helpers ────────────────────────────────────────────────────────

/**
 * Returns true if length is one of the supported ACOID lengths (6, 8, 10, 12).
 * @param {number} length
 * @returns {boolean}
 */
function isSupportedLength(length) {
  return SUPPORTED.has(length);
}

/**
 * Throws AcoidError if length is not 6, 8, 10, or 12.
 * @param {number} length
 */
function validateLength(length) {
  if (!SUPPORTED.has(length)) {
    throw new AcoidError(
      'UNSUPPORTED_LENGTH',
      `acoid: unsupported length ${length}; must be 6, 8, 10, or 12`,
    );
  }
}

/**
 * Throws AcoidError if value is not a valid ACOID of the requested length.
 *
 * Checks (in order):
 *   1. length is a supported value (6, 8, 10, 12)
 *   2. value's character count equals length
 *   3. every character belongs to the ID57 alphabet
 *
 * @param {string} value
 * @param {number} length
 */
function validate(value, length) {
  validateLength(length);
  if (value.length !== length) {
    throw new AcoidError(
      'LENGTH_MISMATCH',
      `acoid: value length ${value.length} does not match requested length ${length}`,
    );
  }
  for (const ch of value) {
    if (!ALPHABET_SET.has(ch)) {
      throw new AcoidError(
        'INVALID_CHARSET',
        `acoid: value contains character '${ch}' outside the ID57 alphabet`,
      );
    }
  }
}

/**
 * Returns true if value is a valid ACOID of the requested supported length.
 * @param {string} value
 * @param {number} length
 * @returns {boolean}
 */
function isValid(value, length) {
  try {
    validate(value, length);
    return true;
  } catch {
    return false;
  }
}

// ── Core projection ───────────────────────────────────────────────────────────

/**
 * Converts a byte array to a BigInt (big-endian).
 * An empty array yields 0n.
 * @param {Uint8Array} bytes
 * @returns {bigint}
 */
function bytesToBigInt(bytes) {
  let n = 0n;
  for (const b of bytes) {
    n = (n << 8n) | BigInt(b);
  }
  return n;
}

/**
 * Encodes digest bytes into a deterministic ACOID of the requested supported length.
 *
 * The digest is interpreted as a big-endian integer. Repeated base-57 division
 * maps each remainder to an ID57 character.
 *
 * A nil or empty digest is treated as zero and produces a deterministic all-'A' output.
 * Callers that want to hash before encoding should use generate() or generateString().
 *
 * @param {Uint8Array} digest
 * @param {number} length
 * @returns {string}
 */
function fromDigest(digest, length) {
  validateLength(length);
  let n = bytesToBigInt(digest);
  const buf = [];
  for (let i = 0; i < length; i++) {
    buf.push(ALPHABET[Number(n % ALPHABET_SIZE)]);
    n = n / ALPHABET_SIZE;
  }
  return buf.join('');
}

/**
 * Hashes input with BLAKE3 and returns a deterministic ACOID of the requested
 * supported length.
 *
 * Same input + same length always produces the same ACOID.
 *
 * @param {Uint8Array} input
 * @param {number} length
 * @returns {string}
 */
function generate(input, length) {
  validateLength(length);
  const digest = blake3(input);
  return fromDigest(digest, length);
}

/**
 * Convenience wrapper around generate() for string input.
 *
 * Equivalent to generate(Buffer.from(input, 'utf8'), length).
 *
 * @param {string} input
 * @param {number} length
 * @returns {string}
 */
function generateString(input, length) {
  return generate(Buffer.from(input, 'utf8'), length);
}

module.exports = {
  AcoidError,
  isSupportedLength,
  validateLength,
  validate,
  isValid,
  fromDigest,
  generate,
  generateString,
};
