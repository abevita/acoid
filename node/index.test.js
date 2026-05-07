'use strict';

const { describe, it } = require('node:test');
const assert = require('node:assert/strict');
const {
  AcoidError,
  isSupportedLength,
  validateLength,
  validate,
  isValid,
  fromDigest,
  generate,
  generateString,
} = require('./index.js');

// ── Story 3: validateLength / isSupportedLength ───────────────────────────────

describe('isSupportedLength', () => {
  it('returns true for supported lengths', () => {
    for (const l of [6, 8, 10, 12]) {
      assert.equal(isSupportedLength(l), true, `expected ${l} to be supported`);
    }
  });

  it('returns false for unsupported lengths', () => {
    for (const l of [0, 1, 5, 7, 9, 11, 13, 16]) {
      assert.equal(isSupportedLength(l), false, `expected ${l} to be unsupported`);
    }
  });
});

describe('validateLength', () => {
  it('does not throw for supported lengths', () => {
    for (const l of [6, 8, 10, 12]) {
      assert.doesNotThrow(() => validateLength(l));
    }
  });

  it('throws AcoidError for unsupported lengths', () => {
    for (const l of [0, 5, 7, 16]) {
      assert.throws(
        () => validateLength(l),
        (err) => err instanceof AcoidError && err.code === 'UNSUPPORTED_LENGTH',
        `expected UNSUPPORTED_LENGTH for length ${l}`,
      );
    }
  });
});

// ── Story 1: generate ─────────────────────────────────────────────────────────

describe('generate', () => {
  it('is deterministic at all supported lengths', () => {
    const input = Buffer.from('test-input', 'utf8');
    for (const l of [6, 8, 10, 12]) {
      assert.equal(generate(input, l), generate(input, l));
    }
  });

  it('returns the correct length', () => {
    const input = Buffer.from('output-length-test', 'utf8');
    for (const l of [6, 8, 10, 12]) {
      assert.equal(generate(input, l).length, l);
    }
  });

  it('returns only ID57 characters', () => {
    const banned = '0oOIl';
    const input = Buffer.from('charset-check', 'utf8');
    for (const l of [6, 8, 10, 12]) {
      const v = generate(input, l);
      for (const ch of banned) {
        assert.ok(!v.includes(ch), `output ${v} at length ${l} contains banned char '${ch}'`);
      }
    }
  });

  it('produces different outputs for different inputs', () => {
    const a = generate(Buffer.from('alpha', 'utf8'), 8);
    const b = generate(Buffer.from('beta', 'utf8'), 8);
    assert.notEqual(a, b);
  });

  it('throws UNSUPPORTED_LENGTH for length 7', () => {
    assert.throws(
      () => generate(Buffer.from('x'), 7),
      (err) => err instanceof AcoidError && err.code === 'UNSUPPORTED_LENGTH',
    );
  });
});

// ── Story 2: fromDigest ───────────────────────────────────────────────────────

describe('fromDigest', () => {
  const digest = Uint8Array.from({ length: 32 }, (_, i) => i);

  it('is deterministic', () => {
    for (const l of [6, 8, 10, 12]) {
      assert.equal(fromDigest(digest, l), fromDigest(digest, l));
    }
  });

  it('returns the correct length', () => {
    for (const l of [6, 8, 10, 12]) {
      assert.equal(fromDigest(digest, l).length, l);
    }
  });

  it('contains no banned characters', () => {
    const banned = '0oOIl';
    const d = Uint8Array.from({ length: 32 }, (_, i) => (i * 7) & 0xff);
    for (const l of [6, 8, 10, 12]) {
      const v = fromDigest(d, l);
      for (const ch of banned) {
        assert.ok(!v.includes(ch));
      }
    }
  });

  it('throws UNSUPPORTED_LENGTH for length 9', () => {
    assert.throws(
      () => fromDigest(digest, 9),
      (err) => err instanceof AcoidError && err.code === 'UNSUPPORTED_LENGTH',
    );
  });
});

// ── Story 4: validate / isValid ───────────────────────────────────────────────

