import 'dart:typed_data';

import 'package:acoid/acoid.dart';
import 'package:test/test.dart';

void main() {
  // ── Story 3: validateLength / isSupportedLength ──────────────────────────

  group('isSupportedLength', () {
    test('returns true for supported lengths', () {
      for (final l in [6, 8, 10, 12]) {
        expect(isSupportedLength(l), isTrue, reason: 'expected $l to be supported');
      }
    });

    test('returns false for unsupported lengths', () {
      for (final l in [0, 1, 5, 7, 9, 11, 13, 16]) {
        expect(isSupportedLength(l), isFalse, reason: 'expected $l to be unsupported');
      }
    });
  });

  group('validateLength', () {
    test('does not throw for supported lengths', () {
      for (final l in [6, 8, 10, 12]) {
        expect(() => validateLength(l), returnsNormally);
      }
    });

    test('throws AcoidError for unsupported lengths', () {
      for (final l in [0, 5, 7, 16]) {
        expect(
          () => validateLength(l),
          throwsA(
            isA<AcoidError>().having(
              (e) => e.code,
              'code',
              AcoidErrorCode.unsupportedLength,
            ),
          ),
          reason: 'expected unsupportedLength for $l',
        );
      }
    });
  });

  // ── Story 1: generate ────────────────────────────────────────────────────

  group('generate', () {
    test('is deterministic at all supported lengths', () {
      final input = 'test-input'.codeUnits;
      for (final l in [6, 8, 10, 12]) {
        expect(generate(input, l), equals(generate(input, l)));
      }
    });

    test('returns the correct length', () {
      final input = 'output-length-test'.codeUnits;
      for (final l in [6, 8, 10, 12]) {
        expect(generate(input, l).length, equals(l));
      }
    });

    test('returns only ID57 characters', () {
      const banned = '0oOIl';
      final input = 'charset-check'.codeUnits;
      for (final l in [6, 8, 10, 12]) {
        final v = generate(input, l);
        for (final ch in banned.split('')) {
          expect(v.contains(ch), isFalse,
              reason: "output $v at length $l contains banned char '$ch'");
        }
      }
    });

    test('produces different outputs for different inputs', () {
      final a = generate('alpha'.codeUnits, 8);
      final b = generate('beta'.codeUnits, 8);
      expect(a, isNot(equals(b)));
    });

    test('throws AcoidError for unsupported length', () {
      expect(
        () => generate([0x78], 7),
        throwsA(
          isA<AcoidError>().having(
            (e) => e.code,
            'code',
            AcoidErrorCode.unsupportedLength,
          ),
        ),
      );
    });
  });

  // ── Story 2: fromDigest ──────────────────────────────────────────────────

  group('fromDigest', () {
    final digest = List<int>.generate(32, (i) => i);

    test('is deterministic', () {
      for (final l in [6, 8, 10, 12]) {
        expect(fromDigest(digest, l), equals(fromDigest(digest, l)));
      }
    });

    test('returns the correct length', () {
      for (final l in [6, 8, 10, 12]) {
        expect(fromDigest(digest, l).length, equals(l));
      }
    });

    test('contains no banned characters', () {
      const banned = '0oOIl';
      final d = List<int>.generate(32, (i) => (i * 7) & 0xff);
      for (final l in [6, 8, 10, 12]) {
        final v = fromDigest(d, l);
        for (final ch in banned.split('')) {
          expect(v.contains(ch), isFalse);
        }
      }
    });

    test('throws AcoidError for unsupported length', () {
      expect(
        () => fromDigest(digest, 9),
        throwsA(isA<AcoidError>().having(
          (e) => e.code,
          'code',
          AcoidErrorCode.unsupportedLength,
        )),
      );
    });
  });

  // ── Story 4: validate / isValid ──────────────────────────────────────────

  group('validate', () {
    test('accepts round-trip generated values', () {
      for (final l in [6, 8, 10, 12]) {
        final v = generateString('validate-roundtrip', l);
        expect(() => validate(v, l), returnsNormally);
      }
    });

    test('throws INVALID_CHARSET for banned chars', () {
      for (final value in ['A0BCDE', 'AoBCDE', 'AOBCDE', 'AIBCDE', 'AlBCDE']) {
        expect(
          () => validate(value, 6),
          throwsA(isA<AcoidError>().having(
            (e) => e.code,
            'code',
            AcoidErrorCode.invalidCharset,
          )),
          reason: 'should reject $value',
        );
      }
    });

    test('throws LENGTH_MISMATCH when char count differs', () {
      expect(
        () => validate('ABCDE', 6),
        throwsA(isA<AcoidError>().having(
          (e) => e.code,
          'code',
          AcoidErrorCode.lengthMismatch,
        )),
      );
    });

    test('throws UNSUPPORTED_LENGTH for length 7', () {
      expect(
        () => validate('ABCDEFG', 7),
        throwsA(isA<AcoidError>().having(
          (e) => e.code,
          'code',
          AcoidErrorCode.unsupportedLength,
        )),
      );
    });
  });

  group('isValid', () {
    test('returns true for a generated value', () {
      final v = generateString('isvalid-true', 8);
      expect(isValid(v, 8), isTrue);
    });

    test('returns false for a banned char', () {
      expect(isValid('AB0CDE', 6), isFalse);
    });

    test('returns false for unsupported length', () {
      expect(isValid('ABCDEFG', 7), isFalse);
    });
  });

  // ── Story 5: generateString / bytes parity ───────────────────────────────

  group('generateString', () {
    test('matches generate for the same logical input', () {
      const input = 'string-bytes-parity';
      for (final l in [6, 8, 10, 12]) {
        expect(
          generate(input.codeUnits, l),
          equals(generateString(input, l)),
        );
      }
    });

    test('is deterministic for empty input', () {
      final a = generate(const [], 8);
      final b = generate(const [], 8);
      expect(a, equals(b));
      expect(a.length, equals(8));
    });

    test('empty string matches empty byte list', () {
      expect(
        generate(const [], 8),
        equals(generateString('', 8)),
      );
    });

    test('output is valid ID57', () {
      const banned = '0oOIl';
      for (final l in [6, 8, 10, 12]) {
        final v = generateString('id57-check', l);
        for (final ch in banned.split('')) {
          expect(v.contains(ch), isFalse);
        }
      }
    });
  });

  // ── Cross-language stable vectors ────────────────────────────────────────
  //
  // Established from the Go reference implementation.
  // All language implementations MUST produce these exact outputs.

  group('stable vectors', () {
    final cases = [
      ('acoid:stable:v1', 6,  'MyZx2x'),
      ('acoid:stable:v1', 8,  'MyZx2x9Y'),
      ('acoid:stable:v1', 10, 'MyZx2x9YGy'),
      ('acoid:stable:v1', 12, 'MyZx2x9YGyMr'),
    ];

    for (final (input, length, want) in cases) {
      test('input="$input" length=$length → "$want"', () {
        expect(generateString(input, length), equals(want));
      });
    }
  });

  // ── E2E round-trip ───────────────────────────────────────────────────────

  group('E2E round-trip', () {
    final cases = [
      ('user:abc123', 6),
      ('venue:central-park', 6),
      ('sport:pickleball', 6),
      ('player:jane-doe', 8),
      ('match:2026-05-01:court-3', 10),
      ('activity:drill:forehand', 10),
      ('pair:match:abc:A', 10),
      ('result:match:abc:final', 10),
      ('actpart:activity:xyz:player:p1', 12),
      ('submission:match:abc:round:1', 12),
    ];

    for (final (input, length) in cases) {
      test('$input → length $length', () {
        final id = generateString(input, length);
        expect(id.length, equals(length));
        expect(() => validate(id, length), returnsNormally);
        expect(id, equals(generateString(input, length))); // deterministic
      });
    }
  });
}