describe('validate', () => {
  it('accepts round-trip generated values', () => {
    for (const l of [6, 8, 10, 12]) {
      const v = generateString('validate-roundtrip', l);
      assert.doesNotThrow(() => validate(v, l));
    }
  });

  it('throws INVALID_CHARSET for banned chars', () => {
    for (const value of ['A0BCDE', 'AoBCDE', 'AOBCDE', 'AIBCDE', 'AlBCDE']) {
      assert.throws(
        () => validate(value, 6),
        (err) => err instanceof AcoidError && err.code === 'INVALID_CHARSET',
        `should reject ${value}`,
      );
    }
  });

  it('throws LENGTH_MISMATCH when char count differs', () => {
    assert.throws(
      () => validate('ABCDE', 6),
      (err) => err instanceof AcoidError && err.code === 'LENGTH_MISMATCH',
    );
  });

  it('throws UNSUPPORTED_LENGTH for length 7', () => {
    assert.throws(
      () => validate('ABCDEFG', 7),
      (err) => err instanceof AcoidError && err.code === 'UNSUPPORTED_LENGTH',
    );
  });
});

describe('isValid', () => {
  it('returns true for a generated value', () => {
    const v = generateString('isvalid-true', 8);
    assert.equal(isValid(v, 8), true);
  });

  it('returns false for a banned char', () => {
    assert.equal(isValid('AB0CDE', 6), false);
  });

  it('returns false for unsupported length', () => {
    assert.equal(isValid('ABCDEFG', 7), false);
  });
});

// ── Story 5: string / bytes parity ───────────────────────────────────────────

describe('generateString', () => {
  it('matches generate for the same logical input', () => {
    const input = 'string-bytes-parity';
    for (const l of [6, 8, 10, 12]) {
      assert.equal(
        generate(Buffer.from(input, 'utf8'), l),
        generateString(input, l),
      );
    }
  });

  it('is deterministic for empty input', () => {
    const a = generate(Buffer.alloc(0), 8);
    const b = generate(Buffer.alloc(0), 8);
    assert.equal(a, b);
    assert.equal(a.length, 8);
  });

  it('empty string matches empty byte buffer', () => {
    assert.equal(
      generate(Buffer.from('', 'utf8'), 8),
      generateString('', 8),
    );
  });

  it('output is valid ID57', () => {
    const banned = '0oOIl';
    for (const l of [6, 8, 10, 12]) {
      const v = generateString('id57-check', l);
      for (const ch of banned) {
        assert.ok(!v.includes(ch));
      }
    }
  });
});

// ── Cross-language stable vectors ─────────────────────────────────────────────
//
// Established from the Go reference implementation.
// All language implementations MUST produce these exact outputs.

describe('stable vectors', () => {
  const cases = [
    ['acoid:stable:v1', 6,  'MyZx2x'],
    ['acoid:stable:v1', 8,  'MyZx2x9Y'],
    ['acoid:stable:v1', 10, 'MyZx2x9YGy'],
    ['acoid:stable:v1', 12, 'MyZx2x9YGyMr'],
  ];

  for (const [input, length, want] of cases) {
    it(`input="${input}" length=${length} → "${want}"`, () => {
      assert.equal(generateString(input, length), want);
    });
  }
});

// ── E2E round-trip ────────────────────────────────────────────────────────────

describe('E2E round-trip', () => {
  const cases = [
    ['user:abc123', 6],
    ['venue:central-park', 6],
    ['sport:pickleball', 6],
    ['player:jane-doe', 8],
    ['match:2026-05-01:court-3', 10],
    ['activity:drill:forehand', 10],
    ['pair:match:abc:A', 10],
    ['result:match:abc:final', 10],
    ['actpart:activity:xyz:player:p1', 12],
    ['submission:match:abc:round:1', 12],
  ];

  for (const [input, length] of cases) {
    it(`${input} → length ${length}`, () => {
      const id = generateString(input, length);
      assert.equal(id.length, length);
      assert.doesNotThrow(() => validate(id, length));
      assert.equal(id, generateString(input, length)); // deterministic
    });
  }
});
